// Code generated by "stringer -type=Register"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Zero-0]
	_ = x[Ra-1]
	_ = x[Sp-2]
	_ = x[Gp-3]
	_ = x[Tp-4]
	_ = x[T0-5]
	_ = x[T1-6]
	_ = x[T2-7]
	_ = x[S0-8]
	_ = x[S1-9]
	_ = x[A1-10]
	_ = x[A2-11]
	_ = x[A3-12]
	_ = x[A4-13]
	_ = x[A5-14]
	_ = x[A6-15]
	_ = x[A7-16]
	_ = x[S2-17]
	_ = x[S3-18]
	_ = x[S4-19]
	_ = x[S5-20]
	_ = x[S6-21]
	_ = x[S7-22]
	_ = x[S8-23]
	_ = x[S9-24]
	_ = x[S10-25]
	_ = x[S11-26]
	_ = x[T3-27]
	_ = x[T4-28]
	_ = x[T5-29]
	_ = x[T6-30]
	_ = x[Pc-31]
}

const _Register_name = "ZeroRaSpGpTpT0T1T2S0S1A1A2A3A4A5A6A7S2S3S4S5S6S7S8S9S10S11T3T4T5T6Pc"

var _Register_index = [...]uint8{0, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 55, 58, 60, 62, 64, 66, 68}

func (i Register) String() string {
	if i >= Register(len(_Register_index)-1) {
		return "Register(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Register_name[_Register_index[i]:_Register_index[i+1]]
}
