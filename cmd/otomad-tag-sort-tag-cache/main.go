package main

import (
	"fmt"

	"github.com/naari3/otomad-tag-sort/pkg/nicovideo"
)

func main() {
	videos, err := nicovideo.ReadAllVideoFromDirectory("jsonl")
	if err != nil {
		panic(err)
	}
	fmt.Println("Tag cacher start")
	tc, err := nicovideo.NewTagCacher("cache_tags.json", "missed_cache_tags.json")
	if err != nil {
		panic(err)
	}
	// err = tc.FetchJSONLDir("jsonl_append")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Tag cacher collected jsonl_append")

	countMap, err := nicovideo.GetCountGroupByOtomadTag(videos, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Collected otomad tags:", len(countMap))

	i := 0
	for tag := range countMap {
		fmt.Println(i, tag)
		_, err := tc.Fetch(tag)
		if err != nil {
			panic(err)
		}
		i++
		if i%1000 == 0 {
			err = tc.SaveToFile()
			if err != nil {
				panic(err)
			}
		}
	}
}
