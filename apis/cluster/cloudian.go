/*
Copyright 2020 The Crossplane Authors.

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

// Package cluster contains Kubernetes API for the Cloudian provider.
package cluster

import (
	"k8s.io/apimachinery/pkg/runtime"

	userv1alpha1 "github.com/statnett/provider-cloudian/apis/cluster/user/v1alpha1"
	cloudianv1alpha1 "github.com/statnett/provider-cloudian/apis/cluster/v1alpha1"
)

var (
	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme may be used to add all cluster-scoped resources defined in the project to a Scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	for _, sb := range []runtime.SchemeBuilder{cloudianv1alpha1.SchemeBuilder, userv1alpha1.SchemeBuilder} {
		if err := sb.AddToScheme(scheme); err != nil {
			return err
		}
	}

	return nil
}
