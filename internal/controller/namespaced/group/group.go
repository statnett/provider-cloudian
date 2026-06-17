/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package group

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"

	userv1alpha1namespaced "github.com/statnett/provider-cloudian/apis/namespaced/user/v1alpha1"
	apisv1alpha1namespaced "github.com/statnett/provider-cloudian/apis/namespaced/v1alpha1"
	controllercommon "github.com/statnett/provider-cloudian/internal/controller/common"
	groupcontrollercommon "github.com/statnett/provider-cloudian/internal/controller/common/group"
	"github.com/statnett/provider-cloudian/internal/sdk/cloudian"
)

const (
	errNotGroup     = "managed resource is not a Group custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient   = "cannot create new Service"
	errCreateGroup = "cannot create Group"
	errDeleteGroup = "cannot delete Group"
	errGetGroup    = "cannot get Group"
	errUpdateGroup = "cannot update Group"
)

// SetupGated registers controller setup with the gate, waiting for the
// required CRD
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	o.Gate.Register(func() {
		if err := Setup(mgr, o); err != nil {
			panic(err)
		}
	}, userv1alpha1namespaced.GroupGroupVersionKind)
	return nil
}

// Setup adds a controller that reconciles Group managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(userv1alpha1namespaced.GroupGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(userv1alpha1namespaced.GroupGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1namespaced.ProviderConfigUsage{}),
			newServiceFn: controllercommon.NewCloudianService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		//nolint:staticcheck // SA1004 crossplane-runtime still depends on deprecated API
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&userv1alpha1namespaced.Group{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        *resource.ProviderConfigUsageTracker
	newServiceFn func(providerConfigEndpoint string, authHeader string) (*cloudian.Client, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*userv1alpha1namespaced.Group)
	if !ok {
		return nil, errors.New(errNotGroup)
	}

	if err := c.usage.Track(ctx, cr); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1namespaced.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.AuthHeader
	authHeader, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(pc.Spec.Endpoint, string(authHeader))
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{cloudianService: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	cloudianService *cloudian.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*userv1alpha1namespaced.Group)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroup)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{}, nil
	}

	observedGroup, err := c.cloudianService.GetGroup(ctx, externalName)
	if errors.Is(err, cloudian.ErrNotFound) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetGroup)
	}

	cr.SetConditions(xpv2.Available())

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: groupcontrollercommon.IsUpToDate(meta.GetExternalName(mg), cr.Spec.ForProvider, *observedGroup),

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*userv1alpha1namespaced.Group)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroup)
	}

	cr.SetConditions(xpv2.Creating())

	if err := c.cloudianService.CreateGroup(ctx, groupcontrollercommon.NewCloudianGroup(meta.GetExternalName(mg), cr.Spec.ForProvider)); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateGroup)
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*userv1alpha1namespaced.Group)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGroup)
	}

	if err := c.cloudianService.UpdateGroup(ctx, groupcontrollercommon.NewCloudianGroup(meta.GetExternalName(mg), cr.Spec.ForProvider)); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateGroup)
	}

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*userv1alpha1namespaced.Group)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotGroup)
	}

	cr.SetConditions(xpv2.Deleting())

	if err := c.cloudianService.DeleteGroup(ctx, meta.GetExternalName(mg)); err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteGroup)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}
