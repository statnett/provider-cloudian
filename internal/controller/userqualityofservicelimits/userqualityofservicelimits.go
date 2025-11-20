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

package userqualityofservicelimits

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/statnett/provider-cloudian/apis/user/v1alpha1"
	apisv1alpha1 "github.com/statnett/provider-cloudian/apis/v1alpha1"
	qoslimits "github.com/statnett/provider-cloudian/internal/controller/qualityofservicelimits"
	"github.com/statnett/provider-cloudian/internal/sdk/cloudian"
)

const (
	errNotUserQualityOfServiceLimits = "managed resource is not a UserQualityOfServiceLimits custom resource"
	errTrackPCUsage                  = "cannot track ProviderConfig usage"
	errGetPC                         = "cannot get ProviderConfig"
	errGetCreds                      = "cannot get credentials"

	errNewClient = "cannot create new Service"
	errCreateQOS = "cannot create QOS"
	errDeleteQOS = "cannot delete QOS"
	errGetQOS    = "cannot get QOS"
)

var (
	newCloudianService = func(providerConfig *apisv1alpha1.ProviderConfig, authHeader string) (*cloudian.Client, error) {
		return cloudian.NewClient(
			providerConfig.Spec.Endpoint,
			authHeader,
		), nil
	}
)

// Setup adds a controller that reconciles UserQualityOfServiceLimits managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.UserQualityOfServiceLimitsGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.UserQualityOfServiceLimitsGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewLegacyProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: newCloudianService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.UserQualityOfServiceLimits{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        *resource.LegacyProviderConfigUsageTracker
	newServiceFn func(providerConfig *apisv1alpha1.ProviderConfig, authHeader string) (*cloudian.Client, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.UserQualityOfServiceLimits)
	if !ok {
		return nil, errors.New(errNotUserQualityOfServiceLimits)
	}

	if err := c.usage.Track(ctx, cr); err != nil {
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
	cr, ok := mg.(*v1alpha1.UserQualityOfServiceLimits)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotUserQualityOfServiceLimits)
	}

	groupID := cr.Spec.ForProvider.GroupID
	if groupID == "" {
		return managed.ExternalObservation{}, nil
	}
	userID := cr.Spec.ForProvider.UserID
	if userID == "" {
		return managed.ExternalObservation{}, nil
	}

	guid := cloudian.GroupUserID{
		GroupID: groupID,
		UserID:  userID,
	}
	qos, err := c.cloudianService.GetQOS(ctx, guid, cr.Spec.ForProvider.Region)

	if errors.Is(err, cloudian.ErrNotFound) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetQOS)
	}

	cr.SetConditions(xpv1.Available())

	expected, err := qoslimits.ToCloudianQOS(cr.Spec.ForProvider.QOS)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: expected.Warning.Equal(qos.Warning) &&
			expected.Hard.Equal(qos.Hard),

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.UserQualityOfServiceLimits)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotUserQualityOfServiceLimits)
	}

	qos, err := qoslimits.ToCloudianQOS(cr.Spec.ForProvider.QOS)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	guid := cloudian.GroupUserID{
		GroupID: cr.Spec.ForProvider.GroupID,
		UserID:  cr.Spec.ForProvider.UserID,
	}
	if err := c.cloudianService.SetQOS(ctx, guid, cr.Spec.ForProvider.Region, qos); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateQOS)
	}

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.UserQualityOfServiceLimits)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotUserQualityOfServiceLimits)
	}

	qos, err := qoslimits.ToCloudianQOS(cr.Spec.ForProvider.QOS)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	guid := cloudian.GroupUserID{
		GroupID: cr.Spec.ForProvider.GroupID,
		UserID:  cr.Spec.ForProvider.UserID,
	}
	if err := c.cloudianService.SetQOS(ctx, guid, cr.Spec.ForProvider.Region, qos); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateQOS)
	}

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.UserQualityOfServiceLimits)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotUserQualityOfServiceLimits)
	}

	cr.SetConditions(xpv1.Deleting())

	guid := cloudian.GroupUserID{
		GroupID: cr.Spec.ForProvider.GroupID,
		UserID:  cr.Spec.ForProvider.UserID,
	}
	err := c.cloudianService.DeleteQOS(ctx, guid, cr.Spec.ForProvider.Region)
	if err != nil && !errors.Is(err, cloudian.ErrNotFound) {
		return managed.ExternalDelete{}, errors.Wrap(err, errGetCreds)
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) Disconnect(ctx context.Context) error {
	return nil
}
