/*
Package clog provides Context with logging metadata, as well as logging helper functions.
*/
package log

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func init() {
	w := os.Stderr

	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
			NoColor:    !isatty.IsTerminal(w.Fd()),
		}),
	))
}

// unique type to prevent assignment.
type clogContextKeyType struct{}

// singleton value to identify our logging metadata in context
var clogContextKey = clogContextKeyType{}

var errorLogLevel glog.Level = 1
var warnLogLevel glog.Level = 2
var defaultLogLevel glog.Level = 3
var debugLogLevel glog.Level = 4
var traceLogLevel glog.Level = 9

// basic type to represent logging container. logging context is immutable after
// creation, so we don't have to worry about locking.
type metadata map[string]any

func init() {
	// Set default v level to 3; this is overridden in main() but is useful for tests
	vFlag := flag.Lookup("v")
	// nolint:errcheck
	vFlag.Value.Set(fmt.Sprintf("%d", defaultLogLevel))
}

type VerboseLogger struct {
	level glog.Level
}

// implementation of our logger aware of glog -v=[0-9] levels
func V(level glog.Level) *VerboseLogger {
	return &VerboseLogger{level: level}
}

func (m metadata) Flat() []any {
	out := []any{}
	for k, v := range m {
		out = append(out, k)
		out = append(out, v)
	}
	return out
}

// Return a new context, adding in the provided values to the logging metadata
func WithLogValues(ctx context.Context, args ...string) context.Context {
	oldMetadata, _ := ctx.Value(clogContextKey).(metadata)
	// No previous logging found, set up a new map
	if oldMetadata == nil {
		oldMetadata = metadata{}
	}
	var newMetadata = metadata{}
	for k, v := range oldMetadata {
		newMetadata[k] = v
	}
	for i := range args {
		if i%2 == 0 {
			continue
		}
		newMetadata[args[i-1]] = args[i]
	}
	return context.WithValue(ctx, clogContextKey, newMetadata)
}

// Actual log handler; the others have wrappers to properly handle stack depth
func (v *VerboseLogger) log(ctx context.Context, message string, fn func(string, ...any), args ...any) {
	if !glog.V(v.level) {
		return
	}
	// I want a compile time assertion for this... but short of that let's be REALLY ANNOYING
	if len(args)%2 != 0 {
		for range 6 {
			fmt.Println("!!!!!!!!!!!!!!!! FOLLOWING LOG LINE HAS AN ODD NUMBER OF ARGUMENTS !!!!!!!!!!!!!!!!")
		}
	}
	meta, _ := ctx.Value(clogContextKey).(metadata)
	hasCaller := false

	allArgs := []any{}
	allArgs = append(allArgs, args...)
	allArgs = append(allArgs, meta.Flat()...)
	for i := range args {
		if i%2 == 0 {
			continue
		}
		if args[i-1] == "caller" {
			hasCaller = true
		}
	}
	if !hasCaller {
		allArgs = append(allArgs, "caller", caller(3))
	}
	fn(message, allArgs...)
}

func (v *VerboseLogger) Log(ctx context.Context, message string, args ...any) {
	if v.level >= 4 {
		v.log(ctx, message, slog.Debug, args...)
	} else {
		v.log(ctx, message, slog.Info, args...)
	}
}

func Error(ctx context.Context, message string, args ...any) {
	V(errorLogLevel).log(ctx, message, slog.Error, args...)
}

func Warn(ctx context.Context, message string, args ...any) {
	V(warnLogLevel).log(ctx, message, slog.Warn, args...)
}

func Log(ctx context.Context, message string, args ...any) {
	V(defaultLogLevel).log(ctx, message, slog.Info, args...)
}

func Debug(ctx context.Context, message string, args ...any) {
	V(debugLogLevel).log(ctx, message, slog.Debug, args...)
}

func Trace(ctx context.Context, message string, args ...any) {
	V(traceLogLevel).log(ctx, message, slog.Debug, args...)
}

// returns filenames relative to aquareum root
// e.g. handlers/misttriggers/triggers.go:58
func caller(depth int) string {
	_, myfile, _, _ := runtime.Caller(0)
	// This assumes that the root directory of aquareum is two levels above this folder.
	// If that changes, please update this rootDir resolution.
	rootDir := filepath.Join(filepath.Dir(myfile), "..", "..")
	_, file, line, _ := runtime.Caller(depth)
	rel, _ := filepath.Rel(rootDir, file)
	return rel + ":" + strconv.Itoa(line)
}
