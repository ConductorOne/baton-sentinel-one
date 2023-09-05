package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	grant "github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sentinel-one/pkg/sentinelone"
)

type roleResourceType struct {
	resourceType *v2.ResourceType
	client       *sentinelone.Client
}

const (
	scopeAccountMember = "account scope"
	scopeSiteMember    = "site scope"
	scopeTenantMember  = "tenant scope"
)

var memberships = map[string]string{
	"account": scopeAccountMember,
	"site":    scopeSiteMember,
	"tenant":  scopeTenantMember,
}

func (r *roleResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return r.resourceType
}

// Create a new connector resource for an SentinelOne Role.
func roleResource(ctx context.Context, role *sentinelone.Role) (*v2.Resource, error) {
	var name string
	var id string

	// scope roles and predefined roles have different fields
	if role.RoleName != "" && role.RoleID != "" {
		id = role.RoleID
		name = role.RoleName
	} else {
		id = role.ID
		name = role.Name
	}

	profile := map[string]interface{}{
		"role_name": name,
		"role_id":   id,
	}

	roleTraitOptions := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	resource, err := rs.NewRoleResource(
		name,
		resourceTypeRole,
		id,
		roleTraitOptions,
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r *roleResourceType) List(ctx context.Context, parentId *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeRole.Id})
	if err != nil {
		return nil, "", nil, err
	}

	var allRoles []sentinelone.Role
	switch bag.ResourceTypeID() {
	case resourceTypeRole.Id:
		predefinedRoles, nextCursor, err := r.client.GetPredefinedRoles(ctx, sentinelone.ParamsMap{
			cursor: page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list predefined roles: %w", err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		allRoles = append(allRoles, predefinedRoles...)
		if nextCursor == "" {
			bag.Pop()
			bag.Push(pagination.PageState{
				ResourceTypeID: resourceTypeUser.Id,
			})

			bag.Push(pagination.PageState{
				ResourceTypeID: resourceTypeServiceUser.Id,
			})
		}

	case resourceTypeUser.Id:
		// we have to fetch all users and service users to get the custom roles, as they are not returned by the API
		// this is very costly now but will be fixed in the future with cache
		users, nextCursor, err := r.client.GetUsers(ctx, sentinelone.ParamsMap{
			cursor: page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to get users for custom roles: %w", err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, user := range users {
			if user.ScopeRoles != nil {
				allRoles = append(allRoles, user.ScopeRoles...)
			}
		}

	case resourceTypeServiceUser.Id:
		serviceUsers, nextCursor, err := r.client.GetServiceUsers(ctx, sentinelone.ParamsMap{
			cursor: page,
		})

		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to get service users for custom roles: %w", err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, serviceUser := range serviceUsers {
			if serviceUser.ScopeRoles != nil {
				allRoles = append(allRoles, serviceUser.ScopeRoles...)
			}
		}

	default:
		return nil, "", nil, fmt.Errorf("unexpected resource type while fetching roles")
	}

	var rv []*v2.Resource
	for _, role := range allRoles {
		roleCopy := role
		rr, err := roleResource(ctx, &roleCopy)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	pageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return rv, pageToken, nil, nil
}

func (r *roleResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement
	var assignmentEntitlement *v2.Entitlement
	for _, membership := range memberships {
		assignmentOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(resourceTypeUser, resourceTypeServiceUser),
			ent.WithDisplayName(fmt.Sprintf("%s Role with %s", resource.DisplayName, membership)),
			ent.WithDescription(fmt.Sprintf("%s role in SentinelOne", resource.DisplayName)),
		}
		assignmentEntitlement = ent.NewAssignmentEntitlement(
			resource,
			membership,
			assignmentOptions...,
		)
		rv = append(rv, assignmentEntitlement)
	}

	return rv, "", nil, nil
}

func (r *roleResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeRole.Id})
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	switch bag.ResourceTypeID() {
	case resourceTypeRole.Id:
		bag.Pop()
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeUser.Id,
		})
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeServiceUser.Id,
		})

	case resourceTypeUser.Id:
		roleUsers, nextCursor, err := r.client.GetUsers(ctx, sentinelone.ParamsMap{
			rolesFilter: resource.Id.Resource,
			cursor:      page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list users for role %s: %w", resource.Id.Resource, err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, roleUser := range roleUsers {
			roleUserCopy := roleUser
			ur, err := userResource(&roleUserCopy, resource.Id)
			if err != nil {
				return nil, "", nil, fmt.Errorf("error creating user resource for role %s: %w", resource.Id.Resource, err)
			}
			rv = append(rv, grant.NewGrant(resource, memberships[roleUser.Scope], ur.Id))
		}

	case resourceTypeServiceUser.Id:
		roleServiceUsers, nextCursor, err := r.client.GetServiceUsers(ctx, sentinelone.ParamsMap{
			rolesFilter: resource.Id.Resource,
			cursor:      page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list service users for role %s: %w", resource.Id.Resource, err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, roleServiceUser := range roleServiceUsers {
			roleServiceUserCopy := roleServiceUser
			sur, err := serviceUserResource(&roleServiceUserCopy, resource.Id)
			if err != nil {
				return nil, "", nil, fmt.Errorf("error creating service user resource for role %s: %w", resource.Id.Resource, err)
			}
			rv = append(rv, grant.NewGrant(resource, memberships[roleServiceUser.Scope], sur.Id))
		}

	default:
		return nil, "", nil, fmt.Errorf("unexpected resource type while fetching grants for a role")
	}

	pageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return rv, pageToken, nil, nil
}

func roleBuilder(client *sentinelone.Client) *roleResourceType {
	return &roleResourceType{
		resourceType: resourceTypeRole,
		client:       client,
	}
}
