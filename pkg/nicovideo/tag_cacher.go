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
	fetched := tc.cache[tag]
	if fetched == "" {
		if _, ok := tc.missedCache[tag]; ok {
			return tag, nil
		}

		fmt.Println("Fetching:", tag)
		fetchedVideo, err := GetSSVideoFromTagLimit(tag)
		time.Sleep(time.Millisecond * 150)
		if err != nil {
			return "", err
		}
		if len(fetchedVideo) == 0 {
			tc.missedCache[tag] = nil
			return tag, nil
		}
		for _, video := range fetchedVideo {
			newTags := strings.Split(video.Tags, " ")
			tc.Save(newTags)
			for _, newTag := range newTags {
				if Normalize(newTag) == tag {
					return newTag, nil
				}
			}
		}
	}
	return fetched, nil
}

func (tc *TagCacher) Save(tags []string) {
	for _, tag := range tags {
		if _, ok := tc.cache[tag]; ok {
			continue
		}
		tc.cache[Normalize(tag)] = tag
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
