// Package base62 implements a compact and fast implementation of base62
// encoding/decoding algorithm, which is inspired by the java implementation
// by glowfall at https://github.com/glowfall/base62. This implementation
// is much faster than big.Int based implementation, and is not much slower
// than typical base64 implementations.
//
// The package has been moved to github.com/jxskiss/base62 as a standalone
// repository. Check the new repository for more details.
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
