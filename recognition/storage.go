package recognition

import (
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
)

type Store interface {
	Get(ts string) *Recognition
	Put(*Recognition)
	All() map[string]*Recognition
}

type boltStore struct {
	db *bolt.DB
}

var bucketName = []byte("recognitions")

func (s *boltStore) Get(ts string) (r *Recognition) {
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return json.Unmarshal(b.Get([]byte(ts)), &r)
	})
	if err != nil {
		log.Println("ERROR fetching recognition:", err)
	}
	return
}

func (s *boltStore) Put(r *Recognition) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		cnt, err := json.Marshal(r)
		if err != nil {
			return err
		}

		b.Put([]byte(r.MsgTimestamp), cnt)

		return nil
	})
	if err != nil {
		log.Println("ERROR saving recognition:", err)
	}
}

func (s *boltStore) All() map[string]*Recognition {
	out := make(map[string]*Recognition)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		return b.ForEach(func(k, v []byte) error {
			r := &Recognition{}

			if err := json.Unmarshal(v, &r); err != nil {
				return err
			}

			out[r.MsgTimestamp] = r

			return nil
		})
	})
	if err != nil {
		log.Println("ERROR fetching recognition:", err)
	}
	return out
}
