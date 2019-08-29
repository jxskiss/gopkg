// package base62 has been moved to github.com/jxskiss/base62.
package base62

import "github.com/jxskiss/base62"

type (
	Encoding          = base62.Encoding
	CorruptInputError = base62.CorruptInputError
)

var (
	NewEncoding = base62.NewEncoding

	Encode         = base62.Encode
	EncodeToString = base62.EncodeToString

	Decode       = base62.Decode
	DecodeString = base62.DecodeString

	FormatInt  = base62.FormatInt
	FormatUint = base62.FormatUint
	AppendInt  = base62.AppendInt
	AppendUint = base62.AppendUint
	ParseInt   = base62.ParseInt
	ParseUint  = base62.ParseUint
)
