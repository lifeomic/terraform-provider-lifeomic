package lambda

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

const Proto = "lambda://"

type URI struct {
	Function  string
	Qualifier string
}

type Client interface {
	Invoke(context.Context, *lambda.InvokeInput, ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

// RoundTripper implements http.RoundTripper to intercept an outgoing http
// request and invoke the underlying service's lambda rather than going through
// apigateway.
type RoundTripper struct {
	uri    *URI
	client Client
	header map[string]string
}

// payloadFromRequest converts the given http.Request into a payload to be
// interpreted by the underlying service's lambda and marshals it to a byte
// slice.
func payloadFromRequest(req *http.Request, additionalHeader map[string]string) ([]byte, error) {
	header := make(map[string]string, len(req.Header))
	for key := range req.Header {
		header[key] = req.Header.Get(key)
	}

	for k, v := range additionalHeader {
		header[k] = v
	}

	payload := map[string]any{
		"headers":    header,
		"httpMethod": req.Method,
		"path":       req.URL.Path,
	}

	if req.Body != nil {
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, req.Body); err != nil {
			return nil, err
		}
		defer req.Body.Close()

		payload["body"] = buf.String()
	}
	return json.Marshal(payload)
}

// responseFromOutput creates an http.Response from a lambda.InvokeOutput to
// satisfy the http transport.
func responseFromOutput(output *lambda.InvokeOutput) (*http.Response, error) {
	respPayload := new(response)
	if err := json.Unmarshal(output.Payload, respPayload); err != nil {
		return nil, err
	}

	header := make(http.Header, len(respPayload.Headers))
	for key, value := range respPayload.Headers {
		header.Set(key, value)
	}

	return &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(respPayload.Body)),
		Header:     header,
		StatusCode: respPayload.StatusCode,
	}, nil
}

type response struct {
	Body       string
	StatusCode int
	Headers    map[string]string
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	payload, err := payloadFromRequest(req, r.header)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal lambda payload: %w", err)
	}

	output, err := r.client.Invoke(req.Context(), &lambda.InvokeInput{
		FunctionName: &r.uri.Function,
		Qualifier:    &r.uri.Qualifier,
		Payload:      payload,
	})
	if err != nil {
		return nil, err
	}

	return responseFromOutput(output)
}

func (r *RoundTripper) Do(req *http.Request) (*http.Response, error) {
	return r.RoundTrip(req)
}

func NewRoundTripper(ctx context.Context, uri URI, header map[string]string) (*RoundTripper, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	if uri.Qualifier == "" {
		uri.Qualifier = "deployed"
	}

	lambdaClient := lambda.NewFromConfig(cfg)
	return &RoundTripper{
		client: lambdaClient,
		uri:    &uri,
		header: header,
	}, nil
}
