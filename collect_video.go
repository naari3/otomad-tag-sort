package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// {
// 	"video_id": "sm3000",
// 	"watch_num": 14614,
// 	"comment_num": "167",
// 	"mylist_num": 76,
// 	"title": "【MAD】最終兵器彼女",
// 	"description": "最終兵器彼女・登場キャラの死亡集を寄せ集めたようなＭＡＤ　少しばかり酷ですが、内容知ってる人は感動もできます。※オリ曲上げてくれってリクがあったが、別にあんたのために上げたんじゃないんだからね!!!オリVer＝sm1866679　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　よく見たらキリ番動画でしたｗ　　　　巡礼者の方々、乙であります！(`・ω・)ゞ",
// 	"category": "anime",
// 	"tags": "MAD アニメ 最終兵器彼女",
// 	"upload_time": "2007-03-06T12:14:17+09:00",
// 	"file_type": "flv",
// 	"length": 231,
// 	"size_high": 16558887,
// 	"size_low": 12039870
// }

type Video struct {
	// VideoID     string `json:"video_id"`
	// WatchNum    int    `json:"watch_num"`
	// CommentNum  string `json:"comment_num"`
	// MylistNum   int    `json:"mylist_num"`
	// Title       string `json:"title"`
	// Description string `json:"description"`
	// Category    string `json:"category"`
	Tags       string    `json:"tags"`
	UploadTime time.Time `json:"upload_time"`
	// FileType    string `json:"file_type"`
	// Length      int    `json:"length"`
	// SizeHigh    int    `json:"size_high"`
	// SizeLow     int    `json:"size_low"`
}

func (v *Video) IsOtomad() bool {
	return strings.Contains(v.Tags, "音MAD") || strings.Contains(v.Tags, "音mad")
}

func (v *Video) NormalizedTags() []string {
	return strings.Split(strings.ToLower(v.Tags), " ")
}

func findJSONL(root string) []string {
	var files []string
	filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".jsonl" {
			files = append(files, path)
		}
		return nil
	})

	return files
}

func readVideoFromJSONL(file string) ([]Video, error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var videos []Video

	decoder := json.NewDecoder(f)
	for {
		var v Video
		if err := decoder.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		videos = append(videos, v)
	}
	if err != nil {
		return nil, err
	}

	return videos, nil
}

func readAllVideoFromDirectory(root string) ([]Video, error) {
	var videos []Video

	// ctx := context.Background()

	eg := errgroup.Group{}
	mutex := sync.Mutex{}
	// sem := semaphore.NewWeighted(3) // 最大数を3に設定

	for _, file := range findJSONL(root) {
		file := file
		// sem.Acquire(ctx, 1)
		eg.Go(func() error {
			fmt.Println(file)
			vs, err := readVideoFromJSONL(file)
			if err != nil {
				return err
			}
			mutex.Lock()
			videos = append(videos, vs...)
			mutex.Unlock()
			// sem.Release(1)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return videos, nil
}

func getCountGroupByOtomadTag(videos []Video, filter func(video Video) bool) (map[string]int, error) {
	counts := make(map[string]int)

	eg := errgroup.Group{}
	mutex := sync.Mutex{}
	for _, v := range videos {
		v := v
		eg.Go(func() error {
			if filter != nil && !filter(v) {
				return nil
			}
			tags := v.NormalizedTags()
			if !v.IsOtomad() {
				return nil
			}

			for _, tag := range tags {
				if tag == "音MAD" || tag == "音mad" {
					return nil
				}
				mutex.Lock()
				counts[tag]++
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return counts, nil
}
