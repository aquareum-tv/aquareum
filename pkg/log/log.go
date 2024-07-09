/*
Package clog provides Context with logging metadata, as well as logging helper functions.
*/
package log

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/golang/glog"
)

// unique type to prevent assignment.
type clogContextKeyType struct{}

// singleton value to identify our logging metadata in context
var clogContextKey = clogContextKeyType{}

var defaultLogLevel glog.Level = 3

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

var badGlyphs = " \n"

// Actual log handler; the others have wrappers to properly handle stack depth
func (v *VerboseLogger) log(ctx context.Context, message string, args ...any) {
	if !glog.V(v.level) {
		return
	}
	keyCol := color.New(color.FgMagenta).SprintFunc()
	valCol := color.New(color.FgGreen).SprintFunc()
	messageCol := color.New(color.FgCyan).SprintFunc()
	meta, _ := ctx.Value(clogContextKey).(metadata)
	allArgs := []any{}
	allArgs = append(allArgs, args...)
	allArgs = append(allArgs, meta.Flat()...)
	allArgs = append(allArgs, "caller", caller(3))
	str := messageCol(message)
	for i := range allArgs {
		if i%2 == 0 {
			continue
		}
		safeVal := fmt.Sprintf("%s", allArgs[i])
		if strings.ContainsAny(safeVal, badGlyphs) {
			safeVal = fmt.Sprintf("%q", allArgs[i])
		}
		str = fmt.Sprintf("%s %s=%s", str, keyCol(allArgs[i-1]), valCol(safeVal))
	}
	fmt.Println(str)
}

func (v *VerboseLogger) Log(ctx context.Context, message string, args ...any) {
	v.log(ctx, message, args...)
}

func Log(ctx context.Context, message string, args ...any) {
	V(defaultLogLevel).log(ctx, message, args...)
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
