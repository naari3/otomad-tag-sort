package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/naari3/otagged/pkg/nicovideo"
)

func main() {
	videos, err := nicovideo.ReadLastVideoFromDirectory("jsonl")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(videos))
	append_videos, err := nicovideo.ReadLastVideoFromDirectory("jsonl_append")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(append_videos))
	videos = append(videos, append_videos...)
	fmt.Println(len(videos))
	sort.SliceStable(videos, func(i, j int) bool {
		a, err := videos[i].GetIDNum()
		if err != nil {
			panic(err)
		}
		b, err := videos[j].GetIDNum()
		if err != nil {
			panic(err)
		}
		return a < b
	})
	lastVideo := videos[len(videos)-1]
	// lastVideo := nicovideo.Video{
	// 	VideoID: "sm40819401",
	// }

	lastSSVideo, err := nicovideo.GetLastSSVideo()
	if err != nil {
		panic(err)
	}

	fetchedLastID, err := lastVideo.GetIDNum()
	if err != nil {
		panic(err)
	}
	ssVideoLastID, err := lastSSVideo.GetIDNum()
	if err != nil {
		panic(err)
	}
	if ssVideoLastID <= fetchedLastID {
		fmt.Println("No new video")
		return
	}

	fmt.Println("from", fetchedLastID, "to", ssVideoLastID)

	// ssVideos := make([]nicovideo.SSVideo, 0)
	targetIDs := make([]string, 0)
	idCollector, err := nicovideo.NewIDCollector(fetchedLastID)
	if err != nil {
		panic(err)
	}
	finish := false
	for idCollector.HasNext() {
		for len(targetIDs) < 35 && idCollector.HasNext() {
			id := idCollector.Next()
			idNum, err := strconv.Atoi(id[2:])
			if err != nil {
				panic(err)
			}
			if idNum > ssVideoLastID {
				// fmt.Println(id, ">", ssVideoLastID)
				finish = true
				break
			}
			targetIDs = append(targetIDs, id)
		}
		// fmt.Println("Fetching", len(targetIDs))
		videos, err := nicovideo.GetSSVideoFromIDs(targetIDs)
		targetIDs = make([]string, 0)
		if err != nil {
			panic(err)
		}
		// ssVideos = append(ssVideos, videos...)
		sort.SliceStable(videos, func(i, j int) bool {
			a, err := videos[i].GetIDNum()
			if err != nil {
				panic(err)
			}
			b, err := videos[j].GetIDNum()
			if err != nil {
				panic(err)
			}
			return a < b
		})
		for _, video := range videos {
			full := video.ToVideo()
			fmt.Println(full.VideoID, full.Title)
			err := full.SaveToDirectory("jsonl_append")
			if err != nil {
				panic(err)
			}
		}
		if finish {
			break
		}
	}
	// fmt.Println("Fetched", len(ssVideos))

	// sort.SliceStable(ssVideos, func(i, j int) bool {
	// 	a, err := ssVideos[i].GetIDNum()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	b, err := ssVideos[j].GetIDNum()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return a < b
	// })
}
