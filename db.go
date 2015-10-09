package slick

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

const slickDBDefaultBucket = "slick"

// GetDBKey retrieves a `key` from persistent storage and JSON
// unmarshales it into `v`.
func (bot *Bot) GetDBKey(key string, v interface{}) error {
	return bot.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(slickDBDefaultBucket))
		if err != nil {
			return err
		}

		val := bucket.Get([]byte(key))
		if val == nil {
			return fmt.Errorf("not found")
		}

		err = json.Unmarshal(val, &v)
		if err != nil {
			return err
		}

		return nil
	})
}

// SetDBKey sets a
func (bot *Bot) PutDBKey(key string, v interface{}) error {
	return bot.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(slickDBDefaultBucket))
		if err != nil {
			return err
		}

		jsonRes, err := json.Marshal(v)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(key), jsonRes)
		if err != nil {
			return err
		}

		return nil
	})
}
