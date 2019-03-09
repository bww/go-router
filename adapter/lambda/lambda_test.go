package lambda

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestAdapter(t *testing.T) {

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
		RequestContext:                  events.APIGatewayProxyRequestContext{},
		Body:                            `{"yo":"my dudes"}`,
		IsBase64Encoded:                 false,
	}

	spew.Dump(req)
	conv, err := ConvertRequest(req)
	spew.Dump(conv)
	if assert.Nil(t, err, fmt.Sprint(err)) {
		assert.Equal(t, req.Path, conv.URL.Path)
		assert.Equal(t, "www.github.com", conv.URL.Host)
		assert.Equal(t, "www.github.com", conv.Host)
		assert.Equal(t, defaultScheme+"://www.github.com"+req.Path+"?X=Y", conv.URL.String())
		assert.Equal(t, "X=Y", conv.URL.RawQuery)
		assert.Equal(t, int64(len(req.Body)), conv.ContentLength)
		body, err := ioutil.ReadAll(conv.Body)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			assert.Equal(t, req.Body, string(body))
		}
	}

}
