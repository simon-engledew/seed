package generators

import (
	"strconv"
	"time"
)

type Unquoted string

type Quoted string

func (v Unquoted) String() string {
	return string(v)
}

func (v Unquoted) Escape() bool {
	return false
}

func (v Quoted) String() string {
	return string(v)
}

func (v Quoted) Escape() bool {
	return true
}

type Bool bool

func (u Bool) String() string {
	return strconv.FormatBool(bool(u))
}

func (u Bool) Escape() bool {
	return false
}

type Date time.Time

func (u Date) String() string {
	return time.Time(u).Format("'2006-01-02 15:04:05'")
}

func (u Date) Escape() bool {
	return false
}

type Double float64

func (u Double) String() string {
	return strconv.FormatFloat(float64(u), 'f', -1, 64)
}

func (u Double) Escape() bool {
	return false
}

type UnsignedInt uint64

func (u UnsignedInt) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

func (u UnsignedInt) Escape() bool {
	return false
}

type SignedInt int64

func (u SignedInt) String() string {
	return strconv.FormatInt(int64(u), 10)
}

func (u SignedInt) Escape() bool {
	return false
}
