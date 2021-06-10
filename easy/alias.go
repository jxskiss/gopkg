package easy

import "github.com/jxskiss/gopkg/gemap"

type (
	Map     = gemap.Map
	SafeMap = gemap.SafeMap
)

var (
	NewMap     = gemap.NewMap
	NewSafeMap = gemap.NewSafeMap
)

var (
	MapKeys      = gemap.MapKeys
	MapValues    = gemap.MapValues
	IntKeys      = gemap.IntKeys
	IntValues    = gemap.IntValues
	StringKeys   = gemap.StringKeys
	StringValues = gemap.StringValues

	MergeMaps   = gemap.MergeMaps
	MergeMapsTo = gemap.MergeMapsTo
)
