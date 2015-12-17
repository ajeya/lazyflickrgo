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
	userID  = flag.String("userid", "", "User number ID")
	albumID = flag.String("albumid", "", "Album/Set number ID")
	groupID = flag.String("groupid", "", "Group number ID")
	apikey  = flag.String("apikey", os.Getenv("FLICKRAPIKEY"), "Flickr API Key")
	secret  = flag.String("secret", os.Getenv("FLICKRSECRET"), "Flickr secret")
	shareN  = flag.Int("n", 6, "Per share num")
	tags    = flag.String("tags", "", "Search tags, ',' for split more")
	info    = color.New(color.Bold, color.FgGreen).SprintfFunc()
	warn    = color.New(color.Bold, color.FgRed).SprintfFunc()
	wg      sync.WaitGroup
	photos  []jsonstruct.Photo
)

func fromSets(f *flickr.Flickr) []jsonstruct.Photo {
	var result []jsonstruct.Photo
	for _, albumid := range strings.Split(*albumID, ",") {
		albumdata := f.PhotosetsGetPhotos(albumid, *userID)
		result = append(result, albumdata.Photoset.Photos.Photo...)
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

	return searchResult.Photos.Photo
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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if num <= *shareN {
		*shareN = num
	}

	wg.Add(*shareN)

	for _, val := range r.Perm(num)[:*shareN] {
		photo := photos[val]
		log.Println(info("Pick up photo: %d [%s] %+v", val, photo.ID, photo))
		go func(photo jsonstruct.Photo, groupID *string, val int) {
			runtime.Gosched()
			resp := f.GroupsPoolsAdd(*groupID, photo.ID)
			if resp.Stat == "ok" {
				log.Println(info("%s %s", photo.ID, photo.Title))
			} else {
				log.Println(warn("%s(%d) %s %s", resp.Message, resp.Code, photo.ID, photo.Title))
			}
			wg.Done()
		}(photo, groupID, val)
	}
	wg.Wait()
	log.Printf("%d/%d photos share to: %s\n", *shareN, num, *groupID)
}
