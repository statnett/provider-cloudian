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

// A GroupStatus represents the observed state of a Group.
type GroupStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 GroupObservation `json:"atProvider,omitempty"`
}
