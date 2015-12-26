package flickr

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"sync"

	"github.com/toomore/lazyflickrgo/jsonstruct"
	"github.com/toomore/lazyflickrgo/utils"
)

var wg sync.WaitGroup

func readPhotosSerch(f Flickr, args map[string]string) jsonstruct.PhotosSearch {
	defer wg.Done()
	resp := f.HTTPGet(utils.APIURL, args)
	jsonData, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var data jsonstruct.PhotosSearch
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Println(err)
	}
	return data
}

// PhotosSearch search photos.
//
// https://www.flickr.com/services/api/flickr.photos.search.html
func (f Flickr) PhotosSearch(Args map[string]string) []jsonstruct.PhotosSearch {
	Args["method"] = "flickr.photos.search"
	Args["per_page"] = "500"

	wg.Add(1)
	data := readPhotosSerch(f, Args)

	if data.Photos.Pages > 1 {
		result := make([]jsonstruct.PhotosSearch, data.Photos.Pages)
		result[0] = data

		wg.Add(data.Photos.Pages - 1)
		go func() {
			for i := 2; i <= data.Photos.Pages; i++ {
				go func(i int, Args map[string]string) {
					args := make(map[string]string)
					for k, v := range Args {
						args[k] = v
					}
					args["page"] = strconv.Itoa(i)
					result[i-1] = readPhotosSerch(f, args)
				}(i, Args)
			}
		}()
		wg.Wait()
		return result
	}
	return []jsonstruct.PhotosSearch{data}
}

// PhotosGetInfo get photo info.
//
// https://www.flickr.com/services/api/flickr.photos.getInfo.html
func (f Flickr) PhotosGetInfo(photoID string) jsonstruct.PhotosGetInfo {
	Args := make(map[string]string)
	Args["method"] = "flickr.photos.getInfo"
	Args["photo_id"] = photoID

	resp := f.HTTPGet(utils.APIURL, Args)
	jsonData, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var data jsonstruct.PhotosGetInfo
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Println(err)
	}
	return data
}
