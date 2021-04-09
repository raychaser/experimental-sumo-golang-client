package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	userAgent = "go-sumologic"
)

// A Client manages communication with the Sumo Logic API.
type Client struct {
	clientMu sync.Mutex   // clientMu protects the client during calls that modify the CheckRedirect func.
	client   *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. Defaults to the public GitHub API, but can be
	// set to a domain endpoint to use with GitHub Enterprise. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the GitHub API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the Sumo Logic API.
	Users           *UsersService
	SearchJob       *SearchJobService
	MetricsQuery    *MetricsQueryService
	DashboardSearch *DashboardSearchService
}

type service struct {
	client *Client
}

// NewClient returns a new Sumo Logic API client. If a nil httpClient is
// provided, a new http.Client will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewClient(endpoint string, httpClient *http.Client) *Client {

	if httpClient == nil {
		httpClient = &http.Client{}
	}
	baseURL, _ := url.Parse(endpoint)

	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent}
	c.common.client = c
	c.Users = (*UsersService)(&c.common)
	c.SearchJob = (*SearchJobService)(&c.common)
	c.MetricsQuery = (*MetricsQueryService)(&c.common)
	c.DashboardSearch = (*DashboardSearchService)(&c.common)
	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Response is a Sumo Logic API response. This wraps the standard http.Response
// returned from Sumo Logic and provides convenient access to things like
// pagination links.
type Response struct {
	*http.Response

	//// These fields provide the page values for paginating through a set of
	//// results. Any or all of these may be set to the zero value for
	//// responses that are not part of a paginated set, or for which there
	//// are no additional pages.
	////
	//// These fields support what is called "offset pagination" and should
	//// be used with the ListOptions struct.
	//NextPage  int
	//PrevPage  int
	//FirstPage int
	//LastPage  int

	//// Additionally, some APIs support "cursor pagination" instead of offset.
	//// This means that a token points directly to the next record which
	//// can lead to O(1) performance compared to O(n) performance provided
	//// by offset pagination.
	////
	//// For APIs that support cursor pagination (such as
	//// TeamsService.ListIDPGroupsInOrganization), the following field
	//// will be populated to point to the next page.
	////
	//// To use this token, set ListCursorOptions.Page to this value before
	//// calling the endpoint again.
	//NextPageToken string
	//
	//// Explicitly specify the Rate type so Rate's String() receiver doesn't
	//// propagate to Response.
	//Rate Rate
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	//response.populatePageValues()
	//response.Rate = parseRate(r)
	return response
}

func withContext(ctx context.Context, req *http.Request) *http.Request {
	return req.WithContext(ctx)
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it is canceled or times out,
// ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}
	req = withContext(ctx, req)

	//rateLimitCategory := category(req.URL.Path)
	//
	//// If we've hit rate limit, don't make further requests before Reset time.
	//if err := c.checkRateLimitBeforeDo(req, rateLimitCategory); err != nil {
	//	return &Response{
	//		Response: err.Response,
	//		Rate:     err.Rate,
	//	}, err
	//}

	log.Debug().Str("method", req.Method).Str("url", req.URL.String()).Msg("Starting request")
	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			log.Debug().Str("method", req.Method).Str("url", req.URL.String()).Err(ctx.Err())
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		//if e, ok := err.(*url.Error); ok {
		//	if url, err := url.Parse(e.URL); err == nil {
		//		e.URL = sanitizeURL(url).String()
		//		return nil, e
		//	}
		//}

		log.Debug().Str("method", req.Method).Str("url", req.URL.String()).Err(err)
		return nil, err
	}
	defer resp.Body.Close()

	response := newResponse(resp)
	log.Debug().Str("method", req.Method).Str("url", req.URL.String()).
		Str("status", response.Status).Msg("Received response")

	//c.rateMu.Lock()
	//c.rateLimits[rateLimitCategory] = response.Rate
	//c.rateMu.Unlock()

	err = CheckResponse(resp)
	if err != nil {
		log.Debug().Str("method", req.Method).Str("url", req.URL.String()).
			Str("status", response.Status).Err(err)
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return response, err
}

// An ErrorResponse reports one or more errors caused by an API request.
// GitHub API docs: https://developer.github.com/v3/#client-errors
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
	//Errors   []Error        `json:"errors"`  // more detail on individual errors
	// Block is only populated on certain types of errors such as code 451.
	// See https://developer.github.com/changes/2016-03-17-the-451-status-code-is-now-supported/
	// for more information.
	Block *struct {
		Reason string `json:"reason,omitempty"`
		//CreatedAt *Timestamp `json:"created_at,omitempty"`
	} `json:"block,omitempty"`
	// Most errors will also include a documentation_url field pointing
	// to some content that might help you resolve the error, see
	// https://developer.github.com/v3/#client-errors
	DocumentationURL string `json:"documentation_url,omitempty"`
}

func (r *ErrorResponse) Error() string {
	//return fmt.Sprintf("%v %v: %d %v %+v",
	return fmt.Sprintf("%v: %d %v",
		//r.Response.Request.Method, sanitizeURL(r.Response.Request.URL),
		r.Response.Request.Method,
		r.Response.StatusCode, r.Message /*, r.Errors*/)
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range or equal to 202 Accepted.
// API error responses are expected to have response
// body, and a JSON response body that maps to ErrorResponse.
//
// The error type will be *RateLimitError for rate limit exceeded errors,
// *AcceptedError for 202 Accepted status codes,
// and *TwoFactorAuthError for two-factor authentication errors.
func CheckResponse(r *http.Response) error {
	//if r.StatusCode == http.StatusAccepted {
	//	return &AcceptedError{}
	//}
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	// Re-populate error response body because GitHub error responses are often
	// undocumented and inconsistent.
	// Issue #1136, #540.
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	switch {
	//case r.StatusCode == http.StatusUnauthorized && strings.HasPrefix(r.Header.Get(headerOTP), "required"):
	//	return (*TwoFactorAuthError)(errorResponse)
	//case r.StatusCode == http.StatusForbidden && r.Header.Get(headerRateRemaining) == "0":
	//	return &RateLimitError{
	//		Rate:     parseRate(r),
	//		Response: errorResponse.Response,
	//		Message:  errorResponse.Message,
	//	}
	//case r.StatusCode == http.StatusForbidden && strings.HasSuffix(errorResponse.DocumentationURL, "/v3/#abuse-rate-limits"):
	//	abuseRateLimitError := &AbuseRateLimitError{
	//		Response: errorResponse.Response,
	//		Message:  errorResponse.Message,
	//	}
	//	if v := r.Header["Retry-After"]; len(v) > 0 {
	//		// According to GitHub support, the "Retry-After" header value will be
	//		// an integer which represents the number of seconds that one should
	//		// wait before resuming making requests.
	//		retryAfterSeconds, _ := strconv.ParseInt(v[0], 10, 64) // Error handling is noop.
	//		retryAfter := time.Duration(retryAfterSeconds) * time.Second
	//		abuseRateLimitError.RetryAfter = &retryAfter
	//	}
	//	return abuseRateLimitError
	default:
		return errorResponse
	}
}

func setCredentialsAsHeaders(req *http.Request, id, secret string) *http.Request {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	convertedRequest := new(http.Request)
	*convertedRequest = *req
	convertedRequest.Header = make(http.Header, len(req.Header))

	for k, s := range req.Header {
		convertedRequest.Header[k] = append([]string(nil), s...)
	}
	convertedRequest.SetBasicAuth(id, secret)
	return convertedRequest
}

// BasicAuthTransport is an http.RoundTripper that authenticates all requests
// using HTTP Basic Authentication with the provided username and password. It
// additionally supports users who have two-factor authentication enabled on
// their GitHub account.
type BasicAuthTransport struct {
	Endpoint string
	Username string // Username
	Password string // Password

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := setCredentialsAsHeaders(req, t.Username, t.Password)
	return t.transport().RoundTrip(req2)
}

// Client returns an *http.Client that makes requests that are authenticated
// using HTTP Basic Authentication.
func (t *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *BasicAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}
