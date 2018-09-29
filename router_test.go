package router

import (
  "fmt"
  "testing"
  
  "github.com/stretchr/testify/assert"
  "github.com/aws/aws-lambda-go/events"
)

func TestPaths(t *testing.T) {
  var m bool
  var v map[string]string
  
  m, _ = parsePath("/").Matches("/")
  assert.Equal(t, true, m)
  m, _ = parsePath("/a").Matches("/a")
  assert.Equal(t, true, m)
  m, _ = parsePath("/a/b").Matches("/a/b")
  assert.Equal(t, true, m)
  
  m, v = parsePath("/a/{var}").Matches("/a/b")
  assert.Equal(t, true, m)
  assert.Equal(t, map[string]string{"var":"b"}, v)
  m, v = parsePath("/a/{var}/c").Matches("/a/b/c")
  assert.Equal(t, true, m)
  assert.Equal(t, map[string]string{"var":"b"}, v)
  m, v = parsePath("/a/{v/r}/c").Matches("/a/b/c")
  assert.Equal(t, true, m)
  assert.Equal(t, map[string]string{"v/r":"b"}, v)
  m, v = parsePath("/a/{var1}/c/{var2}").Matches("/a/b/c/d")
  assert.Equal(t, true, m)
  assert.Equal(t, map[string]string{"var1":"b", "var2":"d"}, v)
  
  m, _ = parsePath("/").Matches("/a")
  assert.Equal(t, false, m)
  m, _ = parsePath("/a").Matches("/")
  assert.Equal(t, false, m)
  m, _ = parsePath("/a/b").Matches("/a/c")
  assert.Equal(t, false, m)
  m, _ = parsePath("/a/{var}").Matches("/a/b/c")
  assert.Equal(t, false, m)
  m, _ = parsePath("/a/c/{var}").Matches("/a/b/c")
  assert.Equal(t, false, m)
  m, _ = parsePath("/a/{var1}/{var2}").Matches("/x/b/c")
  assert.Equal(t, false, m)
  m, _ = parsePath("/a/{var1}/{var2}").Matches("/a/b/c/d")
  assert.Equal(t, false, m)
  
}

func TestRoutes(t *testing.T) {
  var x *Route
  var v map[string]string
  var err error
  
  funcA := func (events.APIGatewayProxyRequest, Context) (events.APIGatewayProxyResponse, error) { return events.APIGatewayProxyResponse{Body:"A"}, nil }
  funcB := func (events.APIGatewayProxyRequest, Context) (events.APIGatewayProxyResponse, error) { return events.APIGatewayProxyResponse{Body:"B"}, nil }
  funcC := func (events.APIGatewayProxyRequest, Context) (events.APIGatewayProxyResponse, error) { return events.APIGatewayProxyResponse{Body:"C"}, nil }
  funcD := func (events.APIGatewayProxyRequest, Context) (events.APIGatewayProxyResponse, error) { return events.APIGatewayProxyResponse{Body:"D"}, nil }
  funcE := func (events.APIGatewayProxyRequest, Context) (events.APIGatewayProxyResponse, error) { return events.APIGatewayProxyResponse{Body:"E"}, nil }
  
  r := &Router{}
  r.Add("/a", funcA).Methods("GET")
  r.Add("/a", funcB).Methods("PUT")
  r.Add("/a", funcC)
  
  r.Add("/b",     funcD)
  r.Add("/{var}", funcE)
  
  x, v, err = r.Find(events.APIGatewayProxyRequest{Path:"/a", HTTPMethod:"GET"})
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    if assert.NotNil(t, x) {
      r, _ := x.handler(events.APIGatewayProxyRequest{}, Context{})
      assert.Equal(t, "A", r.Body)
      assert.Equal(t, map[string]string(nil), v)
    }
  }
  
  x, v, err = r.Find(events.APIGatewayProxyRequest{Path:"/a", HTTPMethod:"PUT"})
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    if assert.NotNil(t, x) {
      r, _ := x.handler(events.APIGatewayProxyRequest{}, Context{})
      assert.Equal(t, "B", r.Body)
      assert.Equal(t, map[string]string(nil), v)
    }
  }
  
  x, v, err = r.Find(events.APIGatewayProxyRequest{Path:"/a", HTTPMethod:"ANYTHING"})
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    if assert.NotNil(t, x) {
      r, _ := x.handler(events.APIGatewayProxyRequest{}, Context{})
      assert.Equal(t, "C", r.Body)
      assert.Equal(t, map[string]string(nil), v)
    }
  }
  
  x, v, err = r.Find(events.APIGatewayProxyRequest{Path:"/b", HTTPMethod:"GET"})
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    if assert.NotNil(t, x) {
      r, _ := x.handler(events.APIGatewayProxyRequest{}, Context{})
      assert.Equal(t, "D", r.Body)
      assert.Equal(t, map[string]string(nil), v)
    }
  }
  
  x, v, err = r.Find(events.APIGatewayProxyRequest{Path:"/c", HTTPMethod:"GET"})
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    if assert.NotNil(t, x) {
      r, _ := x.handler(events.APIGatewayProxyRequest{}, Context{})
      assert.Equal(t, "E", r.Body)
      assert.Equal(t, map[string]string{"var":"c"}, v)
    }
  }
  
}
