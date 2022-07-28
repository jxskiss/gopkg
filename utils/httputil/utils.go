package httputil

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

var (
	jsonCheck = regexp.MustCompile(`(?i:(application|text)/(json|.*\+json|json\-.*)(;|$))`)
	xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(xml|.*\+xml)(;|$))`)
)

// DecodeJSON decodes a json value from r.
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

// DecodeXML decodes a XML value from r.
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
