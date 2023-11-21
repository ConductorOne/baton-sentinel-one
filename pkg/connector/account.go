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

const accountMembership = "member"

type accountResourceType struct {
	resourceType *v2.ResourceType
	client       *sentinelone.Client
}

func (a *accountResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return a.resourceType
}

// Create a new connector resource for a SentinelOne account.
func accountResource(account *sentinelone.Account) (*v2.Resource, error) {
	ret, err := rs.NewResource(
		account.Name,
		resourceTypeAccount,
		account.ID,
		rs.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: resourceTypeUser.Id},
			&v2.ChildResourceType{ResourceTypeId: resourceTypeServiceUser.Id},
			&v2.ChildResourceType{ResourceTypeId: resourceTypeSite.Id},
		),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *accountResourceType) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeAccount.Id})
	if err != nil {
		return nil, "", nil, err
	}

	accounts, nextPage, err := a.client.GetAccounts(ctx, sentinelone.ParamsMap{
		cursor: page,
	})
	if err != nil {
		return nil, "", nil, err
	}

	pageToken, err := bag.NextToken(nextPage)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Resource
	for _, account := range accounts {
		accountCopy := account
		ur, err := accountResource(&accountCopy)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, ur)
	}

	return rv, pageToken, nil, nil
}

func (a *accountResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser, resourceTypeSite, resourceTypeServiceUser),
		ent.WithDisplayName(fmt.Sprintf("%s Account %s", resource.DisplayName, accountMembership)),
		ent.WithDescription(fmt.Sprintf("Access to %s account in SentinelOne", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		accountMembership,
		assignmentOptions...,
	))

	return rv, "", nil, nil
}

func (a *accountResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeAccount.Id})
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	switch bag.ResourceTypeID() {
	case resourceTypeAccount.Id:
		bag.Pop()
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeUser.Id,
		})
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeServiceUser.Id,
		})
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeSite.Id,
		})

	case resourceTypeUser.Id:
		accountUsers, nextCursor, err := a.client.GetUsers(ctx, sentinelone.ParamsMap{
			accountsFilter: resource.Id.Resource,
			cursor:         page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list users for account %s: %w", resource.Id.Resource, err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, accountUser := range accountUsers {
			accountUserCopy := accountUser
			ur, err := userResource(&accountUserCopy, resource.Id)
			if err != nil {
				return nil, "", nil, fmt.Errorf("error creating user resource for account %s: %w", resource.Id.Resource, err)
			}
			rv = append(rv, grant.NewGrant(resource, accountMembership, ur.Id))
		}

	case resourceTypeServiceUser.Id:
		accountServiceUsers, nextCursor, err := a.client.GetServiceUsers(ctx, sentinelone.ParamsMap{
			accountsFilter: resource.Id.Resource,
			cursor:         page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list service users for account %s: %w", resource.Id.Resource, err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, accountServiceUser := range accountServiceUsers {
			accountServiceUserCopy := accountServiceUser
			sur, err := serviceUserResource(&accountServiceUserCopy, resource.Id)
			if err != nil {
				return nil, "", nil, fmt.Errorf("error creating service user resource for account %s: %w", resource.Id.Resource, err)
			}

			rv = append(rv, grant.NewGrant(resource, accountMembership, sur.Id))
		}

	case resourceTypeSite.Id:
		accountSites, nextCursor, err := a.client.GetSites(ctx, sentinelone.ParamsMap{
			accountsFilter: resource.Id.Resource,
			cursor:         page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list sites for account %s: %w", resource.Id.Resource, err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, accountSite := range accountSites {
			accountSiteCopy := accountSite
			sr, err := siteResource(&accountSiteCopy, resource.Id)
			if err != nil {
				return nil, "", nil, fmt.Errorf("error creating site resource for account %s: %w", resource.Id.Resource, err)
			}
			rv = append(rv, grant.NewGrant(resource, accountMembership, sr.Id))
		}
	default:
		return nil, "", nil, fmt.Errorf("unexpected resource type while fetching grants for an account")
	}

	pageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return rv, pageToken, nil, nil
}

func accountBuilder(client *sentinelone.Client) *accountResourceType {
	return &accountResourceType{
		resourceType: resourceTypeAccount,
		client:       client,
	}
}
