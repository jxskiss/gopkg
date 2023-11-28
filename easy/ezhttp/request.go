package ezhttp

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"github.com/jxskiss/gopkg/v2/perf/json"
	"github.com/jxskiss/gopkg/v2/zlog"
)

var (
	hdrContentTypeKey = http.CanonicalHeaderKey("Content-Type")
	contentTypeJSON   = "application/json"
	contentTypeXML    = "application/xml"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// Request represents a request and options to send with the Do function.
type Request struct {

	// Req should be a fully prepared http Request to sent, if not nil,
	// the following `Method`, `URL`, `Params`, `JSON`, `XML`, `Form`, `Body`
	// will be ignored.
	//
	// If Req is nil, it will be filled using the following data `Method`,
	// `URL`, `Params`, `JSON`, `XML`, `Form`, `Body` to construct the `http.Request`.
	//
	// When building http body, the priority is JSON > XML > Form > Body.
	Req *http.Request

	// Method specifies the verb for the http request, it's optional,
	// default is "GET".
	Method string

	// URL specifies the url to make the http request.
	// It's required if Req is nil.
	URL string

	// Params specifies optional params to merge with URL, it must be one of
	// the following types:
	// - map[string]string
	// - map[string][]string
	// - map[string]any
	// - url.Values
	Params any

	// JSON specifies optional body data for request which can take body,
	// the content-type will be "application/json", it must be one of
	// the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	// - any (will be marshaled with json.Marshal)
	JSON any

	// XML specifies optional body data for request which can take body,
	// the content-type will be "application/xml", it must be one of
	// the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	// - any (will be marshaled with xml.Marshal)
	XML any

	// Form specifies optional body data for request which can take body,
	// the content-type will be "application/x-www-form-urlencoded",
	// it must be one of the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	// - url.Values (will be encoded and wrapped as io.Reader)
	// - map[string]string (will be converted to url.Values)
	// - map[string][]string (will be converted to url.Values)
	// - map[string]any (will be converted to url.Values)
	Form any

	// Body specifies optional body data for request which can take body,
	// the content-type will be detected from the content (may be incorrect),
	// it should be one of the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	Body any

	// Headers will be copied to the request before sent.
	//
	// If "Content-Type" presents, it will replace the default value
	// set by `JSON`, `XML`, `Form`, or `Body`.
	Headers map[string]string

	// Resp specifies an optional destination to unmarshal the response data.
	// if `Unmarshal` is not provided, the header "Content-Type" will be used to
	// detect XML content, else `json.Unmarshal` will be used.
	Resp any

	// Unmarshal specifies an optional function to unmarshal the response data.
	Unmarshal func([]byte, any) error

	// Context specifies an optional context.Context to use with http request.
	Context context.Context

	// Timeout specifies an optional timeout of the http request, if
	// timeout > 0, the request will be attached an timeout context.Context.
	// `Timeout` takes higher priority than `Context`, it both available, only
	// `Timeout` takes effect.
	Timeout time.Duration

	// Client specifies an optional http.Client to do the request, instead of
	// the default http client.
	Client *http.Client

	// DisableRedirect tells the http client don't follow response redirection.
	DisableRedirect bool

	// DumpRequest makes the http request being logged before sent.
	DumpRequest bool

	// DumpResponse makes the http response being logged after received.
	DumpResponse bool

	// When DumpRequest or DumpResponse is true, or both are true,
	// DumpFunc optionally specifies a function to dump the request and response,
	// by default `zlog.StdLogger.Infof` is used.
	DumpFunc func(format string, args ...any)

	// RaiseForStatus tells Do to report an error if the response
	// status code >= 400. The error will be formatted as "unexpected status: <STATUS>".
	RaiseForStatus bool

	internalData struct {
		BasicAuth struct {
			Username, Password string
		}
	}
}

// SetBasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password
// are not encrypted. It should generally only be used in an HTTPS
// request.
//
// See http.Request.SetBasicAuth for details.
func (p *Request) SetBasicAuth(username, password string) *Request {
	p.internalData.BasicAuth.Username = username
	p.internalData.BasicAuth.Password = password
	return p
}

func (p *Request) buildClient() *http.Client {
	if p.Client == nil && !p.DisableRedirect {
		return http.DefaultClient
	}
	var client = *http.DefaultClient
	if p.Client != nil {
		client = *p.Client
	}
	if p.DisableRedirect {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return &client
}

func (p *Request) prepareRequest(method string) (err error) {
	if p.Req != nil {
		return p.mergeRequest()
	}
	reqURL := p.URL
	if p.Params != nil {
		reqURL, err = mergeQuery(reqURL, p.Params)
		if err != nil {
			return err
		}
	}
	if method == "" {
		method = p.Method
	}
	if method == "" || method == "GET" {
		p.Req, err = http.NewRequest(method, reqURL, nil)
		return err
	}

	body, contentType, err := p.buildBody()
	if err != nil {
		return err
	}
	p.Req, err = http.NewRequest(method, p.URL, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		p.Req.Header.Set(hdrContentTypeKey, contentType)
	}
	p.setHeaders()
	return nil
}

func (p *Request) mergeRequest() (err error) {
	httpReq := p.Req
	if p.Params != nil {
		addQuery := castQueryParams(p.Params).Encode()
		if addQuery != "" {
			if httpReq.URL.RawQuery != "" && !strings.HasSuffix(httpReq.URL.RawQuery, "&") {
				httpReq.URL.RawQuery += "&" + addQuery
			} else {
				httpReq.URL.RawQuery += addQuery
			}
		}
	}
	p.setHeaders()
	return nil
}

func mergeQuery(reqURL string, params any) (string, error) {
	parsed, err := url.Parse(reqURL)
	if err != nil {
		return "", err
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", err
	}
	addQuery := castQueryParams(params)
	for k, values := range addQuery {
		for _, v := range values {
			query.Add(k, v)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func castQueryParams(params any) url.Values {
	switch x := params.(type) {
	case url.Values:
		return x
	case map[string][]string:
		return x
	case map[string]string:
		var values = make(url.Values, len(x))
		for k, v := range x {
			values.Set(k, v)
		}
		return values
	case map[string]any:
		var values = make(url.Values, len(x))
		for k, v := range x {
			switch val := v.(type) {
			case string:
				values.Add(k, val)
			case []string:
				values[k] = val
			default:
				values.Add(k, fmt.Sprint(v))
			}
		}
		return values
	}
	return nil
}

func (p *Request) buildBody() (body io.Reader, contentType string, err error) {
	if p.JSON != nil { // JSON
		body, err = makeHTTPBody(p.JSON, json.Marshal)
		contentType = contentTypeJSON
	} else if p.XML != nil { // XML
		body, err = makeHTTPBody(p.XML, xml.Marshal)
		contentType = contentTypeXML
	} else if p.Form != nil { // urlencoded form
		body, err = makeHTTPBody(p.Form, marshalForm)
		contentType = contentTypeForm
	} else if p.Body != nil { // detect content-type from the body data
		var bodyBuf []byte
		switch data := p.Body.(type) {
		case io.Reader:
			bodyBuf, err = io.ReadAll(data)
			if err != nil {
				return nil, "", err
			}
		case []byte:
			bodyBuf = data
		case string:
			bodyBuf = unsafeheader.StringToBytes(data)
		default:
			err = fmt.Errorf("unsupported body data type: %T", data)
			return nil, "", err
		}
		body = bytes.NewReader(bodyBuf)
		if p.Headers[hdrContentTypeKey] == "" {
			contentType = http.DetectContentType(bodyBuf)
		}
	} // else no body data

	return
}

func marshalForm(v any) ([]byte, error) {
	var form url.Values
	switch data := v.(type) {
	case url.Values:
		form = data
	case map[string][]string:
		form = data
	case map[string]string:
		form = make(url.Values, len(data))
		for k, v := range data {
			form[k] = []string{v}
		}
	case map[string]any:
		form = make(url.Values, len(data))
		for k, v := range data {
			switch value := v.(type) {
			case string:
				form[k] = []string{value}
			case []string:
				form[k] = value
			default:
				form[k] = []string{fmt.Sprint(v)}
			}
		}
	}
	if form == nil {
		err := fmt.Errorf("unsupported form data type: %T", v)
		return nil, err
	}
	encoded := form.Encode()
	buf := unsafeheader.StringToBytes(encoded)
	return buf, nil
}

type marshalFunc func(any) ([]byte, error)

func makeHTTPBody(data any, marshal marshalFunc) (io.Reader, error) {
	var body io.Reader
	switch x := data.(type) {
	case io.Reader:
		body = x
	case []byte:
		body = bytes.NewReader(x)
	case string:
		body = strings.NewReader(x)
	default:
		buf, err := marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(buf)
	}
	return body, nil
}

func (p *Request) setHeaders() {
	for k, v := range p.Headers {
		p.Req.Header.Set(k, v)
	}
	if p.internalData.BasicAuth.Username != "" || p.internalData.BasicAuth.Password != "" {
		p.Req.SetBasicAuth(p.internalData.BasicAuth.Username, p.internalData.BasicAuth.Password)
	}
}

// Do is a convenient function to send request and control redirect
// and debug options. If `Request.Resp` is provided, it will be used as
// destination to try to unmarshal the response body.
//
// Trade-off was taken to balance simplicity and convenience.
//
// For more powerful controls of a http request and handy utilities,
// you may take a look at the awesome library `https://github.com/go-resty/resty/`.
func Do(req *Request) (header http.Header, respContent []byte, status int, err error) {
	err = req.prepareRequest("")
	if err != nil {
		return header, respContent, status, err
	}

	dumpFunc := req.DumpFunc
	if dumpFunc == nil {
		dumpFunc = zlog.StdLogger.Infof
	}

	httpReq := req.Req
	if req.Context != nil {
		httpReq = httpReq.WithContext(req.Context)
	}
	if req.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(httpReq.Context(), req.Timeout)
		defer cancel()
		httpReq = httpReq.WithContext(timeoutCtx)
	}
	if req.DumpRequest {
		var dump []byte
		dump, err = httputil.DumpRequestOut(httpReq, true)
		if err != nil {
			return header, respContent, status, err
		}
		dumpFunc("dump http request:\n%s", dump)
	}

	httpClient := req.buildClient()
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return header, respContent, status, err
	}
	defer httpResp.Body.Close()

	header = httpResp.Header
	status = httpResp.StatusCode
	if req.DumpResponse {
		var dump []byte
		dump, err = httputil.DumpResponse(httpResp, true)
		if err != nil {
			return header, respContent, status, err
		}
		dumpFunc("dump http response:\n%s", dump)
	}

	respContent, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return header, respContent, status, err
	}
	if req.RaiseForStatus {
		if httpResp.StatusCode >= 400 {
			err = fmt.Errorf("unexpected status: %v", httpResp.Status)
			return header, respContent, status, err
		}
	}

	if req.Resp != nil && len(respContent) > 0 {
		unmarshal := req.Unmarshal
		if unmarshal == nil {
			ct := httpResp.Header.Get(hdrContentTypeKey)
			if IsXMLType(ct) {
				unmarshal = xml.Unmarshal
			}
			// default: JSON
			if unmarshal == nil {
				unmarshal = json.Unmarshal
			}
		}
		err = unmarshal(respContent, req.Resp)
		if err != nil {
			return header, respContent, status, err
		}
	}
	return header, respContent, status, err
}
