package ptr

import "time"

func Bool(v bool) *bool                       { return &v }
func String(v string) *string                 { return &v }
func Int(v int) *int                          { return &v }
func Int8(v int8) *int8                       { return &v }
func Int16(v int16) *int16                    { return &v }
func Int32(v int32) *int32                    { return &v }
func Int64(v int64) *int64                    { return &v }
func Uint(v uint) *uint                       { return &v }
func Uint8(v uint8) *uint8                    { return &v }
func Uint16(v uint16) *uint16                 { return &v }
func Uint32(v uint32) *uint32                 { return &v }
func Uint64(v uint64) *uint64                 { return &v }
func Float32(v float32) *float32              { return &v }
func Float64(v float64) *float64              { return &v }
func Time(v time.Time) *time.Time             { return &v }
func Duration(v time.Duration) *time.Duration { return &v }

func DerefBool(v *bool) bool                       { return Deref(v) }
func DerefString(v *string) string                 { return Deref(v) }
func DerefInt(v *int) int                          { return Deref(v) }
func DerefInt8(v *int8) int8                       { return Deref(v) }
func DerefInt16(v *int16) int16                    { return Deref(v) }
func DerefInt32(v *int32) int32                    { return Deref(v) }
func DerefInt64(v *int64) int64                    { return Deref(v) }
func DerefUint(v *uint) uint                       { return Deref(v) }
func DerefUint8(v *uint8) uint8                    { return Deref(v) }
func DerefUint16(v *uint16) uint16                 { return Deref(v) }
func DerefUint32(v *uint32) uint32                 { return Deref(v) }
func DerefUint64(v *uint64) uint64                 { return Deref(v) }
func DerefFloat32(v *float32) float32              { return Deref(v) }
func DerefFloat64(v *float64) float64              { return Deref(v) }
func DerefTime(v *time.Time) time.Time             { return Deref(v) }
func DerefDuration(v *time.Duration) time.Duration { return Deref(v) }
