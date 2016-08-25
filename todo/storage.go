package todo

import (
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
)

type Store interface {
	Get(channel string) Todo
	Put(channel string, t Todo)
}

type boltStore struct {
	db *bolt.DB
}

var bucketName = []byte("todos")

func (s *boltStore) Get(channel string) (t Todo) {
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return json.Unmarshal(b.Get([]byte(channel)), &t)
	})
	if err != nil {
		log.Println("ERROR fetching Todo:", err)
	}
	return
}

func (s *boltStore) Put(channel string, t Todo) {

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		cnt, err := json.Marshal(t)
		if err != nil {
			return err
		}

		b.Put([]byte(channel), cnt)

		return nil
	})
	if err != nil {
		log.Println("ERROR saving Todo:", err)
	}
}
