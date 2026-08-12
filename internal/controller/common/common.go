package common

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1cluster "github.com/statnett/provider-cloudian/apis/cluster/v1alpha1"
	pcv1alpha1common "github.com/statnett/provider-cloudian/apis/common/providerconfig/v1alpha1"
	apisv1alpha1namespaced "github.com/statnett/provider-cloudian/apis/namespaced/v1alpha1"
	"github.com/statnett/provider-cloudian/internal/sdk/cloudian"
)

//nolint:gocyclo
func GetClient(ctx context.Context, c client.Client, mg resource.Managed) (*cloudian.Client, error) {
	switch mgC := mg.(type) {
	//nolint:staticcheck // SA1019
	case resource.LegacyManaged:
		pcRef := mgC.GetProviderConfigReference()
		if pcRef == nil {
			return nil, errors.New("providerConfigRef is not given")
		}

		pc := &apisv1alpha1cluster.ProviderConfig{}
		if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name}, pc); err != nil {
			return nil, errors.Wrap(err, "cannot get referenced ProviderConfig")
		}

		t := resource.NewLegacyProviderConfigUsageTracker(c, &apisv1alpha1cluster.ProviderConfigUsage{})
		if err := t.Track(ctx, mgC); err != nil {
			return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
		}

		return buildConfigFromSpec(ctx, c, nil, pc.Spec)
	case resource.ModernManaged:
		pcRef := mgC.GetProviderConfigReference()
		if pcRef == nil {
			return nil, errors.New("providerConfigRef is not given")
		}

		if pcRef.Kind == "ClusterProviderConfig" {
			cpc := &apisv1alpha1namespaced.ClusterProviderConfig{}
			if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name}, cpc); err != nil {
				return nil, errors.Wrap(err, "cannot get referenced ClusterProviderConfig")
			}

			t := resource.NewProviderConfigUsageTracker(c, &apisv1alpha1namespaced.ProviderConfigUsage{})
			if err := t.Track(ctx, mgC); err != nil {
				return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
			}

			return buildConfigFromSpec(ctx, c, nil, cpc.Spec)
		} else {
			pc := &apisv1alpha1namespaced.ProviderConfig{}
			if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name, Namespace: mg.GetNamespace()}, pc); err != nil {
				return nil, errors.Wrap(err, "cannot get referenced ProviderConfig")
			}

			t := resource.NewProviderConfigUsageTracker(c, &apisv1alpha1namespaced.ClusterProviderConfigUsage{})
			if err := t.Track(ctx, mgC); err != nil {
				return nil, errors.Wrap(err, "cannot track ClusterProviderConfig usage")
			}

			return buildConfigFromSpec(ctx, c, new(mg.GetNamespace()), pc.Spec)
		}
	default:
		return nil, errors.New("unknown managed resource type")
	}
}

func buildConfigFromSpec(ctx context.Context, c client.Client, ns *string, spec pcv1alpha1common.ProviderConfigSpec) (*cloudian.Client, error) {
	cd := spec.AuthHeader

	// Don't allow cross namespace secret reference with ns scoped ProviderConfig
	if ns != nil && cd.Source == xpv2.CredentialsSourceSecret && cd.SecretRef != nil {
		cd.SecretRef.Namespace = *ns
	}

	authHeader, err := resource.CommonCredentialExtractor(ctx, cd.Source, c, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get credentials")
	}

	svc, err := NewCloudianService(spec.Endpoint, string(authHeader))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Service")
	}

	return svc, nil
}

func NewCloudianService(providerConfigEndpoint string, authHeader string) (*cloudian.Client, error) {
	return cloudian.NewClient(
		providerConfigEndpoint,
		authHeader,
	), nil
}
