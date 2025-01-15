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

// GroupParameters are the configurable fields of a Group.
type GroupParameters struct {
	// Active determines whether the group is enabled (true) or disabled (false) in the system.
	//+optional
	//+kubebuilder:default=true
	Active bool `json:"active"`
	// GroupName is the group name (known as Description in the GUI).
	//+optional
	//+kubebuilder:validation:MaxLength=64
	GroupName string `json:"groupName,omitempty"`
	// LDAPEnabled determines whether LDAP authentication is enabled for members of this group.
	//+optional
	//+kubebuilder:default=false
	LDAPEnabled *bool `json:"ldapEnabled,omitempty"`
	//+optional
	// LDAPGroup us the group's name from the LDAP system.
	LDAPGroup *string `json:"ldapGroup,omitempty"`
	//+optional
	LDAPMatchAttribute *string `json:"ldapMatchAttribute,omitempty"`
	//+optional
	LDAPSearch *string `json:"ldapSearch,omitempty"`
	// LDAPSearchUserBase specifies the LDAP search base from which the CMC should start when retrieving the user's LDAP record in order to apply filtering.
	//+optional
	LDAPSearchUserBase *string `json:"ldapSearchUserBase,omitempty"`
	//+optional
	// LDAPServerURL specifies the URL that the CMC should use to access the LDAP Server when authenticating users in this group.
	LDAPServerURL *string `json:"ldapServerURL,omitempty"`
	// LDAPUserDNTemplate specifies how users within this group will be authenticated against the LDAP system when they log into the CMC.
	//+optional
	LDAPUserDNTemplate *string `json:"ldapUserDNTemplate,omitempty"`
}

// GroupObservation are the observable fields of a Group.
type GroupObservation struct {
}

// A GroupSpec defines the desired state of a Group.
type GroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupParameters `json:"forProvider"`
}

// A GroupStatus represents the observed state of a Group.
type GroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Group is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudian}
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList contains a list of Group
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

// Group type metadata.
var (
	GroupKind             = reflect.TypeOf(Group{}).Name()
	GroupGroupKind        = schema.GroupKind{Group: MetadataGroup, Kind: GroupKind}.String()
	GroupKindAPIVersion   = GroupKind + "." + SchemeGroupVersion.String()
	GroupGroupVersionKind = SchemeGroupVersion.WithKind(GroupKind)
)

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}
