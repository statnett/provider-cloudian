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

func GetClient(ctx context.Context, c client.Client, mg resource.Managed) (*cloudian.Client, error) {
	switch mgC := mg.(type) {
	//nolint:staticcheck // SA1019
	case resource.LegacyManaged:
		return nil, errors.New("legacy managed resource not supported")
	case resource.ModernManaged:
		pcRef := mgC.GetProviderConfigReference()
		if pcRef == nil {
			return nil, errors.New("providerConfigRef is not given")
		}

		switch pcRef.Kind {
		case "ClusterProviderConfig":
			cpc := &apisv1alpha1cluster.ProviderConfig{}
			if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name}, cpc); err != nil {
				return nil, errors.Wrap(err, "cannot get referenced ClusterProviderConfig")
			}
			return buildConfigFromSpec(ctx, c, mgC, cpc.Spec)
		default:
			pc := &apisv1alpha1namespaced.ProviderConfig{}
			if err := c.Get(ctx, types.NamespacedName{Name: pcRef.Name, Namespace: mg.GetNamespace()}, pc); err != nil {
				return nil, errors.Wrap(err, "cannot get referenced ProviderConfig")
			}
			return buildConfigFromSpec(ctx, c, mgC, pc.Spec)
		}
	default:
		return nil, errors.New("unknown managed resource type")
	}
}

func buildConfigFromSpec(ctx context.Context, c client.Client, m resource.ModernManaged, spec pcv1alpha1common.ProviderConfigSpec) (*cloudian.Client, error) {
	cd := spec.AuthHeader

	switch cd.Source { // Don't allow cross namespace secret reference
	case xpv2.CredentialsSourceSecret:
		if cd.SecretRef != nil {
			cd.SecretRef.Namespace = m.GetNamespace()
		}
	default:
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
