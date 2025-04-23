package main

import (
	"fmt"
	"log"
	"strings"

	ladon "github.com/ory/ladon"
	registry_auth "github.com/zhangyoufu/registry_auth"
)

type Authorizer struct {
	warden ladon.Warden
}

func (authz *Authorizer) Authorize(r *registry_auth.AuthorizationRequest) (registry_auth.Scope, error) {
	var b strings.Builder
	_, _ = b.WriteString(fmt.Sprintf("ip=%s username=%s service=%s", r.IP, r.Username, r.Service))
	for i, resourceScope := range r.Scope {
		_, _ = b.WriteString(fmt.Sprintf("\nscope[%d]: type=%q name=%q actions=%q", i, resourceScope.Type, resourceScope.Name, resourceScope.Actions))
	}
	log.Print(b.String())

	authorizedScope := registry_auth.Scope{}
	for _, resourceScope := range r.Scope {
		_r := ladon.Request{
			Resource: resourceScope.Type + ":" + resourceScope.Name,
			Subject:  r.Username,
			Context: ladon.Context{
				"ip":      r.IP,
				"service": r.Service,
			},
		}
		if len(resourceScope.Actions) == 0 {
			if err := authz.warden.IsAllowed(r.Context, &_r); err != nil {
				continue
			}
		}
		authorizedActions := []string{}
		for _, action := range resourceScope.Actions {
			_r.Action = action
			if err := authz.warden.IsAllowed(r.Context, &_r); err != nil {
				continue
			}
			authorizedActions = append(authorizedActions, action)
		}
		if len(authorizedActions) == 0 {
			continue
		}
		authorizedScope = append(authorizedScope, &registry_auth.ResourceScope{
			Type:    resourceScope.Type,
			Name:    resourceScope.Name,
			Actions: authorizedActions,
		})
	}

	// len(authorizedScope) == 0 is still valid
	// docker login use this to check whether credential is valid
	return authorizedScope, nil
}
