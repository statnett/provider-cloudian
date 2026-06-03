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

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"

	userv1alpha1common "github.com/statnett/provider-cloudian/apis/common/user/v1alpha1"
)

// A UserQualityOfServiceLimitsSpec defines the desired state of a UserQualityOfServiceLimits.
type UserQualityOfServiceLimitsSpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     userv1alpha1common.UserQualityOfServiceLimitsParameters `json:"forProvider"`
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

	Spec   UserQualityOfServiceLimitsSpec                      `json:"spec"`
	Status userv1alpha1common.UserQualityOfServiceLimitsStatus `json:"status,omitempty"`
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
