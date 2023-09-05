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

type siteResourceType struct {
	resourceType *v2.ResourceType
	client       *sentinelone.Client
}

func (s *siteResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return s.resourceType
}

const siteMembership = "member"

// Create a new connector resource for an SentinelOne site.
func siteResource(site *sentinelone.Site, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	resource, err := rs.NewResource(
		site.Name,
		resourceTypeSite,
		site.ID,
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *siteResourceType) List(ctx context.Context, parentId *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentId == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeSite.Id})
	if err != nil {
		return nil, "", nil, err
	}

	sites, nextCursor, err := s.client.GetSites(ctx, sentinelone.ParamsMap{
		cursor: page,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list sites: %w", err)
	}

	pageToken, err := bag.NextToken(nextCursor)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Resource
	for _, site := range sites {
		siteCopy := site
		sr, err := siteResource(&siteCopy, parentId)

		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, sr)
	}

	return rv, pageToken, nil, nil
}

func (s *siteResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser, resourceTypeServiceUser),
		ent.WithDisplayName(fmt.Sprintf("%s Site %s", resource.DisplayName, siteMembership)),
		ent.WithDescription(fmt.Sprintf("Access to %s site in SentinelOne", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		siteMembership,
		assignmentOptions...,
	))

	return rv, "", nil, nil
}

func (s *siteResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeSite.Id})
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	switch bag.ResourceTypeID() {
	case resourceTypeSite.Id:
		bag.Pop()
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeUser.Id,
		})
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeServiceUser.Id,
		})

	case resourceTypeUser.Id:
		siteUsers, nextCursor, err := s.client.GetUsers(ctx, sentinelone.ParamsMap{
			sitesFilter: resource.Id.Resource,
			cursor:      page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list users for site %s: %w", resource.Id.Resource, err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, siteUser := range siteUsers {
			siteUserCopy := siteUser
			ur, err := userResource(&siteUserCopy, resource.Id)
			if err != nil {
				return nil, "", nil, fmt.Errorf("error creating user resource for site %s: %w", resource.Id.Resource, err)
			}

			gr := grant.NewGrant(resource, siteMembership, ur.Id)
			rv = append(rv, gr)
		}

	case resourceTypeServiceUser.Id:
		siteServiceUsers, nextCursor, err := s.client.GetServiceUsers(ctx, sentinelone.ParamsMap{
			sitesFilter: resource.Id.Resource,
			cursor:      page,
		})
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to list service users for site %s: %w", resource.Id.Resource, err)
		}

		paginationErr := bag.Next(nextCursor)
		if paginationErr != nil {
			return nil, "", nil, paginationErr
		}

		for _, siteServiceUser := range siteServiceUsers {
			siteServiceUserCopy := siteServiceUser
			ur, err := serviceUserResource(&siteServiceUserCopy, resource.Id)
			if err != nil {
				return nil, "", nil, fmt.Errorf("error creating service user resource for site %s: %w", resource.Id.Resource, err)
			}

			gr := grant.NewGrant(resource, siteMembership, ur.Id)
			rv = append(rv, gr)
		}

	default:
		return nil, "", nil, fmt.Errorf("unexpected resource type while fetching grants for a site")
	}

	pageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return rv, pageToken, nil, nil
}

func siteBuilder(client *sentinelone.Client) *siteResourceType {
	return &siteResourceType{
		resourceType: resourceTypeSite,
		client:       client,
	}
}
