package easy

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/jxskiss/gopkg/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"
)

var (
	hdrContentTypeKey = http.CanonicalHeaderKey("Content-Type")

	jsonCheck = regexp.MustCompile(`(?i:(application|text)/(json|.*\+json|json\-.*)(;|$))`)
	xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(xml|.*\+xml)(;|$))`)
)

func SingleJoin(sep string, path ...string) string {
	if len(path) == 0 {
		return ""
	}
	result := path[0]
	for _, next := range path[1:] {
		asep := strings.HasSuffix(result, sep)
		bsep := strings.HasPrefix(next, sep)
		switch {
		case asep && bsep:
			result += next[1:]
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
	Req  *http.Request
	Resp interface{}

	Context   context.Context
	Timeout   time.Duration
	Client    *http.Client
	Unmarshal func([]byte, interface{}) error

	DisableRedirect bool
	DumpRequest     bool
	DumpResponse    bool
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

// DoRequest is a convenient function to send request and control redirect
// and debug options. If `Request.Resp` is provided, it will be used as
// destination to try to unmarshal the response body.
//
// Tradeoff was taken to balance simplicity and convenience of the function.
//
// For more powerful controls of a http request and convenient utilities,
// one may take a look at the awesome package `https://github.com/go-resty/resty/`.
func DoRequest(req *Request) (respContent []byte, status int, err error) {
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
		log.Printf("dump request: %v\n", dump.String_())
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
		log.Printf("dump response: %v\n", dump.String_())
	}

	respContent, err = ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return
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
