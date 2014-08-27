package standup

import (
	"bytes"
	"encoding/gob"
	"log"
)

func (standup *Standup) LoadData() {
	bot := standup.bot
	redis := bot.RedisPool.Get()
	defer redis.Close()

	fixup := func() {
		dataMap := make(DataMap)
		standup.data = &dataMap
	}

	res, err := redis.Do("GET", "plotbot:standup")
	if err != nil {
		log.Println("Standup: Couldn't load data from redis. Using fresh data.")
		fixup()
		return
	}

	asBytes, _ := res.([]byte)
	dec := gob.NewDecoder(bytes.NewBuffer(asBytes))
	err = dec.Decode(standup.data)
	if err != nil {
		log.Println("Standup: Unable to decode data from redis. Using fresh data.")
		fixup()
	}
}

func (standup *Standup) FlushData() {
	bot := standup.bot
	redis := bot.RedisPool.Get()
	defer redis.Close()

	buf := bytes.NewBuffer([]byte(""))
	enc := gob.NewEncoder(buf)
	enc.Encode(standup.data)

	_, err := redis.Do("SET", "plotbot:standup", buf.String())
	if err != nil {
		log.Println("ERROR: Couldn't redis FlushData()")
	}
}
