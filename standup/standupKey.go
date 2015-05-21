package standup

import (
	"bytes"
	"strconv"
)

const standupPrefix = "standup:stand"

type standupKey struct {
	date  standupDate
	email string
}

func standupKeyFromBytes(key []byte) standupKey {
	// standup:stand:unix:email

	fields := bytes.Split(key, []byte(":"))

	skey := standupKey{}

	if len(fields) == 4 {
		skey.email = string(fields[3])
	}

	if len(fields) > 2 {
		unixStr := string(fields[2])
		unix, err := strconv.ParseInt(unixStr, 10, 64)
		if err != nil {
			skey.date = standupDate{}
		}
		skey.date = unixToStandupDate(unix)
	}

	return skey

}

func (k standupKey) key() []byte {
	keystr := standupPrefix + ":" + k.date.toUnixUTCString()

	// partial key construction is useful as we can use this to grab all users
	// at a certain standupDate. To make a partial key only initialize a keyStruct
	// with an initialized standupDate.
	if k.email != "" {
		keystr += ":" + k.email
	}
	return []byte(keystr)
}
