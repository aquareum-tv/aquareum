package mediatesting

// This exists entirely to give a hardcoded list of LDFLAGS to the testing binary.
// idk. There's probably a better way.

// #cgo LDFLAGS: -L ../../../build/subprojects/ffmpeg -lz -lbz2 -llzma -lavutil -lavcodec -lavfilter -lm -lavformat -lavformat -lavcodec -lavutil -lswresample -lavfilter -lswscale -lpostproc
import "C"

func Nothing() {}
