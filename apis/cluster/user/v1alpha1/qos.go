package v1alpha1

import resource "k8s.io/apimachinery/pkg/api/resource"

// +kubebuilder:validation:Pattern=`^(0|((0|[1-9][0-9]*)[KMGT]i))$`
type Quantity string

func (q *Quantity) ToKiB() (*int64, error) {
	if q == nil {
		return nil, nil
	}

	rq, err := resource.ParseQuantity(string(*q))
	if err != nil {
		return nil, err
	}

	i := rq.ScaledValue(0) / 1024
	return &i, nil
}

// QualityOfService configures data limits. The value -1 indicates unlimited.
type QualityOfServiceLimits struct {
	// StorageQuotaBytes is the limit for total stored data in bytes.
	// +optional
	// +nullable
	StorageQuotaBytes *Quantity `json:"storageQuotaBytes"`
	// StorageQuotaCount is the limit for total number of objects.
	// +optional
	// +nullable
	StorageQuotaCount *uint32 `json:"storageQuotaCount"`
	// RequestsPerMin is the limit for number of HTTP requests per minute.
	// +optional
	// +nullable
	RequestsPerMin *uint32 `json:"requestsPerMin"`
	// InboundBytesPerMin is the limit for inbound data per minute in bytes.
	// +optional
	// +nullable
	InboundBytesPerMin *Quantity `json:"inboundBytesPerMin"`
	// OutboundKiBsPerMin is the limit for outbound data per minute in bytes.
	// +optional
	// +nullable
	OutboundBytesPerMin *Quantity `json:"outboundBytesPerMin"`
}

type QOS struct {
	// Warning is the soft limit that triggers a warning.
	// +optional
	Warning *QualityOfServiceLimits `json:"warning,omitempty"`

	// Hard is the hard limit.
	// +optional
	Hard *QualityOfServiceLimits `json:"hard,omitempty"`
}
