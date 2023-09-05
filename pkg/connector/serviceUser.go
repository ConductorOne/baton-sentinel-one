package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sentinel-one/pkg/sentinelone"
)

type serviceUserResourceType struct {
	resourceType *v2.ResourceType
	client       *sentinelone.Client
}

func (s *serviceUserResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return s.resourceType
}

// Create a new connector resource for a SentinelOne service user.
func serviceUserResource(serviceUser *sentinelone.ServiceUser, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	firstName, lastName := splitFullName(serviceUser.Name)

	profile := map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
		"user_id":    serviceUser.ID,
	}

	userTraitOptions := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithStatus(v2.UserTrait_Status_STATUS_UNSPECIFIED),
		rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_SERVICE),
	}

	ret, err := rs.NewUserResource(
		serviceUser.Name,
		resourceTypeServiceUser,
		serviceUser.ID,
		userTraitOptions,
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *serviceUserResourceType) List(ctx context.Context, parentId *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentId == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: resourceTypeServiceUser.Id})
	if err != nil {
		return nil, "", nil, err
	}

	users, nextCursor, err := s.client.GetServiceUsers(ctx, sentinelone.ParamsMap{
		cursor: page,
	})
	if err != nil {
		return nil, "", nil, err
	}

	pageToken, err := bag.NextToken(nextCursor)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Resource
	for _, serviceUser := range users {
		serviceUserCopy := serviceUser
		sur, err := serviceUserResource(&serviceUserCopy, parentId)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, sur)
	}

	return rv, pageToken, nil, nil
}

func (s *serviceUserResourceType) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (s *serviceUserResourceType) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func serviceUserBuilder(client *sentinelone.Client) *serviceUserResourceType {
	return &serviceUserResourceType{
		resourceType: resourceTypeServiceUser,
		client:       client,
	}
}
