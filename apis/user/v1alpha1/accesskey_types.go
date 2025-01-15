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

// AccessKeyParameters are the configurable fields of a AccessKey.
type AccessKeyParameters struct {
	// GroupID for the new access key.
	// +optional
	// +immutable
	GroupID string `json:"groupId,omitempty"`

	// User for the new access key.
	// +optional
	// +immutable
	UserID string `json:"userId,omitempty"`

	// UserRef is a reference to a user to retrieve its userId.
	// +optional
	// +immutable
	UserRef *xpv1.Reference `json:"userRef,omitempty"`

	// UserSelector selects reference to a user to retrieve its userId.
	// +optional
	UserSelector *xpv1.Selector `json:"userSelector,omitempty"`
}

// AccessKeyObservation are the observable fields of a AccessKey.
type AccessKeyObservation struct {
	// AccessKey is the S3 Access Key ID, with a corresponding SecretKey.
	AccessKey string `json:"accessKey,omitempty"`
}

// A AccessKeySpec defines the desired state of a AccessKey.
type AccessKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AccessKeyParameters `json:"forProvider"`
}

// A AccessKeyStatus represents the observed state of a AccessKey.
type AccessKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AccessKeyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AccessKey is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudian}
type AccessKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessKeySpec   `json:"spec"`
	Status AccessKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessKeyList contains a list of AccessKey
type AccessKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessKey `json:"items"`
}

// AccessKey type metadata.
var (
	AccessKeyKind             = reflect.TypeOf(AccessKey{}).Name()
	AccessKeyGroupKind        = schema.GroupKind{Group: MetadataGroup, Kind: AccessKeyKind}.String()
	AccessKeyKindAPIVersion   = AccessKeyKind + "." + SchemeGroupVersion.String()
	AccessKeyGroupVersionKind = SchemeGroupVersion.WithKind(AccessKeyKind)
)

func init() {
	SchemeBuilder.Register(&AccessKey{}, &AccessKeyList{})
}
