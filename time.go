package util

import (
	"encoding/binary"
	"time"
)

const (
	kNtpJan1970             uint32  = 2208988800
	kMagicNtpFractionalUnit float64 = 4.294967296E+9
	kFractionsPerSecond     uint64  = 0x100000000
	kNtpFracPerMs           float64 = 4.294967296E6 // 2^32 / 1000
)

// NowMs return crrent UTC time(milliseconds) with 32bit
func NowMs() uint32 {
	return uint32(time.Now().UTC().UnixNano() / int64(time.Millisecond))
}

// NowMs64 return crrent UTC time(milliseconds) with 64bit
func NowMs64() int64 {
	return time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

func TimeSince(earlier int64) int64 {
	return NowMs64() - earlier
}

// Sleep to wait some milliseconds and then wake
func Sleep(ms int) {
	timer := time.NewTimer(time.Millisecond * time.Duration(ms))
	<-timer.C
}

type NtpTime struct {
	Value_ uint64
}

func (n *NtpTime) Set(seconds, fractions uint32) {
	n.Value_ = uint64(seconds)*kFractionsPerSecond + uint64(fractions)
}

func (n *NtpTime) Seconds() uint32 {
	return uint32(n.Value_ / kFractionsPerSecond)
}

func (n *NtpTime) Fractions() uint32 {
	return uint32(n.Value_ % kFractionsPerSecond)
}

func (n *NtpTime) ToMs() int64 {
	frac_ms := float64(n.Fractions()) / kNtpFracPerMs
	return 1000*int64(n.Seconds()) + int64(frac_ms+0.5)
}

func (n *NtpTime) Value() uint64 {
	return n.Value_
}

func (n *NtpTime) Parse(combine uint64) {
	var ntp [8]byte
	binary.BigEndian.PutUint64(ntp[0:8], combine)
	seconds := binary.BigEndian.Uint32(ntp[0:4])
	fractions := binary.BigEndian.Uint32(ntp[4:8])
	n.Set(seconds, fractions)
}

func (n *NtpTime) Combine() uint64 {
	var ntp [8]byte
	binary.BigEndian.PutUint32(ntp[0:4], n.Seconds())
	binary.BigEndian.PutUint32(ntp[4:8], n.Fractions())
	return binary.BigEndian.Uint64(ntp[0:8])
}

func NtpToRtp(ntp *NtpTime, freq uint32) uint32 {
	if ntp == nil {
		return 0
	}
	tmp := uint32((uint64(ntp.Fractions()) * uint64(freq)) >> 32)
	return ntp.Seconds()*freq + tmp
}

func CompactNtp(ntp *NtpTime) uint32 {
	if ntp == nil {
		return 0
	}
	return (ntp.Seconds() << 16) | (ntp.Fractions() >> 16)
}

func CurrentNtpTime() *NtpTime {
	now_ms := NowMs64()
	seconds := uint32(now_ms/1000) + kNtpJan1970
	fractions := uint32(float64(now_ms%1000) * kMagicNtpFractionalUnit / 1000.0)
	ntp := &NtpTime{}
	ntp.Set(seconds, fractions)
	return ntp
}

func CurrentNtpInMilliseconds() int64 {
	now_ms := NowMs64()
	return now_ms + 1000*int64(kNtpJan1970)
}

func CompactNtpRttToMs(compact_ntp_interval uint32) int64 {
	// Interval to convert expected to be positive, e.g. rtt or delay.
	// Because interval can be derived from non-monotonic ntp clock,
	// it might become negative that is indistinguishable from very large values.
	// Since very large rtt/delay are less likely than non-monotonic ntp clock,
	// those values consider to be negative and convert to minimum value of 1ms.
	if compact_ntp_interval > 0x80000000 {
		return 1
	}
	// Convert to 64bit value to avoid multiplication overflow.
	value := int64(compact_ntp_interval)
	// To convert to milliseconds need to divide by 2^16 to get seconds,
	// then multiply by 1000 to get milliseconds. To avoid float operations,
	// multiplication and division swapped.
	ms := DivideRoundToNearest(value*1000, 1<<16)
	// Rtt value 0 considered too good to be true and increases to 1.
	if ms < 1 {
		ms = 1
	}
	return ms
}

func DivideRoundToNearest(x int64, y uint32) int64 {
	// Callers ensure x is positive and x + y / 2 doesn't overflow.
	return (x + int64(y/2)) / int64(y)
}
