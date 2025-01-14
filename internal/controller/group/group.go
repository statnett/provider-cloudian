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
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/statnett/provider-cloudian/apis/user/v1alpha1"
	apisv1alpha1 "github.com/statnett/provider-cloudian/apis/v1alpha1"
	"github.com/statnett/provider-cloudian/internal/features"
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

var (
	newCloudianService = func(providerConfig *apisv1alpha1.ProviderConfig, authHeader string) (*cloudian.Client, error) {
		// FIXME: Don't require InsecureSkipVerify
		return cloudian.NewClient(
			providerConfig.Spec.Endpoint,
			authHeader,
			cloudian.WithInsecureTLSVerify(true),
		), nil
	}
)

// Setup adds a controller that reconciles Group managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.GroupGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.GroupGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newCloudianService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Group{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(providerConfig *apisv1alpha1.ProviderConfig, authHeader string) (*cloudian.Client, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return nil, errors.New(errNotGroup)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.AuthHeader
	authHeader, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(pc, string(authHeader))
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
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGroup)
	}

	observedGroup, err := c.cloudianService.GetGroup(ctx, meta.GetExternalName(mg))
	if err != nil {
		if errors.Is(err, cloudian.ErrNotFound) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetGroup)
	}

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: isUpToDate(meta.GetExternalName(mg), cr.Spec.ForProvider, *observedGroup),

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGroup)
	}

	cr.SetConditions(xpv1.Creating())

	if err := c.cloudianService.CreateGroup(ctx, newCloudianGroup(meta.GetExternalName(mg), cr.Spec.ForProvider)); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateGroup)
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGroup)
	}

	if err := c.cloudianService.UpdateGroup(ctx, newCloudianGroup(meta.GetExternalName(mg), cr.Spec.ForProvider)); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateGroup)
	}

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Group)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotGroup)
	}

	cr.SetConditions(xpv1.Deleting())

	if err := c.cloudianService.DeleteGroup(ctx, meta.GetExternalName(mg)); err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteGroup)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}

func isUpToDate(name string, desired v1alpha1.GroupParameters, observed cloudian.Group) bool {
	return newCloudianGroup(name, desired) == observed
}

func newCloudianGroup(name string, gp v1alpha1.GroupParameters) cloudian.Group {
	return cloudian.Group{
		Active:             gp.Active,
		GroupID:            name,
		GroupName:          gp.GroupName,
		LDAPEnabled:        ptr.Deref(gp.LDAPEnabled, false),
		LDAPGroup:          ptr.Deref(gp.LDAPGroup, ""),
		LDAPMatchAttribute: ptr.Deref(gp.LDAPMatchAttribute, ""),
		LDAPSearch:         ptr.Deref(gp.LDAPSearch, ""),
		LDAPSearchUserBase: ptr.Deref(gp.LDAPSearchUserBase, ""),
		LDAPServerURL:      ptr.Deref(gp.LDAPServerURL, ""),
		LDAPUserDNTemplate: ptr.Deref(gp.LDAPUserDNTemplate, ""),
	}
}
