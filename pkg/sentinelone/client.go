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

func (e *ErrorResponse) apiErrors() []Error {
	return e.Errors
}

type apiErrorChecker interface {
	apiErrors() []Error
}

// handleJSONErrors checks for API-level errors returned in a HTTP 200 response body.
// SentinelOne sometimes returns errors as a JSON array rather than using HTTP error status codes.
func handleJSONErrors(res interface{}) error {
	checker, ok := res.(apiErrorChecker)
	if !ok {
		return nil
	}
	// Only the first error is returned; additional errors are noted in the message.
	errs := checker.apiErrors()
	if len(errs) == 0 {
		return nil
	}
	first := errs[0]
	msg := fmt.Sprintf("baton-sentinel-one: %s", first.Detail)
	if extra := len(errs) - 1; extra > 0 {
		msg = fmt.Sprintf("%s (+%d more errors)", msg, extra)
	}
	return uhttp.WrapErrors(
		grpcCodeFromSentinelOneError(first.Code),
		msg,
		fmt.Errorf("code=%d title=%s", first.Code, first.Title),
	)
}

func grpcCodeFromSentinelOneError(code int) codes.Code {
	// SentinelOne error codes encode the HTTP status in the leading digits
	// (e.g. 4010010 → 401, 4030050 → 403). Dividing by 10000 extracts it.
	// The full error code list is only in the authenticated management console Swagger UI
	// at https://<console>/web/api/v2.1 — there is no public page that documents this scheme.
	// Third-party references that show real observed codes:
	//   - LogicHub (4030010 example): https://help.logichub.com/docs/sentinelone
	//   - XSOAR community (4010010 example): https://live.paloaltonetworks.com/t5/cortex-xsoar-discussions/problems-with-sentinelone-v2-integration-401/td-p/515694
	//   - vradchenko PowerShell module (4000040 example): https://github.com/vradchenko/PowerShell-SentinelOne/blob/main/SentinelOne.ps1#L852
	// Known codes: 4010010 = Authentication Failed, 4030010 = Insufficient permissions.
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

// doRequest performs a GET request and decodes the JSON body into res.
//
// Error handling has two paths depending on what the API returns:
//   - Non-200 HTTP status (e.g. 401, 403): uhttp intercepts it and returns a gRPC error
//     derived from the HTTP status code.
//   - HTTP 200 with errors in the JSON body: uhttp returns no error, so handleJSONErrors
//     checks the decoded response for API-level errors and maps them to gRPC codes.
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
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithHeader("Authorization", fmt.Sprintf("ApiToken %s", c.token)),
	)
	if err != nil {
		return fmt.Errorf("baton-sentinel-one: failed to build request: %w", err)
	}

	rawResp, err := c.httpClient.Do(req, uhttp.WithJSONResponse(res))
	if err != nil {
		return fmt.Errorf("baton-sentinel-one: request failed: %w", err)
	}
	defer rawResp.Body.Close()

	if err := handleJSONErrors(res); err != nil {
		return err
	}

	return nil
}
