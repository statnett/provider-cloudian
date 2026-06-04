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

package v1alpha1

// +kubebuilder:object:generate=true

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// UserQualityOfServiceLimitsParameters are the configurable fields of a UserQualityOfServiceLimits.
type UserQualityOfServiceLimitsParameters struct {
	// GroupID of the quality of service limits.
	// +optional
	// +immutable
	GroupID string `json:"groupId,omitempty"`

	// UserID of the quality of service limits.
	// +optional
	// +immutable
	UserID string `json:"userId,omitempty"`

	// UserIDRef references a user to retrieve its groupId and userId.
	// +optional
	// +immutable
	UserIDRef *xpv2.Reference `json:"userIdRef,omitempty"`

	// UserIDSelector selects a user to retrieve its groupId and userId.
	// +optional
	UserIDSelector *xpv2.Selector `json:"userIdSelector,omitempty"`

	// Region in which to apply the quality of service limits. Default region if unspecified.
	// +optional
	Region string `json:"region,omitempty"`

	QOS `json:",inline"`
}

// UserQualityOfServiceLimitsObservation are the observable fields of a UserQualityOfServiceLimits.
type UserQualityOfServiceLimitsObservation struct {
}

// A UserQualityOfServiceLimitsStatus represents the observed state of a UserQualityOfServiceLimits.
type UserQualityOfServiceLimitsStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 UserQualityOfServiceLimitsObservation `json:"atProvider,omitempty"`
}
