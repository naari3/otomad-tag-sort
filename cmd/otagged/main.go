package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/naari3/otagged/pkg/chart"
	"github.com/naari3/otagged/pkg/nicovideo"
)

func createCharts(videos []nicovideo.Video) error {
	fmt.Println("Start")
	countMap, err := nicovideo.GetCountGroupByOtomadTag(videos, nil)
	if err != nil {
		panic(err)
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

	bar := chart.CreateBarChart(countMap, sortedTags)
	fbar, err := os.Create("html/bar.html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := chart.CreateWordCloud(countMap, sortedTags, 5)
	fwc, err := os.Create("html/wc.html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	pie := chart.CreatePieChart(countMap, sortedTags, 75)
	fpie, err := os.Create("html/pie.html")
	if err != nil {
		return err
	}
	pie.Render(fpie)

	fmt.Println("End")

	return nil
}

func createChartsForYear(videos []nicovideo.Video, year int) error {
	fmt.Println("Start year:", year)
	err := os.MkdirAll("html/"+strconv.Itoa(year), os.ModePerm)
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

	bar := chart.CreateBarChart(countMap, sortedTags)
	bar.Title = opts.Title{Title: "音MAD タグ分布 " + strconv.Itoa(year)}
	fbar, err := os.Create("html/" + strconv.Itoa(year) + "/bar.html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := chart.CreateWordCloud(countMap, sortedTags, 5)
	wc.Title = opts.Title{Title: "音MAD タグ分布 " + strconv.Itoa(year)}
	fwc, err := os.Create("html/" + strconv.Itoa(year) + "/wc.html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	pie := chart.CreatePieChart(countMap, sortedTags, 75)
	pie.Title = opts.Title{Title: "音MAD タグ分布 " + strconv.Itoa(year)}
	fpie, err := os.Create("html/" + strconv.Itoa(year) + "/pie.html")
	if err != nil {
		return err
	}
	pie.Render(fpie)

	fmt.Println("End year:", year)

	return nil
}

func main() {
	fmt.Println("Start")
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

	err = createCharts(videos)
	if err != nil {
		panic(err)
	}

	for year := 2007; year <= time.Now().Year(); year++ {
		err = createChartsForYear(videos, year)
		if err != nil {
			panic(err)
		}
	}
}
