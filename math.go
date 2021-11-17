package util

import (
	"math"
	"math/rand"
	"reflect"
	"time"
)

var numbers = []rune("0123456789")

// rune is alias of int, for char
func RandomNumber(n int) string {
	rand.Seed(time.Now().UnixNano())

	s := make([]rune, n)
	for i := range s {
		s[i] = numbers[rand.Intn(10)]
	}
	return string(s)
}

// RandomInt return a random int number.
func RandomInt(n int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(n)
}

// RandomInt32 return a random uint32 number.
func RandomUint32() uint32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}

// RandomString return a random n-char(a-zA-Z0-9) string.
func RandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

// Min return the minimum int of x,y
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Max return the maximum int of x,y
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// Abs return abs value
func Abs(x int) int {
	if x < 0 {
		return 0 - x
	}
	return x
}

// Clamp x in [min, max]
func Clamp(x, min, max int) int {
	if x < min {
		return min
	} else if x > max {
		return max
	}
	return x
}

// The number priorities
var kNumberPriorities = map[reflect.Kind]interface{}{
	reflect.Int8:    1,
	reflect.Uint8:   2,
	reflect.Int16:   3,
	reflect.Uint16:  4,
	reflect.Int32:   5,
	reflect.Uint32:  6,
	reflect.Int:     7,
	reflect.Uint:    8,
	reflect.Int64:   9,
	reflect.Uint64:  10,
	reflect.Float32: 11,
	reflect.Float64: 12,
}

func IsNumberKind(x interface{}) bool {
	kind := reflect.TypeOf(x).Kind()
	_, ok := kNumberPriorities[kind]
	return ok
}

func GetFloat64(x interface{}) float64 {
	return getFloat64(x)
}

func getFloat64(v interface{}) float64 {
	x := reflect.ValueOf(v)
	switch x.Kind() {
	case reflect.Int8:
		return float64(x.Int())
	case reflect.Uint8:
		return float64(x.Uint())
	case reflect.Int16:
		return float64(x.Int())
	case reflect.Uint16:
		return float64(x.Uint())
	case reflect.Int32:
		return float64(x.Int())
	case reflect.Uint32:
		return float64(x.Uint())
	case reflect.Int:
		return float64(x.Int())
	case reflect.Uint:
		return float64(x.Uint())
	case reflect.Int64:
		return float64(x.Int())
	case reflect.Uint64:
		return float64(x.Uint())
	case reflect.Float32:
		return float64(x.Float())
	case reflect.Float64:
		return x.Float()
	default:
	}
	return 0.0
}

// @param x:     return x when in [a, b]; if x < a; return a; if x > b return b
// @param rtype: returned type(base type)
func Saturated(x interface{}, rtype interface{}) interface{} {
	switch rtype.(type) {
	case int8:
		return int8(Clampf(x, math.MinInt8, math.MaxInt8))
	case uint8:
		return uint8(Clampf(x, 0, math.MaxUint8))
	case int16:
		return int16(Clampf(x, math.MinInt16, math.MaxInt16))
	case uint16:
		return uint16(Clampf(x, 0, math.MaxUint16))
	case int32:
		return int32(Clampf(x, math.MinInt32, math.MaxInt32))
	case uint32:
		// NOTE: avoid overflow
		return uint32(Clampf(x, 0, uint32(math.MaxUint32)))
	case int:
		// FIXME: int >= 4bytes(4 or 8)
		return int(Clampf(x, math.MinInt32, math.MaxInt32))
	case uint:
		// NOTE: avoid overflow
		return uint(Clampf(x, 0, uint32(math.MaxUint32)))
	case int64:
		// NOTE: avoid overflow
		return int64(Clampf(x, int64(math.MinInt64), int64(math.MaxInt64)))
	case uint64:
		// NOTE: avoid overflow
		return uint64(Clampf(x, 0, uint64(math.MaxUint64)))
	case float32:
		return float32(getFloat64(x))
	case float64:
		return getFloat64(x)
	default:
	}
	return nil
}

func Absf(x interface{}) float64 {
	return math.Abs(getFloat64(x))
}

func Sinf(x interface{}) float64 {
	return math.Sin(getFloat64(x))
}

func Cosf(x interface{}) float64 {
	return math.Cos(getFloat64(x))
}

func Ceilf(x interface{}) float64 {
	return math.Ceil(getFloat64(x))
}

func Sqrtf(x interface{}) float64 {
	return math.Sqrt(getFloat64(x))
}

func Expf(x interface{}) float64 {
	return math.Exp(getFloat64(x))
}

func Logf(x interface{}) float64 {
	return math.Log(getFloat64(x))
}

func Log10f(x interface{}) float64 {
	return math.Log10(getFloat64(x))
}

func Powf(x, y interface{}) float64 {
	return math.Pow(getFloat64(x), getFloat64(y))
}

func Minf(x, y interface{}) float64 {
	return math.Min(getFloat64(x), getFloat64(y))
}

func Maxf(x, y interface{}) float64 {
	return math.Max(getFloat64(x), getFloat64(y))
}

func Clampf(x, min, max interface{}) float64 {
	x_f := getFloat64(x)
	min_f := getFloat64(min)
	max_f := getFloat64(max)
	if x_f <= min_f {
		return min_f
	} else if x_f >= max_f {
		return max_f
	} else {
		return x_f
	}
}

func SafeClamp(x, min, max interface{}) float64 {
	return Clampf(x, min, max)
}

func ClampToInt16(input int32) int16 {
	if input < math.MinInt16 {
		return math.MinInt16
	} else if input > math.MaxInt16 {
		return math.MaxInt16
	} else {
		return int16(input)
	}
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

func IfInt(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	} else {
		return falseVal
	}
}

/// EQ
type equalInterface interface {
	Equal(b equalInterface) bool
}

// a == b
func EQ(a, b equalInterface) bool {
	return a.Equal(b)
}

// LT/LE/GT/GE
type lessInterface interface {
	Less(b lessInterface) bool
}

// a < b
func LT(a, b lessInterface) bool {
	return a.Less(b)
}

// a <= b or !(a > b)
func LE(a, b lessInterface) bool {
	return !b.Less(a)
}

// a > b
func GT(a, b lessInterface) bool {
	return b.Less(a)
}

// a >= b or !(a < b)
func GE(a, b lessInterface) bool {
	return !a.Less(b)
}
