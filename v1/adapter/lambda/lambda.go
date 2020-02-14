package lambda

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/bww/go-router/v1"

	"github.com/aws/aws-lambda-go/events"
)

const defaultScheme = "https"

func ConvertRequest(req events.APIGatewayProxyRequest) (*router.Request, error) {
	u, err := url.Parse(defaultScheme + "://" + req.Path)
	if err != nil {
		return nil, err
	}

	query := make(url.Values)
	for k, v := range req.QueryStringParameters { // for some reason both single- and multi-value headers are not consistently present...
		query[k] = []string{v}
	}
	for k, v := range req.MultiValueQueryStringParameters { // so we process them both and prefer the multi-value variant...
		query[k] = v
	}
	u.RawQuery = query.Encode()

	var host string
	header := make(http.Header)
	for k, v := range req.Headers { // for some reason both single- and multi-value headers are not consistently present...
		header[k] = []string{v}
		if strings.EqualFold("Host", k) {
			host = v
		}
	}
	for k, v := range req.MultiValueHeaders { // so we process them both and prefer the multi-value variant...
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
	hreq, err := router.NewRequest(req.HTTPMethod, u.String(), bytes.NewReader(entity))
	if err != nil {
		return nil, err
	}

	hreq.Header = header
	hreq.RemoteAddr = req.RequestContext.Identity.SourceIP
	return hreq, nil
}

func ConvertResponse(rsp *router.Response) (events.APIGatewayProxyResponse, error) {
	entity, err := rsp.ReadEntity()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	header := make(map[string]string)
	for k, v := range rsp.Header {
		if len(v) > 0 {
			header[k] = v[0]
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode:        rsp.Status,
		Headers:           header,
		MultiValueHeaders: rsp.Header,
		Body:              string(entity),
		IsBase64Encoded:   false,
	}, nil
}

type lambdaResponseWriter struct {
	status int
	header http.Header
	entity *bytes.Buffer
}

func NewResponseWriter() *lambdaResponseWriter {
	return &lambdaResponseWriter{}
}

func (w *lambdaResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *lambdaResponseWriter) WriteHeader(s int) {
	if w.status == 0 { // don't "write" headers more than once
		w.status = s
	}
}

func (w *lambdaResponseWriter) Write(b []byte) (int, error) {
	if w.entity == nil {
		w.entity = &bytes.Buffer{}
	}
	return w.entity.Write(b)
}

func (w *lambdaResponseWriter) ConvertResponse() (events.APIGatewayProxyResponse, error) {
	if w.status == 0 {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
	}

	header := make(map[string]string)
	for k, v := range w.header {
		if len(v) > 0 {
			header[k] = v[0]
		}
	}

	var entity string
	if w.entity != nil {
		entity = string(w.entity.Bytes())
	}

	return events.APIGatewayProxyResponse{
		StatusCode:        w.status,
		Headers:           header,
		MultiValueHeaders: w.header,
		Body:              entity,
		IsBase64Encoded:   false,
	}, nil
}
