package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

func createCharts(videos []Video) error {
	fmt.Println("Start")
	countMap, err := getCountGroupByOtomadTag(videos, nil)
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

	bar := createBarChart(countMap, sortedTags)
	fbar, err := os.Create("html/bar.html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := createWordCloud(countMap, sortedTags, 5)
	fwc, err := os.Create("html/wc.html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	pie := createPieChart(countMap, sortedTags, 100)
	fpie, err := os.Create("html/pie.html")
	if err != nil {
		return err
	}
	pie.Render(fpie)

	fmt.Println("End")

	return nil
}

func createChartsForYear(videos []Video, year int) error {
	fmt.Println("Start year:", year)
	err := os.Mkdir("html/"+strconv.Itoa(year), os.ModePerm)
	if err != nil {
		return err
	}

	countMap, err := getCountGroupByOtomadTag(videos, func(video Video) bool {
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

	bar := createBarChart(countMap, sortedTags)
	fbar, err := os.Create("html/" + strconv.Itoa(year) + "/bar.html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := createWordCloud(countMap, sortedTags, 5)
	fwc, err := os.Create("html/" + strconv.Itoa(year) + "/wc.html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	pie := createPieChart(countMap, sortedTags, 100)
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
	videos, err := readAllVideoFromDirectory("jsonl")
	if err != nil {
		panic(err)
	}
	fmt.Println("Collected all videos:", len(videos))

	err = createCharts(videos)
	if err != nil {
		panic(err)
	}
	for year := 2007; year <= 2021; year++ {
		err = createChartsForYear(videos, year)
		if err != nil {
			panic(err)
		}
	}
}
