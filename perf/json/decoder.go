package json

import "io"

// Decoder is a wrapper of encoding/json.Decoder.
// It provides same methods as encoding/json.Decoder but with method
// chaining capabilities.
//
// See encoding/json.Decoder for detailed document.
type Decoder struct {
	UnderlyingDecoder
}

// NewDecoder returns a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{getImpl().NewDecoder(r)}
}

// UseNumber causes the Decoder to unmarshal a number into an interface{}
// as a Number instead of as a float64.
func (dec *Decoder) UseNumber() *Decoder {
	dec.UnderlyingDecoder.UseNumber()
	return dec
}

// DisallowUnknownFields causes the Decoder to return an error when the
// destination is a struct and the input contains object keys which do
// not match any non-ignored, exported fields in the destination.
func (dec *Decoder) DisallowUnknownFields() *Decoder {
	dec.UnderlyingDecoder.DisallowUnknownFields()
	return dec
}
