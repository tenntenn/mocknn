package timesample

import (
	"testing"
	"time"

	"github.com/tenntenn/testtime"
)

func Test(t *testing.T) {
	testtime.SetTime(t, parseTime(t, "03:00:00"))
	F()
}

func parseTime(t *testing.T, v string) time.Time {
	t.Helper()
	tm, err := time.Parse("2006/01/02 15:04:05", "2006/01/02 "+v)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	return tm
}
