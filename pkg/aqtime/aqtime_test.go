package aqtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTimeFormat(t *testing.T) {
	aqt := FromMillis(1726251017090)
	require.Equal(t, "2024-09-13T18:10:17.090Z", aqt.String())
	yr, mon, day, hr, min, sec, ms := aqt.Parts()
	require.Equal(t, "2024", yr)
	require.Equal(t, "09", mon)
	require.Equal(t, "13", day)
	require.Equal(t, "18", hr)
	require.Equal(t, "10", min)
	require.Equal(t, "17", sec)
	require.Equal(t, "090", ms)
}
