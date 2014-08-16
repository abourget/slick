package blaster

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Blaster struct {
	URL        string
	Done       chan bool
	Reply      chan string
	tries      []*try
	triesPipe  chan *try
	triesClear chan bool
}

// When you close `done`, the reply follows ASAP
func New(url string) *Blaster {
	b := &Blaster{
		URL:        url,
		Done:       make(chan bool),
		Reply:      make(chan string),
		triesPipe:  make(chan *try, 100),
		triesClear: make(chan bool),
	}
	return b
}

func (b *Blaster) Start(numBlasters int, seconds time.Duration) {
	for i := 0; i < numBlasters; i++ {
		go b.doBlast()
	}
	go b.manageFlow()
	go b.timedReplies()

	go func() {
		time.Sleep(seconds)
		close(b.Done)
	}()
}

func (b *Blaster) doBlast() {
	for {
		select {
		case <-b.Done:
			return
		default:
		}

		t := &try{start: time.Now()}
		log.Println("Blaster: connecting to ", b.URL)
		res, err := http.Get(b.URL)
		if err != nil {
			log.Println("Blaster: error connecting: ", err)
			continue
		}
		t.end = time.Now()
		t.status = res.StatusCode
		log.Println("Blaster: appending ", t)

		b.triesPipe <- t
	}
}

// Manages flow of tries, and prepares reply
func (b *Blaster) manageFlow() {
	for {
		select {
		case <-b.Done:
			close(b.Reply)
			return
		case t := <-b.triesPipe:
			b.tries = append(b.tries, t)
		case <-b.triesClear:
			b.tries = []*try{}
		}
	}
}

func (b *Blaster) timedReplies() {
	for {
		replyDelay := time.After(1 * time.Second)
		select {
		case <-b.Done:
			return
		case <-replyDelay:
			b.prepareReply()
		}
	}
}

// No one writes to "tries" anymore, we read that and prepare a reply
func (b *Blaster) prepareReply() {
	sum := int64(0)
	numTries := len(b.tries)
	numErrors := 0
	for _, t := range b.tries {
		sum += t.Delta().Nanoseconds()
		if t.status > 200 {
			numErrors += 1
		}
	}
	avg := float64(sum) / float64(numTries)

	b.Reply <- fmt.Sprintf("Blaster: avg %.3fms, %d errors", avg/1000000.0, numErrors)
	b.triesClear <- true
}
