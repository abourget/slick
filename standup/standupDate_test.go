package standup

import (
	"strconv"
	"testing"
	"time"
)

func TestNextDate(t *testing.T) {

	tomorrow := time.Now().Add(24 * time.Hour)

	d := getStandupDate(0).next()

	if d.year != tomorrow.Year() {
		t.Error("expected", d.year, "to be", tomorrow.Year())
	}

	if d.month != tomorrow.Month() {
		t.Error("expected", d.month, "to be", tomorrow.Month())
	}

	if d.day != tomorrow.Day() {
		t.Error("expected", d.day, "to be", tomorrow.Day())
	}

}

func TestUnixAndBack(t *testing.T) {

	d := getStandupDate(0)
	unixStr := d.toUnixUTCString()
	unix, err := strconv.ParseInt(unixStr, 10, 64)
	if err != nil {
		t.Error("Error parsing unix date")
	}

	d2 := unixToStandupDate(unix)

	if d.year != d2.year {
		t.Error("expected", d2.year, "to be", d.year)
	}

	if d.month != d.month {
		t.Error("expected", d2.month, "to be", d.month)
	}

	if d.day != d2.day {
		t.Error("expected", d2.day, "to be", d.day)
	}

}
