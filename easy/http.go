package easy

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/jxskiss/gopkg/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	hdrContentTypeKey = http.CanonicalHeaderKey("Content-Type")
	contentTypeJSON   = "application/json"
	contentTypeXML    = "application/xml"
	contentTypeForm   = "application/x-www-form-urlencoded"

	jsonCheck = regexp.MustCompile(`(?i:(application|text)/(json|.*\+json|json\-.*)(;|$))`)
	xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(xml|.*\+xml)(;|$))`)
)

func SingleJoin(sep string, text ...string) string {
	if len(text) == 0 {
		return ""
	}
	result := text[0]
	for _, next := range text[1:] {
		asep := strings.HasSuffix(result, sep)
		bsep := strings.HasPrefix(next, sep)
		switch {
		case asep && bsep:
			result += next[len(sep):]
		case !asep && !bsep:
			result += sep + next
		default:
			result += next
		}
	}
	return result
}

func SlashJoin(path ...string) string {
	return SingleJoin("/", path...)
}

func JSONToReader(obj interface{}) (io.Reader, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func DecodeJSON(r io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// IsJSONType method is to check JSON content type or not.
func IsJSONType(contentType string) bool {
	return jsonCheck.MatchString(contentType)
}

func XMLToReader(obj interface{}) (io.Reader, error) {
	b, err := xml.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func DecodeXML(r io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// IsXMLType method is to check XML content type or not.
func IsXMLType(contentType string) bool {
	return xmlCheck.MatchString(contentType)
}

// Request represents a request and options to send with the DoRequest function.
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
	// - map[string]interface{}
	// - url.Values
	Params interface{}

	// JSON specifies optional body data for request which can take body,
	// the content-type will be "application/json", it must be one of
	// the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	// - interface{} (will be marshaled with json.Marshal)
	JSON interface{}

	// XML specifies optional body data for request which can take body,
	// the content-type will be "application/xml", it must be one of
	// the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	// - interface{} (will be marshaled with xml.Marshal)
	XML interface{}

	// Form specifies optional body data for request which can take body,
	// the content-type will be "application/x-www-form-urlencoded",
	// it must be one of the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	// - url.Values (will be encoded and wrapped as io.Reader)
	// - map[string]string (will be converted to url.Values)
	// - map[string][]string (will be converted to url.Values)
	// - map[string]interface{} (will be converted to url.Values)
	Form interface{}

	// Body specifies optional body data for request which can take body,
	// the content-type will be detected from the content (may be incorrect),
	// it should be one of the following types:
	// - io.Reader
	// - []byte (will be wrapped with bytes.NewReader)
	// - string (will be wrapped with strings.NewReader)
	Body interface{}

	// Headers will be copied to the request before sent.
	//
	// If "Content-Type" presents, it will replace the default value
	// set by `JSON`, `XML`, `Form`, or `Body`.
	Headers map[string]string

	// Resp specifies an optional destination to unmarshal the response data.
	// if `Unmarshal` is not provided, the header "Content-Type" will be used to
	// detect XML content, else `json.Unmarshal` will be used.
	Resp interface{}

	// Unmarshal specifies an optional function to unmarshal the response data.
	Unmarshal func([]byte, interface{}) error

	// Context specifies an optional context.Context to use with http request.
	Context context.Context

	// Timeout specifies an optional timeout of the http request, if
	// timeout > 0, the request will be attached an timeout context.Context.
	// `Timeout` takes higher priority than `Context`, it both available, only
	// `Timeout` takes effect.
	Timeout time.Duration

	// Client specifies an optional http.Client to do the request, instead of
	// the default http.DefaultClient.
	Client *http.Client

	// DisableRedirect tells the http client don't follow response redirection.
	DisableRedirect bool

	// DumpRequest makes the http request being logged before sent.
	DumpRequest bool

	// DumpResponse makes the http response being logged after received.
	DumpResponse bool

	// RaiseForStatus tells `DoRequest` to report an error if the response
	// status code >= 400. The error will be formatted as "unexpected status: <STATUS>".
	RaiseForStatus bool
}

func (p *Request) buildClient() *http.Client {
	if p.Client == nil &&
		!p.DisableRedirect {
		return http.DefaultClient
	}
	var client http.Client
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
		return nil
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
		return
	}

	var body io.Reader
	var contentType string

	if p.JSON != nil { // JSON
		body, err = p.makeBody(p.JSON, json.Marshal)
		contentType = contentTypeJSON
	} else if p.XML != nil { // XML
		body, err = p.makeBody(p.XML, xml.Marshal)
		contentType = contentTypeXML
	} else if p.Form != nil { // urlencoded form
		body, err = p.makeBody(p.Form, marshalForm)
		contentType = contentTypeForm
	} else if p.Body != nil { // detect content-type from the body data
		var bodyBuf []byte
		switch data := p.Body.(type) {
		case io.Reader:
			bodyBuf, err = ioutil.ReadAll(data)
			if err != nil {
				return err
			}
		case []byte:
			bodyBuf = data
		case string:
			bodyBuf = ToBytes_(data)
		default:
			err = fmt.Errorf("unsupported body data type: %T", data)
			return err
		}
		body = bytes.NewReader(bodyBuf)
		if p.Headers[hdrContentTypeKey] == "" {
			contentType = http.DetectContentType(bodyBuf)
		}
	} // else no body data

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
	return
}

func mergeQuery(reqURL string, params interface{}) (string, error) {
	parsed, err := url.Parse(reqURL)
	if err != nil {
		return "", err
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", err
	}
	switch params := params.(type) {
	case map[string]string:
		for k, v := range params {
			query.Add(k, v)
		}
	case map[string][]string:
		for k, values := range params {
			for _, v := range values {
				query.Add(k, v)
			}
		}
	case map[string]interface{}:
		for k, v := range params {
			switch value := v.(type) {
			case string:
				query.Add(k, fmt.Sprintf(value))
			case []string:
				for _, v := range value {
					query.Add(k, v)
				}
			default:
				query.Add(k, fmt.Sprint(v))
			}
		}
	case url.Values:
		for k, values := range params {
			for _, v := range values {
				query.Add(k, v)
			}
		}
	default:
		err = fmt.Errorf("unsupported params type: %T", params)
		return "", err
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func marshalForm(v interface{}) ([]byte, error) {
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
	case map[string]interface{}:
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
	buf := ToBytes_(form.Encode())
	return buf, nil
}

type marshalFunc func(interface{}) ([]byte, error)

func (p *Request) makeBody(data interface{}, marshal marshalFunc) (io.Reader, error) {
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

func (p *Request) prepareHeaders() {
	if p.Req == nil {
		return
	}
	for k, v := range p.Headers {
		p.Req.Header.Set(k, v)
	}
}

// DoRequest is a convenient function to send request and control redirect
// and debug options. If `Request.Resp` is provided, it will be used as
// destination to try to unmarshal the response body.
//
// Trade-off was taken to balance simplicity and convenience of the function.
//
// For more powerful controls of a http request and convenient utilities,
// one may take a look at the awesome package `https://github.com/go-resty/resty/`.
func DoRequest(req *Request) (respContent []byte, status int, err error) {
	if err = req.prepareRequest(""); err != nil {
		return
	}
	req.prepareHeaders()

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
		var dump Bytes
		dump, err = httputil.DumpRequestOut(httpReq, true)
		if err != nil {
			return
		}
		log.Printf("dump http request:\n%s", dump.String_())
	}

	httpClient := req.buildClient()
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	status = httpResp.StatusCode
	if req.DumpResponse {
		var dump Bytes
		dump, err = httputil.DumpResponse(httpResp, true)
		if err != nil {
			return
		}
		log.Printf("dump http response:\n%s", dump.String_())
	}

	respContent, err = ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return
	}
	if req.RaiseForStatus {
		if httpResp.StatusCode >= 400 {
			err = fmt.Errorf("unexpected status: %v", httpResp.Status)
			return
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
			return
		}
	}
	return
}
