package sentinelone

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	httpClient *http.Client
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
		httpClient: httpClient,
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

	var res Response[User]
	if err := c.doRequest(ctx, fmt.Sprint(c.baseUrl, usersEndpoint), &res, queryParams); err != nil {
		return nil, "", err
	}

	if res.ErrorResponse.Errors != nil {
		return nil, "", fmt.Errorf("failed to get users: %v", res.ErrorResponse.Errors)
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

	var res Response[ServiceUser]
	if err := c.doRequest(ctx, fmt.Sprint(c.baseUrl, serviceUsersEndpoint), &res, queryParams); err != nil {
		return nil, "", err
	}

	if res.ErrorResponse.Errors != nil {
		return nil, "", fmt.Errorf("failed to get service users: %v", res.ErrorResponse.Errors)
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

	var res Response[Account]
	if err := c.doRequest(ctx, fmt.Sprint(c.baseUrl, accountsEndpoint), &res, queryParams); err != nil {
		return nil, "", err
	}

	if res.ErrorResponse.Errors != nil {
		return nil, "", fmt.Errorf("failed to get accounts: %v", res.ErrorResponse.Errors)
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
	if err := c.doRequest(ctx, fmt.Sprint(c.baseUrl, sitesEndpoint), &res, queryParams); err != nil {
		return nil, "", err
	}

	if res.ErrorResponse.Errors != nil {
		return nil, "", fmt.Errorf("failed to get sites: %v", res.ErrorResponse.Errors)
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

	var res Response[Role]
	if err := c.doRequest(ctx, fmt.Sprint(c.baseUrl, rolesEndpoint), &res, queryParams); err != nil {
		return nil, "", err
	}

	if res.ErrorResponse.Errors != nil {
		return nil, "", fmt.Errorf("failed to get roles: %v", res.ErrorResponse.Errors)
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

func (c *Client) doRequest(ctx context.Context, url string, res interface{}, queryParams url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if queryParams != nil {
		req.URL.RawQuery = queryParams.Encode()
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("ApiToken %s", c.token))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	return nil
}
