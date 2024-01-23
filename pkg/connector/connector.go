package connector

import (
	"context"
	"fmt"
	"path"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/conductorone/baton-sentinel-one/pkg/sentinelone"
)

type SentinelOne struct {
	client *sentinelone.Client
}

var (
	resourceTypeAccount = &v2.ResourceType{
		Id:          "account",
		DisplayName: "Account",
	}
	resourceTypeUser = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_USER,
		},
		Annotations: annotationsForUserResourceType(),
	}
	resourceTypeServiceUser = &v2.ResourceType{
		Id:          "service_user",
		DisplayName: "Service User",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_USER,
		},
		Annotations: annotationsForUserResourceType(),
	}
	resourceTypeRole = &v2.ResourceType{
		Id:          "role",
		DisplayName: "Role",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_ROLE,
		},
	}
	resourceTypeSite = &v2.ResourceType{
		Id:          "site",
		DisplayName: "Site",
	}
)

func (s *SentinelOne) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		accountBuilder(s.client),
		userBuilder(s.client),
		serviceUserBuilder(s.client),
		roleBuilder(s.client),
		siteBuilder(s.client),
	}
}

func (s *SentinelOne) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "SentinelOne",
		Description: "Connector syncing SentinelOne accounts, users, service users, roles and sites to Baton.",
	}, nil
}

// Validates that the user has access to all relevant resources.
// It's not defined which role is needed to fetch all resources so we need to check that user has access to all of them.
func (s *SentinelOne) Validate(ctx context.Context) (annotations.Annotations, error) {
	_, _, err := s.client.GetAccounts(ctx, sentinelone.ParamsMap{
		"limit": "1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	_, _, err = s.client.GetSites(ctx, sentinelone.ParamsMap{
		"limit": "1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sites: %w", err)
	}
	_, _, err = s.client.GetUsers(ctx, sentinelone.ParamsMap{
		"limit": "1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	_, _, err = s.client.GetServiceUsers(ctx, sentinelone.ParamsMap{
		"limit": "1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service users: %w", err)
	}

	_, _, err = s.client.GetPredefinedRoles(ctx, sentinelone.ParamsMap{
		"limit": "1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return nil, nil
}

// New returns the SentinelOne connector.
func New(ctx context.Context, baseUrl, token string) (*SentinelOne, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	client := sentinelone.NewClient(httpClient, path.Join(baseUrl, "web/api/v2.1/"), token)

	return &SentinelOne{
		client: client,
	}, nil
}
