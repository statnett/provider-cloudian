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

package group

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/statnett/provider-cloudian/apis/user/v1alpha1"
	"github.com/statnett/provider-cloudian/internal/sdk/cloudian"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

func TestObserve(t *testing.T) {
	type fields struct {
		_ interface{}
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		// TODO: Add test cases.
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{cloudianService: nil}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	tests := []struct {
		name                   string
		desired                v1alpha1.GroupParameters
		observed               cloudian.Group
		wantConsideredUpToDate bool
	}{
		{
			name: "GroupID and GroupName is set",
			desired: v1alpha1.GroupParameters{
				Active:    true, // This is the kube spec default
				GroupID:   "QA",
				GroupName: "",
			},
			observed: cloudian.Group{
				Active:             true,
				GroupID:            "QA",
				GroupName:          "",
				LDAPEnabled:        false,
				LDAPGroup:          "",
				LDAPMatchAttribute: "",
				LDAPSearch:         "",
				LDAPSearchUserBase: "",
				LDAPServerURL:      "",
				LDAPUserDNTemplate: "",
			},
			wantConsideredUpToDate: true,
		},
		{
			name: "GroupID has changed",
			desired: v1alpha1.GroupParameters{
				Active:    true, // This is the kube spec default
				GroupID:   "QA",
				GroupName: "",
			},
			observed: cloudian.Group{
				Active:             true,
				GroupID:            "QA2",
				GroupName:          "",
				LDAPEnabled:        false,
				LDAPGroup:          "",
				LDAPMatchAttribute: "",
				LDAPSearch:         "",
				LDAPSearchUserBase: "",
				LDAPServerURL:      "",
				LDAPUserDNTemplate: "",
			},
			wantConsideredUpToDate: false,
		},
		{
			name: "GroupName has changed",
			desired: v1alpha1.GroupParameters{
				Active:    true, // This is the kube spec default
				GroupID:   "QA",
				GroupName: "A",
			},
			observed: cloudian.Group{
				Active:             true,
				GroupID:            "QA",
				GroupName:          "",
				LDAPEnabled:        false,
				LDAPGroup:          "",
				LDAPMatchAttribute: "",
				LDAPSearch:         "",
				LDAPSearchUserBase: "",
				LDAPServerURL:      "",
				LDAPUserDNTemplate: "",
			},
			wantConsideredUpToDate: false,
		},
		{
			name: "Active has changed",
			desired: v1alpha1.GroupParameters{
				Active:    true, // This is the kube spec default
				GroupID:   "QA",
				GroupName: "A",
			},
			observed: cloudian.Group{
				Active:             false,
				GroupID:            "QA",
				GroupName:          "",
				LDAPEnabled:        false,
				LDAPGroup:          "",
				LDAPMatchAttribute: "",
				LDAPSearch:         "",
				LDAPSearchUserBase: "",
				LDAPServerURL:      "",
				LDAPUserDNTemplate: "",
			},
			wantConsideredUpToDate: false,
		},
		{
			name: "LDAPEnabled has changed",
			desired: v1alpha1.GroupParameters{
				Active:      true, // This is the kube spec default
				GroupID:     "QA",
				GroupName:   "A",
				LDAPEnabled: ptr.To(true),
			},
			observed: cloudian.Group{
				Active:             false,
				GroupID:            "QA",
				GroupName:          "",
				LDAPEnabled:        false,
				LDAPGroup:          "A",
				LDAPMatchAttribute: "",
				LDAPSearch:         "",
				LDAPSearchUserBase: "",
				LDAPServerURL:      "",
				LDAPUserDNTemplate: "",
			},
			wantConsideredUpToDate: false,
		},
		{
			name: "MR has a lot of values set, but matches observed values",
			desired: v1alpha1.GroupParameters{
				Active:             false,
				GroupID:            "QA",
				GroupName:          "X",
				LDAPEnabled:        ptr.To(true),
				LDAPGroup:          ptr.To("G"),
				LDAPMatchAttribute: ptr.To("A"),
				LDAPSearch:         ptr.To("B"),
				LDAPSearchUserBase: ptr.To("C"),
				LDAPServerURL:      ptr.To("D"),
				LDAPUserDNTemplate: ptr.To("E"),
			},
			observed: cloudian.Group{
				Active:             false,
				GroupID:            "QA",
				GroupName:          "X",
				LDAPEnabled:        true,
				LDAPGroup:          "G",
				LDAPMatchAttribute: "A",
				LDAPSearch:         "B",
				LDAPSearchUserBase: "C",
				LDAPServerURL:      "D",
				LDAPUserDNTemplate: "E",
			},
			wantConsideredUpToDate: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isUpToDate, diff := isUpToDate(tt.desired, tt.observed)
			if isUpToDate != tt.wantConsideredUpToDate {
				t.Errorf("isUpToDate() = %v, want %v, but the diff was %s", isUpToDate, tt.wantConsideredUpToDate, diff)
			}
		})
	}
}
