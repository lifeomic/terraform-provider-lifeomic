package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

// Git attributes are set at build-time (via LDFlags) so that it's accessible
// via the binary at runtime (see Makefile).
var (
	GitCommit string
	GitRef    string
)

const (
	userAgentPrefix = "terraform-provider-lifeomic/"

	defaultAPIVersion = "v1"
	defaultHost       = "api.us.lifeomic.com"

	defaultMaxRetries       = 3
	defaultRetryMaxWaitTime = time.Second

	accountHeader = "LifeOmic-Account"

	AuthTokenEnvVar = "LIFEOMIC_TOKEN"
	HostEnvVar      = "LIFEOMIC_HOST"
	AccountIDEnvVar = "LIFEOMIC_ACCOUNT"
	DebugEnvVar     = "LIFEOMIC_DEBUG"
)

type Interface interface {
	Accounts() AccountService
	Policies() PolicyService
}

// Config configures a PHC Client.
type Config struct {
	APIVersion string
	Host       string

	AccountID string
	AuthToken string
	Header    map[string]string

	MaxRetries       int
	MaxRetryWaitTime time.Duration

	Debug bool

	ServiceName string
	Transport   http.RoundTripper
}

// Client interfaces with the PHC API.
type Client struct {
	config     *Config
	httpClient *resty.Client
	transport  *AuthedTransport

	accounts AccountService
	policies PolicyService
}

type APIError struct {
	Message string `json:"error"`
}

func (e APIError) Error() string { return e.Message }

// New creates a new Client with the given Config.
func New(config Config) *Client {
	if config.AuthToken == "" {
		config.AuthToken = os.Getenv(AuthTokenEnvVar)
	}
	if config.AccountID == "" {
		config.AccountID = os.Getenv(AccountIDEnvVar)
	}
	if config.Host == "" {
		config.Host = defaultStr(os.Getenv(HostEnvVar), defaultHost)
	}
	if config.APIVersion == "" {
		config.APIVersion = defaultAPIVersion
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = defaultMaxRetries
	}
	if config.MaxRetryWaitTime == 0 {
		config.MaxRetryWaitTime = defaultRetryMaxWaitTime
	}
	if config.Header == nil {
		config.Header = map[string]string{}
	}

	if !config.Debug {
		// Treat any malformed value as false.
		config.Debug, _ = strconv.ParseBool(os.Getenv(DebugEnvVar))
	}

	transport := NewAuthedTransport(config.AuthToken, config.AccountID, config.ServiceName, config.Header)
	httpClient := &http.Client{Transport: transport}
	client := &Client{httpClient: resty.NewWithClient(httpClient), config: &config}
	client.transport = transport
	client.policies = &policyService{Client: client}
	client.httpClient.SetDebug(config.Debug)
	client.init()
	return client
}

// Client implements interface.
var _ Interface = &Client{}

// Accounts returns an AccountService client.
func (c *Client) Accounts() AccountService {
	return c.accounts
}

// Policies returns a PolicyService client.
func (c *Client) Policies() PolicyService {
	return c.policies
}

// SetUserAgent sets the UserAgent header on the underlying http client.
func (c *Client) SetUserAgent(userAgent string) {
	c.httpClient.SetHeader("User-Agent", userAgent)
}

// SetAPIVersion updates the baseURL of the underlying http client to use the
// given API version.
func (c *Client) SetAPIVersion(version string) {
	c.config.APIVersion = version
	c.setBaseURL()
}

// SetAuthToken updates the Authorization header on the underlying http client
// to the given value.
func (c *Client) SetAuthToken(token string) {
	c.transport.AuthToken = token
}

// SetAccount updates the client to send a LifeOmic-Account header with the
// given name.
func (c *Client) SetAccount(account string) {
	c.transport.AccountID = account
}

// init is ran to initialize the underlying http client.
func (c *Client) init() {
	c.setBaseURL()

	// Set default headers for all requests.
	c.httpClient.SetHeader("Content-Type", "application/json")
	c.httpClient.SetHeader("Accept", "application/json")
	c.SetUserAgent(fmt.Sprintf("%s%s %s", userAgentPrefix, GitRef, GitCommit))

	// Configure retries.
	c.httpClient.SetRetryCount(c.config.MaxRetries)
	c.httpClient.SetRetryMaxWaitTime(c.config.MaxRetryWaitTime)
}

// Request creates a new HTTP request object.
func (c *Client) Request(ctx context.Context) *resty.Request {
	return c.httpClient.NewRequest().SetContext(ctx).SetError(&APIError{})
}

func (c *Client) setBaseURL() {
	c.httpClient.SetBaseURL(fmt.Sprintf("https://%s/%s", c.config.Host, c.config.APIVersion))
}

func checkResponse(res *resty.Response, err error) (*resty.Response, error) {
	if err != nil {
		// There was an error making the request.
		return res, err
	}

	if res.IsError() {
		if apiErr, ok := res.Error().(*APIError); ok {
			// The API responded with an error.
			return res, apiErr
		}
	}

	return res, nil
}

func defaultStr(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
