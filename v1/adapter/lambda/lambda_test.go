package lambda

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/bww/go-router/v1"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	// "github.com/davecgh/go-spew/spew"
)

func TestAdapter(t *testing.T) {
	entity := `{"yo":"my dudes"}`

	req := events.APIGatewayProxyRequest{
		Resource:                        "/a/b/c",
		Path:                            "/a/b/c",
		HTTPMethod:                      "GET",
		Headers:                         map[string]string{"Host": "www.github.com"},
		MultiValueHeaders:               map[string][]string{"Host": []string{"www.github.com"}},
		QueryStringParameters:           map[string]string{"X": "Y"},
		MultiValueQueryStringParameters: map[string][]string{"X": []string{"Y"}},
		PathParameters:                  map[string]string{},
		StageVariables:                  map[string]string{},
		Body:                            entity,
		IsBase64Encoded:                 false,
		RequestContext: events.APIGatewayProxyRequestContext{
			AccountID:    "001",
			ResourceID:   "002",
			Stage:        "003",
			RequestID:    "004",
			ResourcePath: "/a/b/c",
			HTTPMethod:   "GET",
			APIID:        "005",
			Identity: events.APIGatewayRequestIdentity{
				SourceIP: "1.1.1.1",
			},
		},
	}

	hreq, err := ConvertRequest(req)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, req.Path, hreq.URL.Path)
		assert.Equal(t, "www.github.com", hreq.URL.Host)
		assert.Equal(t, "www.github.com", hreq.Host)
		assert.Equal(t, defaultScheme+"://www.github.com"+req.Path+"?X=Y", hreq.URL.String())
		assert.Equal(t, "X=Y", hreq.URL.RawQuery)
		assert.Equal(t, "1.1.1.1", hreq.RemoteAddr)
		assert.Equal(t, int64(len(req.Body)), hreq.ContentLength)
		body, err := ioutil.ReadAll(hreq.Body)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			assert.Equal(t, req.Body, string(body))
		}
	}

	req = events.APIGatewayProxyRequest{
		Resource:              "/a/b/c",
		Path:                  "/a/b/c",
		HTTPMethod:            "GET",
		Headers:               map[string]string{"Host": "www.github.com"},
		QueryStringParameters: map[string]string{"X": "Y"},
		PathParameters:        map[string]string{},
		StageVariables:        map[string]string{},
		Body:                  entity,
		IsBase64Encoded:       false,
		RequestContext: events.APIGatewayProxyRequestContext{
			AccountID:    "001",
			ResourceID:   "002",
			Stage:        "003",
			RequestID:    "004",
			ResourcePath: "/a/b/c",
			HTTPMethod:   "GET",
			APIID:        "005",
			Identity: events.APIGatewayRequestIdentity{
				SourceIP: "1.1.1.1",
			},
		},
	}

	hreq, err = ConvertRequest(req)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, req.Path, hreq.URL.Path)
		assert.Equal(t, "www.github.com", hreq.URL.Host)
		assert.Equal(t, "www.github.com", hreq.Host)
		assert.Equal(t, defaultScheme+"://www.github.com"+req.Path+"?X=Y", hreq.URL.String())
		assert.Equal(t, "X=Y", hreq.URL.RawQuery)
		assert.Equal(t, "1.1.1.1", hreq.RemoteAddr)
		assert.Equal(t, int64(len(req.Body)), hreq.ContentLength)
		body, err := ioutil.ReadAll(hreq.Body)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			assert.Equal(t, req.Body, string(body))
		}
	}

	hrsp, err := router.NewResponse(http.StatusUnauthorized).SetString("text/plain", entity)
	hrsp.Header = make(http.Header)
	hrsp.Header.Set("Host", "www.github.com")
	rsp, err := ConvertResponse(hrsp)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
		assert.Equal(t, map[string]string{"Host": "www.github.com"}, rsp.Headers)
		assert.Equal(t, map[string][]string{"Host": []string{"www.github.com"}}, rsp.MultiValueHeaders)
		assert.Equal(t, entity, rsp.Body)
		assert.Equal(t, false, rsp.IsBase64Encoded)
	}

	hwriter := NewResponseWriter()
	hwriter.Header().Set("Host", "www.github.com")
	hwriter.WriteHeader(http.StatusForbidden)
	_, err = hwriter.Write([]byte(entity))
	assert.Nil(t, err, fmt.Sprint(err))

	rsp, err = hwriter.ConvertResponse()
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, http.StatusForbidden, rsp.StatusCode)
		assert.Equal(t, map[string]string{"Host": "www.github.com"}, rsp.Headers)
		assert.Equal(t, map[string][]string{"Host": []string{"www.github.com"}}, rsp.MultiValueHeaders)
		assert.Equal(t, entity, rsp.Body)
		assert.Equal(t, false, rsp.IsBase64Encoded)
	}

}
