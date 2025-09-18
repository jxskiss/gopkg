package set

// Int is an int set collection.
// The zero value of Int is an empty instance ready to use. A zero Int
// value shall not be copied, or it may result incorrect behavior.
//
// This type is mainly for compatibility with old code, the generic
// implementation Generic[int] is favored over this.
type Int struct {
	Generic[int]
}

// NewInt creates an Int set instance.
func NewInt(vals ...int) Int {
	return Int{New(vals...)}
}

// NewIntWithSize creates an Int set instance with given initial size.
func NewIntWithSize(size int) Int {
	return Int{NewWithSize[int](size)}
}

func (s Int) Del(vals ...int)                 { s.Delete(vals...) }
func (s Int) Diff(other Int) Int              { return Int{s.Generic.Diff(other.Generic)} }
func (s Int) DiffSlice(other []int) Int       { return Int{s.Generic.DiffSlice(other)} }
func (s Int) FilterInclude(slice []int) []int { return s.FilterContains(slice) }
func (s Int) FilterExclude(slice []int) []int { return s.FilterNotContains(slice) }
func (s Int) Intersect(other Int) Int         { return Int{s.Generic.Intersect(other.Generic)} }
func (s Int) IntersectSlice(other []int) Int  { return Int{s.Generic.IntersectSlice(other)} }
func (s Int) Union(other Int) Int             { return Int{s.Generic.Union(other.Generic)} }
func (s Int) UnionSlice(other []int) Int      { return Int{s.Generic.UnionSlice(other)} }

// Int64 is an int64 set collection.
// The zero value of Int64 is an empty instance ready to use. A zero Int64
// value shall not be copied, or it may result incorrect behavior.
//
// This type is mainly for compatibility with old code, the generic
// implementation Generic[int64] is favored over this.
type Int64 struct {
	Generic[int64]
}

// NewInt64 creates an Int64 set instance.
func NewInt64(vals ...int64) Int64 {
	return Int64{New(vals...)}
}

// NewInt64WithSize creates an Int64 set instance with given initial size.
func NewInt64WithSize(size int) Int64 {
	return Int64{NewWithSize[int64](size)}
}

func (s Int64) Del(vals ...int64)                   { s.Delete(vals...) }
func (s Int64) Diff(other Int64) Int64              { return Int64{s.Generic.Diff(other.Generic)} }
func (s Int64) DiffSlice(other []int64) Int64       { return Int64{s.Generic.DiffSlice(other)} }
func (s Int64) FilterInclude(slice []int64) []int64 { return s.FilterContains(slice) }
func (s Int64) FilterExclude(slice []int64) []int64 { return s.FilterNotContains(slice) }
func (s Int64) Intersect(other Int64) Int64         { return Int64{s.Generic.Intersect(other.Generic)} }
func (s Int64) IntersectSlice(other []int64) Int64  { return Int64{s.Generic.IntersectSlice(other)} }
func (s Int64) Union(other Int64) Int64             { return Int64{s.Generic.Union(other.Generic)} }
func (s Int64) UnionSlice(other []int64) Int64      { return Int64{s.Generic.UnionSlice(other)} }

// Int32 is an int32 set collection.
// The zero value of Int32 is an empty instance ready to use. A zero Int32
// value shall not be copied, or it may result incorrect behavior.
//
// This type is mainly for compatibility with old code, the generic
// implementation Generic[int32] is favored over this.
type Int32 struct {
	Generic[int32]
}

// NewInt32 creates an Int32 set instance.
func NewInt32(vals ...int32) Int32 {
	return Int32{New(vals...)}
}

// NewInt32WithSize creates an Int32 set instance with given initial size.
func NewInt32WithSize(size int) Int32 {
	return Int32{NewWithSize[int32](size)}
}

func (s Int32) Del(vals ...int32)                   { s.Delete(vals...) }
func (s Int32) Diff(other Int32) Int32              { return Int32{s.Generic.Diff(other.Generic)} }
func (s Int32) DiffSlice(other []int32) Int32       { return Int32{s.Generic.DiffSlice(other)} }
func (s Int32) FilterInclude(slice []int32) []int32 { return s.FilterContains(slice) }
func (s Int32) FilterExclude(slice []int32) []int32 { return s.FilterNotContains(slice) }
func (s Int32) Intersect(other Int32) Int32         { return Int32{s.Generic.Intersect(other.Generic)} }
func (s Int32) IntersectSlice(other []int32) Int32  { return Int32{s.Generic.IntersectSlice(other)} }
func (s Int32) Union(other Int32) Int32             { return Int32{s.Generic.Union(other.Generic)} }
func (s Int32) UnionSlice(other []int32) Int32      { return Int32{s.Generic.UnionSlice(other)} }

// String is a string set collection.
// The zero value of String is an empty instance ready to use. A zero String
// value shall not be copied, or it may result incorrect behavior.
//
// This type is mainly for compatibility with old code, the generic
// implementation Generic[string] is favored over this.
type String struct {
	Generic[string]
}

// NewString creates a String set instance.
func NewString(vals ...string) String {
	return String{New(vals...)}
}

// NewStringWithSize creates a String set instance with given initial size.
func NewStringWithSize(size int) String {
	return String{NewWithSize[string](size)}
}

func (s String) Del(vals ...string)                    { s.Delete(vals...) }
func (s String) Diff(other String) String              { return String{s.Generic.Diff(other.Generic)} }
func (s String) DiffSlice(other []string) String       { return String{s.Generic.DiffSlice(other)} }
func (s String) FilterInclude(slice []string) []string { return s.FilterContains(slice) }
func (s String) FilterExclude(slice []string) []string { return s.FilterNotContains(slice) }
func (s String) Intersect(other String) String         { return String{s.Generic.Intersect(other.Generic)} }
func (s String) IntersectSlice(other []string) String  { return String{s.Generic.IntersectSlice(other)} }
func (s String) Union(other String) String             { return String{s.Generic.Union(other.Generic)} }
func (s String) UnionSlice(other []string) String      { return String{s.Generic.UnionSlice(other)} }
