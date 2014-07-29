package main

import (
	"time"
	"math/rand"
)


var gifmap = makeGif()


func makeGif () ( map[string][]string  ) {

	gifmap := make(map[string][]string)
	gifmap["herpderp"] = []string{
		"http://www.gifcrap.com/g2data/albums/TV/Star%20Wars%20-%20Force%20Push%20-%20Goats%20fall%20over.gif",
		"http://i.imgur.com/ZvZR6Ff.jpg",
		"http://i3.kym-cdn.com/photos/images/original/000/014/538/5FCNWPLR2O3TKTTMGSGJIXFERQTAEY2K.gif",
		"http://i167.photobucket.com/albums/u123/KevinB550/FORCEPUSH/starwarsagain.gif",
		"http://i.imgur.com/dqSIv6j.gif",
		"http://www.gifcrap.com/g2data/albums/TV/Star%20Wars%20-%20Force%20Push%20-%20Gun%20breaks.gif",
		"http://media0.giphy.com/media/qeWa5wV5aeEHC/giphy.gif",
		"http://img40.imageshack.us/img40/2529/obiwan20is20a20jerk.gif",
		"http://img856.imageshack.us/img856/2364/obiwanforcemove.gif",
		"http://img526.imageshack.us/img526/4750/bc6.gif",
		"http://img825.imageshack.us/img825/6373/tumblrluaj77qaoa1qzrlhg.gif",
		"http://img543.imageshack.us/img543/6222/basketballdockingbay101.gif",
		"http://img687.imageshack.us/img687/5711/frap.gif",
		"http://img96.imageshack.us/img96/812/starpigdockingbay101.gif",
		"http://img2.wikia.nocookie.net/__cb20131117184206/halo/images/2/2a/Xt0rt3r.gif",
	}

	gifmap["storm"] = []string{
		"http://static.tumblr.com/ikqttte/OlElnumnn/f9cb7_tumblr_lkfd09xr2y1qfuje9o1_500.gif",
		"http://media.giphy.com/media/8cdBgACkApvt6/giphy.gif",
		"http://25.media.tumblr.com/tumblr_luucaug87A1qluhjfo1_500.gif",
		"http://cdn.mdjunction.com/components/com_joomlaboard/uploaded/images/storm.gif",
		"http://www.churchhousecollection.com/resources/animated-jesus-calms-storm.gif",
		"http://i251.photobucket.com/albums/gg307/angellovernumberone/HEATHERS%20%20MIXED%20WATER%20ANIMATIONS/LightningStorm02.gif",
		"http://i.imgur.com/IF1QM.gif",
		"http://wac.450f.edgecastcdn.net/80450F/screencrush.com/files/2013/04/x-men-storm.gif",
		"http://i.imgur.com/SNLbnO8.gif?1",

	}

	return gifmap
}

func RandomGIF (giftype string) (url string) {

	gifs := gifmap[giftype]
	rand.Seed(time.Now().UTC().UnixNano())
	idx := rand.Int() % len(gifs)
	url = gifs[idx]

	return url
}
