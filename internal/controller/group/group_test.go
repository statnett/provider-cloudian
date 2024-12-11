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

func TestGroupIsConsideredUpToDate(t *testing.T) {
	tests := []struct {
		name              string
		desired, observed v1alpha1.GroupParameters
		wantEquality      bool
	}{
		{
			name: "GroupID and GroupName is set",
			desired: v1alpha1.GroupParameters{
				GroupID:   "QA",
				GroupName: ptr.To("Hello"),
			},
			observed: v1alpha1.GroupParameters{
				Active:             ptr.To("true"),
				GroupID:            "QA",
				GroupName:          ptr.To("Hello"),
				LDAPEnabled:        ptr.To(false),
				LDAPGroup:          nil,
				LDAPMatchAttribute: nil,
				LDAPSearch:         nil,
				LDAPSearchUserBase: nil,
				LDAPServerURL:      nil,
				LDAPUserDNTemplate: nil,
				S3EndpointsHTTP:    []string{"ALL"},
				S3EndpointsHTTPS:   []string{"ALL"},
				S3WebsiteEndpoints: []string{"ALL"},
			},
			wantEquality: true,
		},
		{
			name: "GroupID is set GroupName is not set",
			desired: v1alpha1.GroupParameters{
				GroupID: "QA",
			},
			observed: v1alpha1.GroupParameters{
				Active:             ptr.To("true"),
				GroupID:            "QA",
				GroupName:          ptr.To(""),
				LDAPEnabled:        ptr.To(false),
				LDAPGroup:          nil,
				LDAPMatchAttribute: nil,
				LDAPSearch:         nil,
				LDAPSearchUserBase: nil,
				LDAPServerURL:      nil,
				LDAPUserDNTemplate: nil,
				S3EndpointsHTTP:    []string{"ALL"},
				S3EndpointsHTTPS:   []string{"ALL"},
				S3WebsiteEndpoints: []string{"ALL"},
			},
			wantEquality: true,
		},
		{
			name: "Desired GroupID is unlike observed GroupID",
			desired: v1alpha1.GroupParameters{
				GroupID: "desired",
			},
			observed: v1alpha1.GroupParameters{
				Active:             ptr.To("true"),
				GroupID:            "observed",
				GroupName:          ptr.To(""),
				LDAPEnabled:        ptr.To(false),
				LDAPGroup:          nil,
				LDAPMatchAttribute: nil,
				LDAPSearch:         nil,
				LDAPSearchUserBase: nil,
				LDAPServerURL:      nil,
				LDAPUserDNTemplate: nil,
				S3EndpointsHTTP:    []string{"ALL"},
				S3EndpointsHTTPS:   []string{"ALL"},
				S3WebsiteEndpoints: []string{"ALL"},
			},
			wantEquality: false,
		},
		{
			name: "Desired GroupName is unlike observed GroupName",
			desired: v1alpha1.GroupParameters{
				GroupID:   "desired",
				GroupName: ptr.To("desired description"),
			},
			observed: v1alpha1.GroupParameters{
				Active:             ptr.To("true"),
				GroupID:            "desired",
				GroupName:          ptr.To(""),
				LDAPEnabled:        ptr.To(false),
				LDAPGroup:          nil,
				LDAPMatchAttribute: nil,
				LDAPSearch:         nil,
				LDAPSearchUserBase: nil,
				LDAPServerURL:      nil,
				LDAPUserDNTemplate: nil,
				S3EndpointsHTTP:    []string{"ALL"},
				S3EndpointsHTTPS:   []string{"ALL"},
				S3WebsiteEndpoints: []string{"ALL"},
			},
			wantEquality: false,
		},
		{
			name: "We have not set a GroupName in the desired state, and the observed GroupName is empty string",
			desired: v1alpha1.GroupParameters{
				GroupID:   "desired",
				GroupName: nil,
			},
			observed: v1alpha1.GroupParameters{
				Active:             ptr.To("true"),
				GroupID:            "desired",
				GroupName:          ptr.To(""),
				LDAPEnabled:        ptr.To(false),
				LDAPGroup:          nil,
				LDAPMatchAttribute: nil,
				LDAPSearch:         nil,
				LDAPSearchUserBase: nil,
				LDAPServerURL:      nil,
				LDAPUserDNTemplate: nil,
				S3EndpointsHTTP:    []string{"ALL"},
				S3EndpointsHTTPS:   []string{"ALL"},
				S3WebsiteEndpoints: []string{"ALL"},
			},
			wantEquality: true,
		},
		{
			name: "We can desire a LDAPServerURL and get a diff if observed does not have one",
			desired: v1alpha1.GroupParameters{
				GroupID:       "desired",
				GroupName:     nil,
				LDAPServerURL: ptr.To("ldap://example.com"),
			},
			observed: v1alpha1.GroupParameters{
				Active:             ptr.To("true"),
				GroupID:            "desired",
				GroupName:          ptr.To(""),
				LDAPEnabled:        ptr.To(false),
				LDAPGroup:          nil,
				LDAPMatchAttribute: nil,
				LDAPSearch:         nil,
				LDAPSearchUserBase: nil,
				LDAPServerURL:      nil,
				LDAPUserDNTemplate: nil,
				S3EndpointsHTTP:    []string{"ALL"},
				S3EndpointsHTTPS:   []string{"ALL"},
				S3WebsiteEndpoints: []string{"ALL"},
			},
			wantEquality: false,
		},
		{
			name: "We can desire some s3urls and this will lead to a diff against observed [\"ALL\"]",
			desired: v1alpha1.GroupParameters{
				GroupID:       "desired",
				GroupName:     nil,
				S3EndpointsHTTP: []string{"oslo,bergen"},
				S3EndpointsHTTPS: []string{"oslo,bergen"},
				S3WebsiteEndpoints: []string{"oslo,bergen"},
			},
			observed: v1alpha1.GroupParameters{
				Active:             ptr.To("true"),
				GroupID:            "desired",
				GroupName:          ptr.To(""),
				LDAPEnabled:        ptr.To(false),
				LDAPGroup:          nil,
				LDAPMatchAttribute: nil,
				LDAPSearch:         nil,
				LDAPSearchUserBase: nil,
				LDAPServerURL:      nil,
				LDAPUserDNTemplate: nil,
				S3EndpointsHTTP:    []string{"ALL"},
				S3EndpointsHTTPS:   []string{"ALL"},
				S3WebsiteEndpoints: []string{"ALL"},
			},
			wantEquality: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isUpToDate, diff, err := isUpToDate(tt.desired, tt.observed)
			if err != nil {
				t.Errorf("isUpToDate() error = %v", err)
			}
			if isUpToDate != tt.wantEquality {
				t.Errorf("isUpToDate() = %v, want %v, but the diff was %s", isUpToDate, tt.wantEquality, diff)
			}
		})
	}
}
