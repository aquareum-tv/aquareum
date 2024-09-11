package aqtime

import "time"

// return a consistently formatted timestamp
func FormatMillis(ms int64) string {
	return time.UnixMilli(ms).UTC().Format("2006-01-02T15:04:05.999Z")
}
