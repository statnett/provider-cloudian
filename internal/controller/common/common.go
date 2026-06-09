package common

import "github.com/statnett/provider-cloudian/internal/sdk/cloudian"

func NewCloudianService(providerConfigEndpoint string, authHeader string) (*cloudian.Client, error) {
	return cloudian.NewClient(
		providerConfigEndpoint,
		authHeader,
	), nil
}
