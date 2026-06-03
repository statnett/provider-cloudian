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

// GroupQualityOfServiceLimitsParameters are the configurable fields of a GroupQualityOfServiceLimits.
type GroupQualityOfServiceLimitsParameters struct {
	// GroupID of the quality of service limits.
	// +optional
	// +immutable
	GroupID string `json:"groupId,omitempty"`

	// GroupIDRef references a group to retrieve its groupId.
	// +optional
	// +immutable
	GroupIDRef *xpv2.Reference `json:"groupIdRef,omitempty"`

	// GroupIDSelector selects a group to retrieve its groupId.
	// +optional
	GroupIDSelector *xpv2.Selector `json:"groupIdSelector,omitempty"`

	// Region in which to apply the quality of service limits. Default region if unspecified.
	// +optional
	Region string `json:"region,omitempty"`

	QOS `json:",inline"`
}

// GroupQualityOfServiceLimitsObservation are the observable fields of a GroupQualityOfServiceLimits.
type GroupQualityOfServiceLimitsObservation struct {
}

// A GroupQualityOfServiceLimitsStatus represents the observed state of a GroupQualityOfServiceLimits.
type GroupQualityOfServiceLimitsStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 GroupQualityOfServiceLimitsObservation `json:"atProvider,omitempty"`
}
