package standup

import (
	"testing"
	"time"
)

func TestKeyFromBytes(t *testing.T) {

	now := time.Now()

	key := standupKey{
		date:  getStandupDate(0),
		email: "bot@bot.ly",
	}

	key2 := standupKeyFromBytes(key.key())

	if key2.email != "bot@bot.ly" {
		t.Error("expected email to be 'bot@bot.ly' instead got", key2.email)
	}

	if key2.date.year != now.Year() {
		t.Error("expected", key2.date.year, "to be", now.Year())
	}

	if key2.date.month != now.Month() {
		t.Error("expected", key2.date.month, "to be", now.Month())
	}

	if key2.date.day != now.Day() {
		t.Error("expected", key2.date.day, "to be", now.Day())
	}
}
