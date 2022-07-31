package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/naari3/otomad-tag-sort/pkg/chart"
	"github.com/naari3/otomad-tag-sort/pkg/nicovideo"
)

func createCharts(videos []nicovideo.Video, tc *nicovideo.TagCacher) error {
	fmt.Println("Start")
	countMap, err := nicovideo.GetCountGroupByOtomadTag(videos, nil)
	if err != nil {
		return err
	}
	fmt.Println("Collected otomad tags:", len(countMap))

	sortedTags := make([]string, 0, len(countMap))
	for tag := range countMap {
		sortedTags = append(sortedTags, tag)
	}
	sort.SliceStable(sortedTags, func(i, j int) bool {
		return countMap[sortedTags[i]] > countMap[sortedTags[j]]
	})
	fmt.Println("Sorted By counts")

	bar := chart.CreateBarChart(countMap, sortedTags, tc)
	fbar, err := os.Create("docs/bar.html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := chart.CreateWordCloud(countMap, sortedTags, 5, tc)
	fwc, err := os.Create("docs/wc.html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	pie := chart.CreatePieChart(countMap, sortedTags, 75, tc)
	fpie, err := os.Create("docs/pie.html")
	if err != nil {
		return err
	}
	pie.Render(fpie)

	err = tc.SaveToFile()
	if err != nil {
		return err
	}
	fmt.Println("End")

	return nil
}

func createChartsForYear(videos []nicovideo.Video, year int, tc *nicovideo.TagCacher) error {
	fmt.Println("Start year:", year)
	err := os.MkdirAll("docs/"+strconv.Itoa(year), os.ModePerm)
	if err != nil {
		return err
	}

	countMap, err := nicovideo.GetCountGroupByOtomadTag(videos, func(video nicovideo.Video) bool {
		return video.UploadTime.Year() == year
	})
	if err != nil {
		return err
	}
	fmt.Println("Collected otomad tags:", len(countMap), "for year:", year)

	sortedTags := make([]string, 0, len(countMap))
	for tag := range countMap {
		sortedTags = append(sortedTags, tag)
	}
	sort.SliceStable(sortedTags, func(i, j int) bool {
		return countMap[sortedTags[i]] > countMap[sortedTags[j]]
	})
	fmt.Println("Sorted By counts")

	bar := chart.CreateBarChart(countMap, sortedTags, tc)
	bar.Title = opts.Title{Title: "音MAD タグ分布 " + strconv.Itoa(year)}
	fbar, err := os.Create("docs/" + strconv.Itoa(year) + "/bar.html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := chart.CreateWordCloud(countMap, sortedTags, 5, tc)
	wc.Title = opts.Title{Title: "音MAD タグ分布 " + strconv.Itoa(year)}
	fwc, err := os.Create("docs/" + strconv.Itoa(year) + "/wc.html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	pie := chart.CreatePieChart(countMap, sortedTags, 75, tc)
	pie.Title = opts.Title{Title: "音MAD タグ分布 " + strconv.Itoa(year)}
	fpie, err := os.Create("docs/" + strconv.Itoa(year) + "/pie.html")
	if err != nil {
		return err
	}
	pie.Render(fpie)

	err = tc.SaveToFile()
	if err != nil {
		return err
	}
	fmt.Println("End year:", year)

	return nil
}

func main() {
	videos, err := nicovideo.ReadAllVideoFromDirectory("jsonl")
	if err != nil {
		panic(err)
	}
	fmt.Println("Collected all videos:", len(videos))

	append_videos, err := nicovideo.ReadAllVideoFromDirectory("jsonl_append")
	if err != nil {
		panic(err)
	}
	fmt.Println("Collected all append_videos:", len(append_videos))

	videos = append(videos, append_videos...)

	tc, err := nicovideo.NewTagCacher("cache_tags.json", "missed_cache_tags.json")
	if err != nil {
		panic(err)
	}
	// a, err := tc.Fetch("かわいい!")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(a)
	// tc.SaveToFile()

	err = createCharts(videos, tc)
	if err != nil {
		panic(err)
	}

	for year := 2007; year <= time.Now().Year(); year++ {
		err = createChartsForYear(videos, year, tc)
		if err != nil {
			panic(err)
		}
	}
}
