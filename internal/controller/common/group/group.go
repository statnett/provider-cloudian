package group

import (
	"k8s.io/utils/ptr"

	userv1alpha1common "github.com/statnett/provider-cloudian/apis/common/user/v1alpha1"
	"github.com/statnett/provider-cloudian/internal/sdk/cloudian"
)

func IsUpToDate(name string, desired userv1alpha1common.GroupParameters, observed cloudian.Group) bool {
	return NewCloudianGroup(name, desired) == observed
}

func NewCloudianGroup(name string, gp userv1alpha1common.GroupParameters) cloudian.Group {
	return cloudian.Group{
		Active:             gp.Active,
		GroupID:            name,
		GroupName:          gp.GroupName,
		LDAPEnabled:        ptr.Deref(gp.LDAPEnabled, false),
		LDAPGroup:          ptr.Deref(gp.LDAPGroup, ""),
		LDAPMatchAttribute: ptr.Deref(gp.LDAPMatchAttribute, ""),
		LDAPSearch:         ptr.Deref(gp.LDAPSearch, ""),
		LDAPSearchUserBase: ptr.Deref(gp.LDAPSearchUserBase, ""),
		LDAPServerURL:      ptr.Deref(gp.LDAPServerURL, ""),
		LDAPUserDNTemplate: ptr.Deref(gp.LDAPUserDNTemplate, ""),
	}
}
