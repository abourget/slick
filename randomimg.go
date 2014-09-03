package plotbot

import (
	"math/rand"
	"time"
)

var registeredStrings = make(map[string][]string)

func RegisterStringList(category string, list []string) {
	registeredStrings[category] = list
}

func RandomString(category string) string {
	strList, ok := registeredStrings[category]
	if !ok {
		return ""
	}

	rand.Seed(time.Now().UTC().UnixNano())
	idx := rand.Int() % len(strList)
	return strList[idx]
}
