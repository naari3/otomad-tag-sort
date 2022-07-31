package nicovideo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type TagCacher struct {
	cacheFile       string
	missedCacheFile string
	cache           map[string]string
	missedCache     map[string]interface{}
}

func NewTagCacher(cacheFile, missedCacheFile string) (*TagCacher, error) {
	file, err := os.OpenFile(cacheFile, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var cache map[string]string
	json.Unmarshal(raw, &cache)
	if cache == nil {
		cache = make(map[string]string)
	}

	file2, err := os.OpenFile(missedCacheFile, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	raw2, err := ioutil.ReadAll(file2)
	if err != nil {
		return nil, err
	}
	var missedCache map[string]interface{}
	json.Unmarshal(raw2, &missedCache)
	if missedCache == nil {
		missedCache = make(map[string]interface{})
	}
	return &TagCacher{
		cacheFile:       cacheFile,
		missedCacheFile: missedCacheFile,
		cache:           cache,
		missedCache:     missedCache,
	}, nil
}

func (tc *TagCacher) FetchJSONLDir(root string) error {
	videos, err := ReadAllVideoFromDirectory(root)
	if err != nil {
		return err
	}
	for _, video := range videos {
		tags := strings.Split(video.Tags, " ")
		tc.Save(tags)
	}
	bytes, err := json.Marshal(tc.cache)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tc.cacheFile, bytes, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (tc *TagCacher) Fetch(tag string) (string, error) {
	key := Normalize(tag)
	fetched := tc.cache[key]
	if fetched == "" {
		// fmt.Println("missed", key)
		if _, ok := tc.missedCache[key]; ok {
			// fmt.Println("already missed", key)
			return tag, nil
		}
		// fmt.Println("missed 2", key)

		fmt.Println("Fetching:", key)
		fetchedVideo, err := GetSSVideoFromTagLimit(key)
		time.Sleep(time.Millisecond * 150)
		if err != nil {
			return "", err
		}
		if len(fetchedVideo) == 0 {
			fmt.Println("there is no", key)
			tc.missedCache[key] = nil
			return tag, nil
		}
		video := fetchedVideo[0]
		newTags := strings.Split(video.Tags, " ")
		tc.Save(newTags)
		for _, newTag := range newTags {
			// fmt.Println(Normalize(newTag), "==", key, Normalize(newTag) == key)
			if Normalize(newTag) == key {
				return newTag, nil
			}
		}

		fmt.Println("all missed", key)
		tc.missedCache[key] = nil
		return tag, nil
	}
	return fetched, nil
}

func (tc *TagCacher) Save(tags []string) {
	for _, tag := range tags {
		if _, ok := tc.cache[Normalize(tag)]; ok {
			continue
		}
		tc.cache[Normalize(tag)] = tag
		fmt.Println("cached", Normalize(tag), "=", tag)
	}
}

func (tc *TagCacher) SaveToFile() error {
	bytes, err := json.MarshalIndent(tc.cache, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tc.cacheFile, bytes, 0666)
	if err != nil {
		return err
	}

	missedBytes, err := json.MarshalIndent(tc.missedCache, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tc.missedCacheFile, missedBytes, 0666)
	if err != nil {
		return err
	}

	return nil
}
