package mediatesting

// This exists entirely to give a hardcoded list of LDFLAGS to the testing binary.
// idk. There's probably a better way.

// #cgo LDFLAGS: -L ../../../build/subprojects/FFmpeg -L ../../../build/subprojects/c2pa_go -lavcodec -lavfilter -lavformat -lavutil -lc2pa -lm -lpostproc -lswresample -lswscale -lz
import "C"
