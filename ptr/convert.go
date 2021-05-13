package ptr

import "strconv"

// IntToStringp converts x to a string pointer.
func IntToStringp(x int) *string {
	str := strconv.FormatInt(int64(x), 10)
	return &str
}

// IntpToStringp converts x to a string pointer.
// It returns nil if x is nil.
func IntpToStringp(x *int) *string {
	if x == nil {
		return nil
	}
	return IntToStringp(*x)
}

// IntpToString converts x to a string.
// It returns an empty string if x is nil.
func IntpToString(x *int) string {
	if x == nil {
		return ""
	}
	return strconv.FormatInt(int64(*x), 10)
}

// Int32ToStringp converts x to a string pointer.
func Int32ToStringp(x int32) *string {
	str := strconv.FormatInt(int64(x), 10)
	return &str
}

// Int32pToStringp converts x to a string pointer.
// It returns nil if x is nil.
func Int32pToStringp(x *int32) *string {
	if x == nil {
		return nil
	}
	return Int32ToStringp(*x)
}

// Int32pToString converts x to a string.
// It returns an emtpy string if x is nil.
func Int32pToString(x *int32) string {
	if x == nil {
		return ""
	}
	return strconv.FormatInt(int64(*x), 10)
}

// Int64ToStringp converts x to a string pointer.
func Int64ToStringp(x int64) *string {
	str := strconv.FormatInt(x, 10)
	return &str
}

// Int64pToStringp converts x to a string pointer.
// It returns nil if x is nil.
func Int64pToStringp(x *int64) *string {
	if x == nil {
		return nil
	}
	return Int64ToStringp(*x)
}

// Int64pToString converts x to a string.
// It returns an empty string if x is nil.
func Int64pToString(x *int64) string {
	if x == nil {
		return ""
	}
	return strconv.FormatInt(*x, 10)
}

// StringToIntp converts x to an int pointer.
// It returns nil if x is not a valid number string.
func StringToIntp(x string) *int {
	i, err := strconv.ParseInt(x, 0, 0)
	if err != nil {
		return nil
	}
	ii := int(i)
	return &ii
}

// StringpToIntp converts x to an int pointer.
// It returns nil if x is nil or not a valid number string.
func StringpToIntp(x *string) *int {
	if x == nil {
		return nil
	}
	return StringToIntp(*x)
}

// StringpToInt converts x to an integer.
// It returns zero if x is nil or not a valid number string.
func StringpToInt(x *string) int {
	if x == nil {
		return 0
	}
	i, _ := strconv.ParseInt(*x, 0, 0)
	return int(i)
}

// StringToInt32p converts x to an int32 pointer.
// It returns nil if x is not a valid number string.
func StringToInt32p(x string) *int32 {
	i, err := strconv.ParseInt(x, 0, 0)
	if err != nil {
		return nil
	}
	ii := int32(i)
	return &ii
}

// StringpToInt32p converts x to an int32 pointer.
// It returns nil if x is nil or not a valid number string.
func StringpToInt32p(x *string) *int32 {
	if x == nil {
		return nil
	}
	return StringToInt32p(*x)
}

// StringpToInt32 converts x to an integer.
// It returns zero if x is nil or not a valid number string.
func StringpToInt32(x *string) int32 {
	if x == nil {
		return 0
	}
	i, _ := strconv.ParseInt(*x, 0, 0)
	return int32(i)
}

// StringToInt64p converts x to an int64 pointer.
// It returns nil if x is not a valid number string.
func StringToInt64p(x string) *int64 {
	i, err := strconv.ParseInt(x, 0, 0)
	if err != nil {
		return nil
	}
	return &i
}

// StringpToInt64p converts x to an int64 pointer.
// It returns nil if x is nil or not a valid number string.
func StringpToInt64p(x *string) *int64 {
	if x == nil {
		return nil
	}
	return StringToInt64p(*x)
}

// StringpToInt64 converts x to an int64 value.
// It returns zero if x is nil or not a valid number string.
func StringpToInt64(x *string) int64 {
	if x == nil {
		return 0
	}
	i, _ := strconv.ParseInt(*x, 0, 0)
	return i
}
