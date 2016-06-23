package faceoff

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
)

// Challenge is one shot of images being shown, and replies being
// gathered for 10 seconds before declaring a winner.
type Challenge struct {
	// UsersShown is the list of all IDs shown
	UsersShown []string
	// RightAnswerIndex holds the index in `UsersShown` that is the
	// person we're looking for.
	RightAnswerIndex int
	// Replies holds the first replies for each user ID, only first
	// response is accounted for.
	Replies map[string]int
	// FirstCorrectReply holds the user ID of the first user that
	// replied correctly.
	FirstCorrectReply string
	// ImageURL is the image URL to show
	ImageURL string
}

func newChallenge() *Challenge {
	return &Challenge{
		Replies: make(map[string]int),
	}
}

// PickUsers set `UsersShown` and `RightAnswerIndex`
func (c *Challenge) PickUsers(users map[string]*User) {
	// TODO: determine the one we're looking for
	// TODO: determine the other users we'll show
	maxPerf := 0
	for _, u := range users {
		if u.PerformanceScore > maxPerf {
			maxPerf = u.PerformanceScore
		}
	}

	var userPool []string
	for id, u := range users {
		chances := 4
		if u.PerformanceScore != 0 {
			quarter := float64(maxPerf) / float64(u.PerformanceScore)
			if quarter > 0.75 {
				chances = 1
			} else if quarter > 0.50 {
				chances = 2
			} else if quarter > 0.25 {
				chances = 3
			} else {
				chances = 4
			}
		}

		for i := 0; i < chances; i++ {
			userPool = append(userPool, id)
		}
	}

	// Shuffle the userPool - http://stackoverflow.com/questions/12264789/shuffle-array-in-go
	for i := range userPool {
		j := rand.Intn(i + 1)
		userPool[i], userPool[j] = userPool[j], userPool[i]
	}

	// Pick the first 4 different users
	chosenUsers := make(map[string]bool)
	for _, userID := range userPool {
		chosenUsers[userID] = true
		if len(chosenUsers) == 4 {
			break
		}
	}

	for userID := range chosenUsers {
		c.UsersShown = append(c.UsersShown, userID)
	}
	c.RightAnswerIndex = rand.Intn(4)
}

// BuildImage builds a new image with 4 square images and overlays
// numbers on top (translucent). It sends to S3 and returns the public
// URL.
func (c *Challenge) BuildImage(profileURLs []string) ([]byte, error) {
	var profileImgs []image.Image
	for _, url := range profileURLs {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("Couldn't get avatar: %s", err)
		}
		defer resp.Body.Close()

		cnt, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("couldn't read avatar body: %s", err)
		}

		img, _, err := image.Decode(bytes.NewBuffer(cnt))
		if err != nil {
			return nil, fmt.Errorf("decoding image: %s", err)
		}

		profileImgs = append(profileImgs, img)
	}

	numbersImage, _, _ := image.Decode(bytes.NewBuffer(numbersPNG))

	// for 4 images
	finalImage := image.NewRGBA(image.Rect(0, 0, 192+192, 192+192))
	mask := &image.Uniform{color.RGBA{0, 0, 100, 100}}

	// First
	draw.Draw(finalImage, image.Rect(0, 0, 192, 192), profileImgs[0], image.Point{0, 0}, draw.Over)
	draw.DrawMask(finalImage, image.Rect(0, 0, 64, 64), numbersImage, numbersPositions[1].Min, mask, image.Point{0, 0}, draw.Over)

	// Second
	draw.Draw(finalImage, image.Rect(192, 0, 192*2, 192), profileImgs[1], image.Point{0, 0}, draw.Over)
	draw.DrawMask(finalImage, image.Rect(192, 0, 192+64, 64), numbersImage, numbersPositions[2].Min, mask, image.Point{0, 0}, draw.Over)

	// Third
	draw.Draw(finalImage, image.Rect(0, 192, 192, 192*2), profileImgs[2], image.Point{0, 0}, draw.Over)
	draw.DrawMask(finalImage, image.Rect(0, 192, 64, 192+64), numbersImage, numbersPositions[3].Min, mask, image.Point{0, 0}, draw.Over)

	// Fourth
	draw.Draw(finalImage, image.Rect(192, 192, 192*2, 192*2), profileImgs[3], image.Point{0, 0}, draw.Over)
	draw.DrawMask(finalImage, image.Rect(192, 192, 192+64, 192+64), numbersImage, numbersPositions[4].Min, mask, image.Point{0, 0}, draw.Over)

	buf := &bytes.Buffer{}
	png.Encode(buf, finalImage)

	return buf.Bytes(), nil

	//return "https://avatars.slack-edge.com/2014-12-08/3167931031_42ef453717f47b15aa3b_192.jpg", nil
}

func (c *Challenge) HandleUserReply(userID string, index int) {
	if _, present := c.Replies[userID]; present {
		return
	}

	c.Replies[userID] = index

	if c.RightAnswerIndex == index {
		if c.FirstCorrectReply == "" {
			c.FirstCorrectReply = userID
		}
	}
}
