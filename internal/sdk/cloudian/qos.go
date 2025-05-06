package cloudian

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

const DefaultRegion = ""

// QualityOfService configures data limits for a Group or User.
type QualityOfService struct {
	// Warning is the soft limit that triggers a warning.
	Warning QualityOfServiceLimits
	// Hard is the hard limit.
	Hard QualityOfServiceLimits
}

// QualityOfService configures data limits.
type QualityOfServiceLimits struct {
	// StorageQuotaKiBs is the limit for total stored data in KiB.
	StorageQuotaKiBs *int64
	// StorageQuotaCount is the limit for total number of objects.
	StorageQuotaCount *int64
	// RequestsPerMin is the limit for number of HTTP requests per minute.
	RequestsPerMin *int64
	// InboundKiBsPerMin is the limit for inbound data per minute in KiB.
	InboundKiBsPerMin *int64
	// OutboundKiBsPerMin is the limit for outbound data per minute in KiB.
	OutboundKiBsPerMin *int64
}

func (a *QualityOfServiceLimits) Equal(b QualityOfServiceLimits) bool {
	// k8s.io/utils/ptr Equal
	eq := func(a, b *int64) bool {
		if (a == nil) != (b == nil) {
			return false
		}
		if a == nil {
			return true
		}
		return *a == *b
	}
	return eq(a.InboundKiBsPerMin, b.InboundKiBsPerMin) &&
		eq(a.OutboundKiBsPerMin, b.OutboundKiBsPerMin) &&
		eq(a.RequestsPerMin, b.RequestsPerMin) &&
		eq(a.StorageQuotaCount, b.StorageQuotaCount) &&
		eq(a.StorageQuotaKiBs, b.StorageQuotaKiBs)
}

func (qos *QualityOfService) allMinusOne() bool {
	return qos.Warning.allMinusOne() && qos.Hard.allMinusOne()
}

func (l *QualityOfServiceLimits) allMinusOne() bool {
	return (l.StorageQuotaKiBs == nil || *l.StorageQuotaKiBs == -1) &&
		(l.StorageQuotaCount == nil || *l.StorageQuotaCount == -1) &&
		(l.RequestsPerMin == nil || *l.RequestsPerMin == -1) &&
		(l.InboundKiBsPerMin == nil || *l.InboundKiBsPerMin == -1) &&
		(l.OutboundKiBsPerMin == nil || *l.OutboundKiBsPerMin == -1)
}

// nolint: gocyclo
func (qos *QualityOfService) unmarshalQOSList(raw []byte) error {
	var data struct {
		QOSLimitList []struct {
			Type  string `json:"type"`
			Value int64  `json:"value"`
		} `json:"qosLimitList"`
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	for _, item := range data.QOSLimitList {
		if item.Value == -1 {
			continue
		}

		v := &item.Value
		switch item.Type {
		case "STORAGE_QUOTA_KBYTES_LH":
			qos.Hard.StorageQuotaKiBs = v
		case "STORAGE_QUOTA_KBYTES_LW":
			qos.Warning.StorageQuotaKiBs = v
		case "STORAGE_QUOTA_COUNT_LH":
			qos.Hard.StorageQuotaCount = v
		case "STORAGE_QUOTA_COUNT_LW":
			qos.Warning.StorageQuotaCount = v
		case "REQUEST_RATE_LH":
			qos.Hard.RequestsPerMin = v
		case "REQUEST_RATE_LW":
			qos.Warning.RequestsPerMin = v
		case "DATAKBYTES_IN_LH":
			qos.Hard.InboundKiBsPerMin = v
		case "DATAKBYTES_IN_LW":
			qos.Warning.InboundKiBsPerMin = v
		case "DATAKBYTES_OUT_LH":
			qos.Hard.OutboundKiBsPerMin = v
		case "DATAKBYTES_OUT_LW":
			qos.Warning.OutboundKiBsPerMin = v
		}
	}
	return nil
}

func (qos *QualityOfService) rawQueryParams() map[string]*int64 {
	return map[string]*int64{
		"hlStorageQuotaKBytes": qos.Hard.StorageQuotaKiBs,
		"wlStorageQuotaKBytes": qos.Warning.StorageQuotaKiBs,
		"hlStorageQuotaCount":  qos.Hard.StorageQuotaCount,
		"wlStorageQuotaCount":  qos.Warning.StorageQuotaCount,
		"hlRequestRate":        qos.Hard.RequestsPerMin,
		"wlRequestRate":        qos.Warning.RequestsPerMin,
		"hlDataKBytesIn":       qos.Hard.InboundKiBsPerMin,
		"wlDataKBytesIn":       qos.Warning.InboundKiBsPerMin,
		"hlDataKBytesOut":      qos.Hard.OutboundKiBsPerMin,
		"wlDataKBytesOut":      qos.Warning.OutboundKiBsPerMin,
	}
}

func (qos *QualityOfService) queryParams(params map[string]string) error {
	for key, raw := range qos.rawQueryParams() {
		val := int64(-1)
		if raw != nil {
			val = *raw
		}
		params[key] = strconv.FormatInt(val, 10)
	}
	return nil
}

// SetQOS sets QualityOfService limits for a Group or User, depending on the value of GroupID and UserID.
//
// User-level QoS for a specific user (GroupID="<groupId>", UserID="<userId>")
// Default user-level QoS for a specific group (GroupID="<groupId>", UserID="ALL")
// Default user-level QoS for the whole region (GroupID="*", UserID="ALL")
// Group-level QoS for a specific group (GroupID="<groupId>", UserID="*")
// Default group-level QoS for the whole region (GroupID="ALL", UserID="*")
func (client Client) SetQOS(ctx context.Context, guid GroupUserID, region string, qos QualityOfService) error {
	for _, val := range qos.rawQueryParams() {
		if val != nil && *val < -1 {
			return fmt.Errorf("QoS limit values must be >= -1")
		}
	}

	if qos.allMinusOne() {
		existanceMarker := int64(-2)
		qos.Warning.RequestsPerMin = &existanceMarker
	}

	params := make(map[string]string)
	if err := qos.queryParams(params); err != nil {
		return err
	}

	if region != DefaultRegion {
		params["region"] = region
	}

	resp, err := client.newRequest(ctx).
		SetQueryParam("userId", guid.UserID).
		SetQueryParam("groupId", guid.GroupID).
		SetQueryParams(params).
		Post("/qos/limits")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return nil
	default:
		return fmt.Errorf("POST quota unexpected status: %d", resp.StatusCode())
	}
}

// SetQOS gets QualityOfService limits for a Group or User, depending on the value of GroupID and UserID.
// See SetQOS for details.
func (client Client) GetQOS(ctx context.Context, guid GroupUserID, region string) (*QualityOfService, error) {
	params := make(map[string]string)
	if region != DefaultRegion {
		params["region"] = region
	}

	resp, err := client.newRequest(ctx).
		SetQueryParam("userId", guid.UserID).
		SetQueryParam("groupId", guid.GroupID).
		SetQueryParams(params).
		Get("/qos/limits")
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case 200:
		qos := &QualityOfService{}
		if err := qos.unmarshalQOSList(resp.Body()); err != nil {
			return nil, err
		}

		if qos.allMinusOne() {
			return nil, ErrNotFound
		}
		if qos.Warning.RequestsPerMin != nil &&
			*qos.Warning.RequestsPerMin == -2 {
			unlimited := int64(-1)
			qos.Warning.RequestsPerMin = &unlimited
		}

		return qos, nil
	default:
		return nil, fmt.Errorf("GET quota unexpected status: %d", resp.StatusCode())
	}
}

// DeleteQOS deletes QualityOfService limits for a Group or User, depending on the value of GroupID and UserID.
// See SetQOS for details.
func (client Client) DeleteQOS(ctx context.Context, guid GroupUserID, region string) error {
	params := make(map[string]string)
	if region != DefaultRegion {
		params["region"] = region
	}

	resp, err := client.newRequest(ctx).
		SetQueryParam("userId", guid.UserID).
		SetQueryParam("groupId", guid.GroupID).
		SetQueryParams(params).
		Delete("/qos/limits")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return nil
	default:
		return fmt.Errorf("DELETE quota unexpected status: %d", resp.StatusCode())
	}
}
