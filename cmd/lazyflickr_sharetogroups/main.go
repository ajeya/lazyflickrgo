// cmd/lazyflickr_sharetogroups share photo to groups.
/*
Install:

	go install github.com/toomore/lazyflickrgo/cmd/lazyflickr_sharetogroups

Usage:

	lazyflickr_sharetogroups [flags]

The flags are:

	-apikey
		Flickr API key, default get from env `FLICKRAPIKEY`

	-secret
		Flickr secret, default get from env `FLICKRSECRET`

	-userid
		Flickr userid(nsid), default get from env `FLICKRUSER`

	-albumid
		Album/Set number ID

	-groupid
		Group number ID

	-n
		share photos num. default is 6

	-tags
		Search tags, ',' for split more

	-dryrun
		Show result without post to groups

Example:

share tag:`lomo,japan` to lomo, lomo.tw groups

	lazyflickr_sharetogroups -tags lomo,japan -groupid 40732537997@N01,72262428@N00

*/
package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/toomore/lazyflickrgo/flickr"
	"github.com/toomore/lazyflickrgo/jsonstruct"
)

var (
	userID  = flag.String("userid", os.Getenv("FLICKRUSER"), "User number ID")
	albumID = flag.String("albumid", "", "Album/Set number ID")
	groupID = flag.String("groupid", "", "Group number ID")
	apikey  = flag.String("apikey", os.Getenv("FLICKRAPIKEY"), "Flickr API Key")
	secret  = flag.String("secret", os.Getenv("FLICKRSECRET"), "Flickr secret")
	shareN  = flag.Int("n", 6, "Per share num")
	tags    = flag.String("tags", "", "Search tags, ',' for split more")
	dryrun  = flag.Bool("dryrun", false, "Show result without post to groups")
	info    = color.New(color.Bold, color.FgGreen).SprintfFunc()
	warn    = color.New(color.Bold, color.FgRed).SprintfFunc()
	debugc  = color.New(color.Bold, color.FgHiYellow).SprintfFunc()
	wg      sync.WaitGroup
	photos  []jsonstruct.Photo
)

func fromSets(f *flickr.Flickr) []jsonstruct.Photo {
	var result []jsonstruct.Photo
	for _, albumid := range strings.Split(*albumID, ",") {
		for _, albumdata := range f.PhotosetsGetPhotosAll(albumid, *userID) {
			result = append(result, albumdata.Photoset.Photos.Photo...)
		}
	}

	return result
}

func fromSearch(f *flickr.Flickr) []jsonstruct.Photo {
	args := make(map[string]string)
	args["tags"] = *tags
	args["tag_mode"] = "all"
	args["sort"] = "date-posted-desc"
	args["per_page"] = "500"
	args["user_id"] = *userID

	searchResult := f.PhotosSearch(args)

	var result []jsonstruct.Photo
	for _, val := range searchResult {
		result = append(result, val.Photos.Photo...)
	}

	return result
}

func addToPool(f *flickr.Flickr, photo jsonstruct.Photo, groupid string, val int) {
	runtime.Gosched()
	defer wg.Done()
	if *dryrun == false {
		resp := f.GroupsPoolsAdd(groupid, photo.ID)
		if resp.Stat == "ok" {
			log.Println(info("[%s] %s %s", groupid, photo.ID, photo.Title))
		} else {
			log.Println(warn("[%s] %s(%d) %s %s", groupid, resp.Message, resp.Code, photo.ID, photo.Title))
		}
	} else {
		log.Println(debugc("[DryRun] [%s] %s %s", groupid, photo.ID, photo.Title))
	}
}

func send(groupid string, photos []jsonstruct.Photo, randlist []int, f *flickr.Flickr) {
	runtime.Gosched()
	for _, val := range randlist {
		photo := photos[val]
		log.Println(info("Pick up photo: %d [%s] %+v", val, photo.ID, photo))
		go addToPool(f, photo, groupid, val)
	}
}

func main() {
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	var (
		num int
		f   *flickr.Flickr
	)

	f = flickr.NewFlickr(*apikey, *secret)
	f.AuthToken = os.Getenv("FLICKRUSERTOKEN")

	if *tags == "" {
		photos = fromSets(f)
	} else {
		photos = fromSearch(f)
	}

	num = len(photos)
	if num <= *shareN {
		*shareN = num
	}

	for _, groupid := range strings.Split(*groupID, ",") {
		wg.Add(*shareN)

		var randlist []int
		startInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
		if (num - *shareN) > 0 {
			startInt = startInt % (num - *shareN)
			randlist = rand.New(rand.NewSource(time.Now().UnixNano())).Perm(num)[startInt : startInt+*shareN]
		} else {
			randlist = rand.New(rand.NewSource(time.Now().UnixNano())).Perm(num)[:*shareN]
		}

		go send(groupid, photos, randlist, f)
	}
	wg.Wait()
	log.Printf("%d/%d photos share to: %s\n", *shareN, num, *groupID)
}
