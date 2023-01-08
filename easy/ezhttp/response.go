package ezhttp

import (
	"net/http"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

var (
	newlineBytes = []byte{'\n'}

	rspFailedMarshalJSON = []byte("Failed to marshal JSON data.")
)

// JSON serializes the given data as JSON into the response body.
// It also sets the Content-Type as "application/json".
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	jsonBuf, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(rspFailedMarshalJSON)
		return
	}
	w.Header().Set(hdrContentTypeKey, contentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(jsonBuf)
}

// JSONHumanFriendly serializes the given data as pretty JSON (indented + newlines)
// into the response body.
// It also sets the Content-Type as "application/json".
//
// WARNING: we recommend using this only for development purposes
// since printing pretty JSON is more CPU and bandwidth consuming.
// Use JSON() instead.
func JSONHumanFriendly(w http.ResponseWriter, statusCode int, data interface{}) {
	jsonBuf, err := json.HumanFriendly.MarshalIndent(data, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(rspFailedMarshalJSON)
		return
	}
	addNewline := false
	if len(jsonBuf) > 0 && jsonBuf[len(jsonBuf)-1] != '\n' {
		addNewline = true
	}
	w.Header().Set(hdrContentTypeKey, contentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(jsonBuf)
	if addNewline {
		w.Write(newlineBytes)
	}
}
