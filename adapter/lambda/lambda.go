package lambda

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const defaultScheme = "https"

func ConvertRequest(req events.APIGatewayProxyRequest) (*http.Request, error) {
	u, err := url.Parse(defaultScheme + "://" + req.Path)
	if err != nil {
		return nil, err
	}

	query := make(url.Values)
	for k, v := range req.MultiValueQueryStringParameters {
		query[k] = v
	}
	u.RawQuery = query.Encode()

	var host string
	header := make(http.Header)
	for k, v := range req.MultiValueHeaders {
		header[k] = v
		if strings.EqualFold("Host", k) {
			if len(v) > 0 {
				host = v[0]
			}
		}
	}

	var entity []byte
	if req.IsBase64Encoded {
		entity, err = base64.StdEncoding.DecodeString(req.Body)
	} else {
		entity = []byte(req.Body)
	}
	if err != nil {
		return nil, err
	}

	u.Host = host
	conv, err := http.NewRequest(req.HTTPMethod, u.String(), bytes.NewReader(entity))
	if err != nil {
		return nil, err
	}

	conv.Header = header
	conv.RemoteAddr = req.RequestContext.Identity.SourceIP
	// conv.ContentLength = len(req.Body)
	// conv.Host = host

	return conv, nil
}

func WriteResponse(w http.ResponseWriter, rsp events.APIGatewayProxyResponse) error {
	return fmt.Errorf("Unimplemented")
}
