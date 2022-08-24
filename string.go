// Code generated by "stringer -type=Register,Perm,MemErrType -output=string.go"; DO NOT EDIT.

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
	_ = x[A0-10]
	_ = x[A1-11]
	_ = x[A2-12]
	_ = x[A3-13]
	_ = x[A4-14]
	_ = x[A5-15]
	_ = x[A6-16]
	_ = x[A7-17]
	_ = x[S2-18]
	_ = x[S3-19]
	_ = x[S4-20]
	_ = x[S5-21]
	_ = x[S6-22]
	_ = x[S7-23]
	_ = x[S8-24]
	_ = x[S9-25]
	_ = x[S10-26]
	_ = x[S11-27]
	_ = x[T3-28]
	_ = x[T4-29]
	_ = x[T5-30]
	_ = x[T6-31]
	_ = x[Pc-32]
}

const _Register_name = "ZeroRaSpGpTpT0T1T2S0S1A0A1A2A3A4A5A6A7S2S3S4S5S6S7S8S9S10S11T3T4T5T6Pc"

var _Register_index = [...]uint8{0, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 57, 60, 62, 64, 66, 68, 70}

func (i Register) String() string {
	if i >= Register(len(_Register_index)-1) {
		return "Register(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Register_name[_Register_index[i]:_Register_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PERM_EXEC-1]
	_ = x[PERM_WRITE-2]
	_ = x[PERM_READ-4]
	_ = x[PERM_RAW-3]
}

const _Perm_name = "PERM_EXECPERM_WRITEPERM_RAWPERM_READ"

var _Perm_index = [...]uint8{0, 9, 19, 27, 36}

func (i Perm) String() string {
	i -= 1
	if i >= Perm(len(_Perm_index)-1) {
		return "Perm(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Perm_name[_Perm_index[i]:_Perm_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ErrCopy-0]
	_ = x[ErrPerms-1]
}

const _MemErrType_name = "ErrCopyErrPerms"

var _MemErrType_index = [...]uint8{0, 7, 15}

func (i MemErrType) String() string {
	if i >= MemErrType(len(_MemErrType_index)-1) {
		return "MemErrType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MemErrType_name[_MemErrType_index[i]:_MemErrType_index[i+1]]
}
