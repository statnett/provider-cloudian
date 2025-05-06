/*
Copyright 2025 The Crossplane Authors.

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
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this User
func (mg *User) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.GroupID,
		Reference:    mg.Spec.ForProvider.GroupIDRef,
		Selector:     mg.Spec.ForProvider.GroupIDSelector,
		To:           reference.To{Managed: &Group{}, List: &GroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to resolve spec.forProvider.groupId")
	}

	mg.Spec.ForProvider.GroupID = rsp.ResolvedValue
	mg.Spec.ForProvider.GroupIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this GroupQualityOfServiceLimits
func (mg *GroupQualityOfServiceLimits) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.GroupID,
		Reference:    mg.Spec.ForProvider.GroupIDRef,
		Selector:     mg.Spec.ForProvider.GroupIDSelector,
		To:           reference.To{Managed: &Group{}, List: &GroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to resolve spec.forProvider.groupId")
	}

	mg.Spec.ForProvider.GroupID = rsp.ResolvedValue
	mg.Spec.ForProvider.GroupIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this AccessKey
func (mg *AccessKey) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.UserID,
		Reference:    mg.Spec.ForProvider.UserIDRef,
		Selector:     mg.Spec.ForProvider.UserIDSelector,
		To:           reference.To{Managed: &User{}, List: &UserList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to resolve spec.forProvider.userId")
	}

	mg.Spec.ForProvider.UserID = rsp.ResolvedValue
	mg.Spec.ForProvider.UserIDRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		Reference: mg.Spec.ForProvider.UserIDRef,
		Selector:  mg.Spec.ForProvider.UserIDSelector,
		To:        reference.To{Managed: &User{}, List: &UserList{}},
		Extract: func(mg resource.Managed) string {
			user, ok := mg.(*User)
			if !ok {
				return ""
			}
			return user.Spec.ForProvider.GroupID
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to resolve spec.forProvider.groupId")
	}

	mg.Spec.ForProvider.GroupID = rsp.ResolvedValue

	return nil
}

// ResolveReferences of this UserQualityOfServiceLimits
func (mg *UserQualityOfServiceLimits) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.UserID,
		Reference:    mg.Spec.ForProvider.UserIDRef,
		Selector:     mg.Spec.ForProvider.UserIDSelector,
		To:           reference.To{Managed: &User{}, List: &UserList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to resolve spec.forProvider.userId")
	}

	mg.Spec.ForProvider.UserID = rsp.ResolvedValue
	mg.Spec.ForProvider.UserIDRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		Reference: mg.Spec.ForProvider.UserIDRef,
		Selector:  mg.Spec.ForProvider.UserIDSelector,
		To:        reference.To{Managed: &User{}, List: &UserList{}},
		Extract: func(mg resource.Managed) string {
			user, ok := mg.(*User)
			if !ok {
				return ""
			}
			return user.Spec.ForProvider.GroupID
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to resolve spec.forProvider.groupId")
	}

	mg.Spec.ForProvider.GroupID = rsp.ResolvedValue

	return nil
}
