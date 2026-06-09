package accesskey

import (
	"fmt"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"

	"github.com/statnett/provider-cloudian/internal/sdk/cloudian"
)

func ConnectionDetails(creds *cloudian.SecurityInfo) managed.ConnectionDetails {
	return managed.ConnectionDetails{
		"secretKey": []byte(creds.SecretKey),
		"config.toml": []byte(fmt.Sprintf(
			`[default]
aws_access_key_id = %s
aws_secret_access_key = %s`,
			creds.AccessKey,
			creds.SecretKey,
		)),
	}
}
