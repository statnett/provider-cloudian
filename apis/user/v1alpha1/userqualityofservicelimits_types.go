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

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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
	UserIDRef *xpv1.Reference `json:"userIdRef,omitempty"`

	// UserIDSelector selects a user to retrieve its groupId and userId.
	// +optional
	UserIDSelector *xpv1.Selector `json:"userIdSelector,omitempty"`

	// Region in which to apply the quality of service limits. Default region if unspecified.
	// +optional
	Region string `json:"region,omitempty"`

	QOS `json:",inline"`
}

// UserQualityOfServiceLimitsObservation are the observable fields of a UserQualityOfServiceLimits.
type UserQualityOfServiceLimitsObservation struct {
}

// A UserQualityOfServiceLimitsSpec defines the desired state of a UserQualityOfServiceLimits.
type UserQualityOfServiceLimitsSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       UserQualityOfServiceLimitsParameters `json:"forProvider"`
}

// A UserQualityOfServiceLimitsStatus represents the observed state of a UserQualityOfServiceLimits.
type UserQualityOfServiceLimitsStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          UserQualityOfServiceLimitsObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// UserQualityOfServiceLimits represents the quality of service limits for a Cloudian user, within a region.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudian}
type UserQualityOfServiceLimits struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserQualityOfServiceLimitsSpec   `json:"spec"`
	Status UserQualityOfServiceLimitsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserQualityOfServiceLimitsList contains a list of UserQualityOfServiceLimits
type UserQualityOfServiceLimitsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserQualityOfServiceLimits `json:"items"`
}

// UserQualityOfServiceLimits type metadata.
var (
	UserQualityOfServiceLimitsKind             = reflect.TypeOf(UserQualityOfServiceLimits{}).Name()
	UserQualityOfServiceLimitsGroupKind        = schema.GroupKind{Group: MetadataGroup, Kind: UserQualityOfServiceLimitsKind}.String()
	UserQualityOfServiceLimitsKindAPIVersion   = UserQualityOfServiceLimitsKind + "." + SchemeGroupVersion.String()
	UserQualityOfServiceLimitsGroupVersionKind = SchemeGroupVersion.WithKind(UserQualityOfServiceLimitsKind)
)

func init() {
	SchemeBuilder.Register(&UserQualityOfServiceLimits{}, &UserQualityOfServiceLimitsList{})
}
