package mediatesting

// This exists entirely to give a hardcoded list of LDFLAGS to the testing binary.
// idk. There's probably a better way.

// #cgo LDFLAGS: -L ../../../build/subprojects/ffmpeg -lavcodec -lavfilter -lavformat -lavutil -lbz2 -llzma -lm -lpostproc -lswresample -lswscale -lz
import "C"
