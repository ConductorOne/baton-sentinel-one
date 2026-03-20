package sentinelone

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"google.golang.org/grpc/codes"
)

type Client struct {
	httpClient *uhttp.BaseHttpClient
	token      string
	baseUrl    string
}

type ParamsMap map[string]string

type PaginationResponse struct {
	Pagination struct {
		TotalItems int    `json:"totalItems"`
		NextCursor string `json:"nextCursor"`
	} `json:"pagination"`
}

type Error struct {
	Code   int    `json:"code"`
	Detail string `json:"detail"`
	Title  string `json:"title"`
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

func grpcCodeFromSentinelOneError(code int) codes.Code {
	// SentinelOne error codes encode the HTTP status in the leading digits
	// (e.g. 4010010 → 401, 4030050 → 403). Dividing by 10000 extracts it.
	switch code / 10000 {
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 429:
		return codes.DeadlineExceeded
	case 500:
		return codes.Internal
	case 503:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}

func (e ErrorResponse) Err(op string) error {
	if len(e.Errors) == 0 {
		return nil
	}
	first := e.Errors[0]
	return uhttp.WrapErrors(
		grpcCodeFromSentinelOneError(first.Code),
		fmt.Sprintf("baton-sentinel-one: %s: %s", op, first.Detail),
		fmt.Errorf("code=%d title=%s", first.Code, first.Title),
	)
}

type Response[T any] struct {
	PaginationResponse
	ErrorResponse
	Data []T `json:"data"`
}

const (
	usersEndpoint        = "users"
	serviceUsersEndpoint = "service-users"
	accountsEndpoint     = "accounts"
	sitesEndpoint        = "sites"
	rolesEndpoint        = "rbac/roles"
)

func NewClient(httpClient *http.Client, baseUrl, token string) *Client {
	return &Client{
		httpClient: uhttp.NewBaseHttpClient(httpClient),
		token:      token,
		baseUrl:    baseUrl,
	}
}

// GetUsers returns a list of all users.
func (c *Client) GetUsers(ctx context.Context, params ParamsMap) ([]User, string, error) {
	var queryParams url.Values
	if params != nil {
		queryParams = createParams(params)
	}

	rawURL, err := url.JoinPath(c.baseUrl, usersEndpoint)
	if err != nil {
		return nil, "", fmt.Errorf("baton-sentinel-one: get users: failed to build URL: %w", err)
	}

	var res Response[User]
	if err := c.doRequest(ctx, rawURL, &res, queryParams); err != nil {
		return nil, "", err
	}
	if err := res.Err("get users"); err != nil {
		return nil, "", err
	}

	if res.Pagination.NextCursor != "" {
		return res.Data, res.Pagination.NextCursor, nil
	}

	return res.Data, "", nil
}

// GetServiceUsers returns a list of all service users.
func (c *Client) GetServiceUsers(ctx context.Context, params ParamsMap) ([]ServiceUser, string, error) {
	var queryParams url.Values
	if params != nil {
		queryParams = createParams(params)
	}

	rawURL, err := url.JoinPath(c.baseUrl, serviceUsersEndpoint)
	if err != nil {
		return nil, "", fmt.Errorf("baton-sentinel-one: get service users: failed to build URL: %w", err)
	}

	var res Response[ServiceUser]
	if err := c.doRequest(ctx, rawURL, &res, queryParams); err != nil {
		return nil, "", err
	}
	if err := res.Err("get service users"); err != nil {
		return nil, "", err
	}

	if res.Pagination.NextCursor != "" {
		return res.Data, res.Pagination.NextCursor, nil
	}

	return res.Data, "", nil
}

// GetAccounts returns a list of all accounts.
func (c *Client) GetAccounts(ctx context.Context, params ParamsMap) ([]Account, string, error) {
	var queryParams url.Values
	if params != nil {
		queryParams = createParams(params)
	}

	rawURL, err := url.JoinPath(c.baseUrl, accountsEndpoint)
	if err != nil {
		return nil, "", fmt.Errorf("baton-sentinel-one: get accounts: failed to build URL: %w", err)
	}

	var res Response[Account]
	if err := c.doRequest(ctx, rawURL, &res, queryParams); err != nil {
		return nil, "", err
	}
	if err := res.Err("get accounts"); err != nil {
		return nil, "", err
	}

	if res.Pagination.NextCursor != "" {
		return res.Data, res.Pagination.NextCursor, nil
	}

	return res.Data, "", nil
}

// GetSites returns a list of all sites.
func (c *Client) GetSites(ctx context.Context, params ParamsMap) ([]Site, string, error) {
	var queryParams url.Values
	if params != nil {
		queryParams = createParams(params)
	}

	var res struct {
		PaginationResponse
		ErrorResponse
		Data struct {
			Sites []Site `json:"sites"`
		} `json:"data"`
	}
	rawURL, err := url.JoinPath(c.baseUrl, sitesEndpoint)
	if err != nil {
		return nil, "", fmt.Errorf("baton-sentinel-one: get sites: failed to build URL: %w", err)
	}
	if err := c.doRequest(ctx, rawURL, &res, queryParams); err != nil {
		return nil, "", err
	}
	if err := res.Err("get sites"); err != nil {
		return nil, "", err
	}

	if res.Pagination.NextCursor != "" {
		return res.Data.Sites, res.Pagination.NextCursor, nil
	}

	return res.Data.Sites, "", nil
}

// GetPredefinedRoles returns a list of all predefined roles.
func (c *Client) GetPredefinedRoles(ctx context.Context, params ParamsMap) ([]Role, string, error) {
	var queryParams url.Values
	if params != nil {
		queryParams = createParams(params)
	}

	rawURL, err := url.JoinPath(c.baseUrl, rolesEndpoint)
	if err != nil {
		return nil, "", fmt.Errorf("baton-sentinel-one: get predefined roles: failed to build URL: %w", err)
	}

	var res Response[Role]
	if err := c.doRequest(ctx, rawURL, &res, queryParams); err != nil {
		return nil, "", err
	}
	if err := res.Err("get predefined roles"); err != nil {
		return nil, "", err
	}

	if res.Pagination.NextCursor != "" {
		return res.Data, res.Pagination.NextCursor, nil
	}

	return res.Data, "", nil
}

func createParams(params ParamsMap) url.Values {
	urlParams := url.Values{}

	for k, v := range params {
		urlParams.Add(k, v)
	}

	// this will speed up the execution time.
	urlParams.Add("skipCount", "true")

	return urlParams
}

func (c *Client) doRequest(ctx context.Context, rawURL string, res interface{}, queryParams url.Values) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("baton-sentinel-one: failed to parse URL: %w", err)
	}
	if queryParams != nil {
		parsedURL.RawQuery = queryParams.Encode()
	}

	req, err := c.httpClient.NewRequest(
		ctx,
		http.MethodGet,
		parsedURL,
		uhttp.WithHeader("Accept", "application/json"),
		uhttp.WithHeader("Authorization", fmt.Sprintf("ApiToken %s", c.token)),
	)
	if err != nil {
		return fmt.Errorf("baton-sentinel-one: failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req, uhttp.WithJSONResponse(res))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("baton-sentinel-one: request failed: %w", err)
	}
	return nil
}
