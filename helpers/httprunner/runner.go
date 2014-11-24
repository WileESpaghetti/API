package httprunner

import (
	"bytes"
	"encoding/json"
	"github.com/curt-labs/GoAPI/controllers/middleware"
	"github.com/curt-labs/GoAPI/helpers/encoding"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/rakyll/pb"
)

type result struct {
	err           error
	statusCode    int
	duration      time.Duration
	contentLength int64
}

type ReqOpts struct {
	Method     string
	URL        string
	Header     http.Header
	Username   string
	Password   string
	Handler    martini.Handler
	Middleware []martini.Handler
	Body       *url.Values
	// OriginalHost represents the original host name user is provided.
	// Request host is an resolved IP. TLS/SSL handshakes may require
	// the original server name, keep it to initate the TLS client.
	OriginalHost string
}

// Creates a req object from req options
func (r *ReqOpts) GenerateRequest() *http.Request {
	var req *http.Request
	if r.Body != nil && strings.ToUpper(r.Method) != "GET" {
		req, _ = http.NewRequest(r.Method, r.URL, bytes.NewBufferString(r.Body.Encode()))
	} else if r.Body != nil {
		req, _ = http.NewRequest(r.Method, r.URL+"?"+r.Body.Encode(), nil)
	} else {
		req, _ = http.NewRequest(r.Method, r.URL, nil)
	}
	req.Header = r.Header

	// update the Host value in the Request - this is used as the host header in any subsequent request
	req.Host = r.OriginalHost

	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}
	return req
}

type Runner struct {
	// Req represents the options of the request to be made.
	// TODO(jbd): Make it work with an http.Request instead.
	Req *ReqOpts

	// N is the total number of requests to make.
	N int

	// C is the concurrency level, the number of concurrent workers to run.
	C int

	// Timeout in seconds.
	Timeout int

	// Qps is the rate limit.
	Qps int

	// AllowInsecure is an option to allow insecure TLS/SSL certificates.
	AllowInsecure bool

	// DisableCompression is an option to disable compression in response
	DisableCompression bool

	// DisableKeepAlives is an option to prevents re-use of TCP connections between different HTTP requests
	DisableKeepAlives bool

	// Output represents the output type. If "csv" is provided, the
	// output will be dumped as a csv stream.
	Output string

	// ProxyAddr is the address of HTTP proxy server in the format on "host:port".
	// Optional.
	ProxyAddr *url.URL

	bar     *pb.ProgressBar
	results chan *result
}

func newPb(size int) (bar *pb.ProgressBar) {
	bar = pb.New(size)
	bar.Format("Bom !")
	bar.Start()
	return
}

func Request(method, route string, body *url.Values, handler martini.Handler) *httptest.ResponseRecorder {
	m := martini.New()
	r := martini.NewRouter()
	switch strings.ToUpper(method) {
	case "GET":
		r.Get(route, handler)
	case "POST":
		r.Post(route, handler)
	case "PUT":
		r.Put(route, handler)
	case "PATCH":
		r.Patch(route, handler)
	case "DELETE":
		r.Delete(route, handler)
	case "HEAD":
		r.Head(route, handler)
	default:
		r.Any(route, handler)
	}
	m.Use(render.Renderer())
	m.Use(encoding.MapEncoder)
	m.Use(middleware.Meddler())
	m.Action(r.Handle)

	var request *http.Request
	if body != nil && strings.ToUpper(method) != "GET" {
		request, _ = http.NewRequest(method, route, bytes.NewBufferString(body.Encode()))
	} else if body != nil {
		request, _ = http.NewRequest(method, route+"?"+body.Encode(), nil)
	} else {
		request, _ = http.NewRequest(method, route, nil)
	}

	response := httptest.NewRecorder()
	m.ServeHTTP(response, request)

	return response
}

func ParamterizedRequest(method, prepared_route string, route string, body *url.Values, handler martini.Handler) *httptest.ResponseRecorder {
	m := martini.New()
	r := martini.NewRouter()
	switch strings.ToUpper(method) {
	case "GET":
		r.Get(prepared_route, handler)
	case "POST":
		r.Post(prepared_route, handler)
	case "PUT":
		r.Put(prepared_route, handler)
	case "PATCH":
		r.Patch(prepared_route, handler)
	case "DELETE":
		r.Delete(prepared_route, handler)
	case "HEAD":
		r.Head(prepared_route, handler)
	default:
		r.Any(prepared_route, handler)
	}
	m.Use(render.Renderer())
	m.Use(encoding.MapEncoder)
	m.Use(middleware.Meddler())
	m.Action(r.Handle)

	var request *http.Request
	if body != nil && strings.ToUpper(method) != "GET" {
		request, _ = http.NewRequest(method, route, bytes.NewBufferString(body.Encode()))
	} else if body != nil {
		request, _ = http.NewRequest(method, route+"?"+body.Encode(), nil)
	} else {
		request, _ = http.NewRequest(method, route, nil)
	}

	response := httptest.NewRecorder()
	m.ServeHTTP(response, request)

	return response
}

func JsonRequest(method, route string, qs *url.Values, iface interface{}, handler martini.Handler) *httptest.ResponseRecorder {
	m := martini.New()
	r := martini.NewRouter()
	switch strings.ToUpper(method) {
	case "GET":
		r.Get(route, handler)
	case "POST":
		r.Post(route, handler)
	case "PUT":
		r.Put(route, handler)
	case "PATCH":
		r.Patch(route, handler)
	case "DELETE":
		r.Delete(route, handler)
	case "HEAD":
		r.Head(route, handler)
	default:
		r.Any(route, handler)
	}

	m.Use(render.Renderer())
	m.Use(encoding.MapEncoder)
	m.Use(middleware.Meddler())
	m.Action(r.Handle)

	js, err := json.Marshal(iface)
	if err != nil {
		return nil
	}

	if qs != nil {
		route = route + "?" + qs.Encode()
	}

	request, _ := http.NewRequest(method, route, bytes.NewBuffer(js))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	m.ServeHTTP(response, request)

	return response
}

func RequestBenchmark(runs int, method, route string, body *url.Values, handler martini.Handler) {

	opts := ReqOpts{
		Body:    body,
		Handler: handler,
		URL:     route,
		Method:  method,
	}

	(&Runner{
		Req: &opts,
		N:   runs,
		C:   1,
	}).Run()
}
