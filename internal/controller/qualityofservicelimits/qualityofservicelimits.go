package qualityofservicelimits

import (
	"k8s.io/utils/ptr"

	"github.com/statnett/provider-cloudian/apis/user/v1alpha1"
	"github.com/statnett/provider-cloudian/internal/sdk/cloudian"
)

func ToCloudianQOS(qos v1alpha1.QOS) (cloudian.QualityOfService, error) {
	var err error
	cQOS := cloudian.QualityOfService{}

	if cQOS.Warning, err = ToCloudianLimits(qos.Warning); err != nil {
		return cloudian.QualityOfService{}, err
	}
	if cQOS.Hard, err = ToCloudianLimits(qos.Hard); err != nil {
		return cloudian.QualityOfService{}, err
	}
	return cQOS, nil
}

func ToCloudianLimits(limits *v1alpha1.QualityOfServiceLimits) (cloudian.QualityOfServiceLimits, error) {
	if limits == nil {
		return cloudian.QualityOfServiceLimits{}, nil
	}

	var err error
	qosl := cloudian.QualityOfServiceLimits{}

	if qosl.StorageQuotaKiBs, err = limits.StorageQuotaBytes.ToKiB(); err != nil {
		return cloudian.QualityOfServiceLimits{}, err
	}
	if qosl.InboundKiBsPerMin, err = limits.InboundBytesPerMin.ToKiB(); err != nil {
		return cloudian.QualityOfServiceLimits{}, err
	}
	if qosl.OutboundKiBsPerMin, err = limits.OutboundBytesPerMin.ToKiB(); err != nil {
		return cloudian.QualityOfServiceLimits{}, err
	}

	if limits.StorageQuotaCount != nil {
		qosl.StorageQuotaCount = ptr.To(int64(*limits.StorageQuotaCount))
	}
	if limits.RequestsPerMin != nil {
		qosl.RequestsPerMin = ptr.To(int64(*limits.RequestsPerMin))
	}

	return qosl, nil
}
