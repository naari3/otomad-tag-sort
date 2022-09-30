package main

import (
	"fmt"
	"os"
	"runtime/pprof"
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

	countMap, err := nicovideo.GetCountGroupByOtomadTag(videos, func(video nicovideo.Video) bool { return video.UploadTime.Year() == year })
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
	bar.Title = opts.Title{Title: "音MAD タグ頒布 " + strconv.Itoa(year)}
	fbar, err := os.Create("docs/" + strconv.Itoa(year) + "/bar.html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := chart.CreateWordCloud(countMap, sortedTags, 5, tc)
	wc.Title = opts.Title{Title: "音MAD タグ頒布 " + strconv.Itoa(year)}
	fwc, err := os.Create("docs/" + strconv.Itoa(year) + "/wc.html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	pie := chart.CreatePieChart(countMap, sortedTags, 75, tc)
	pie.Title = opts.Title{Title: "音MAD タグ頒布 " + strconv.Itoa(year)}
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

func createMICharts(videos []nicovideo.Video, tc *nicovideo.TagCacher, limit int, allLimit int) error {
	fmt.Println("Start")
	countMap, err := nicovideo.GetCountGroupByOtomadTag(videos, func(video nicovideo.Video) bool { return true })
	if err != nil {
		return err
	}
	fmt.Println("Collected otomad tags:", len(countMap))

	targets := make([]string, 0, len(countMap))
	for tag := range countMap {
		if countMap[tag] < limit {
			continue
		}
		targets = append(targets, tag)
	}

	allCountMap, err := nicovideo.GetCountGroupByTargetsWithDB(targets)
	if err != nil {
		return err
	}
	fmt.Println("Collected tags:", len(allCountMap))

	MIs := make(map[string]float32, 0)
	for tag := range countMap {
		if allCountMap[tag] == 0 {
			continue
		}
		if allCountMap[tag] < allLimit {
			continue
		}
		if countMap[tag] == allCountMap[tag] && countMap[tag] < 10 {
			continue
		}
		mi := float32(countMap[tag]) / float32(allCountMap[tag])
		MIs[tag] = mi
	}

	sortedTags := make([]string, 0, len(MIs))
	for tag := range MIs {
		sortedTags = append(sortedTags, tag)
	}
	sort.SliceStable(sortedTags, func(i, j int) bool {
		return MIs[sortedTags[i]] > MIs[sortedTags[j]]
	})
	fmt.Println("Sorted By MI")

	additionalInfo := ""
	if limit != 0 {
		additionalInfo += " そのタグの音MADの投稿数が" + strconv.Itoa(limit) + "件以上"
	}
	if allLimit != 0 {
		additionalInfo += " タグの投稿数が" + strconv.Itoa(allLimit) + "件以上"
	}

	additionalFileName := ""
	if limit != 0 || allLimit != 0 {
		additionalFileName += "_" + strconv.Itoa(limit)
		additionalFileName += "_" + strconv.Itoa(allLimit)
	}

	bar := chart.CreateBarChart(MIs, sortedTags, tc)
	bar.Title = opts.Title{Title: "音MAD タグ 相互情報量" + additionalInfo}
	fbar, err := os.Create("docs/mibar" + additionalFileName + ".html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := chart.CreateWordCloud(MIs, sortedTags, 0, tc)
	wc.Title = opts.Title{Title: "音MAD タグ 相互情報量" + additionalInfo}
	fwc, err := os.Create("docs/miwc" + additionalFileName + ".html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	// pie := chart.CreatePieChart(MIs, sortedTags, 75, tc)
	// pie.Title = opts.Title{Title: "音MAD タグ 相互情報量" + additionalInfo}
	// fpie, err := os.Create("docs/mipie" + additionalFileName + ".html")
	// if err != nil {
	// 	return err
	// }
	// pie.Render(fpie)

	err = tc.SaveToFile()
	if err != nil {
		return err
	}
	fmt.Println("End")

	return nil
}

func createMIChartsForYear(videos []nicovideo.Video, year int, tc *nicovideo.TagCacher, limit int, allLimit int) error {
	fmt.Println("Start")
	countMap, err := nicovideo.GetCountGroupByOtomadTag(videos, func(video nicovideo.Video) bool { return video.UploadTime.Year() == year })
	if err != nil {
		return err
	}
	fmt.Println("Collected otomad tags:", len(countMap), "for year:", year)

	targets := make([]string, 0, len(countMap))
	for tag := range countMap {
		if countMap[tag] < limit {
			continue
		}
		targets = append(targets, tag)
	}

	allCountMap, err := nicovideo.GetCountGroupByTargetsWithDBForYear(targets, year)
	if err != nil {
		return err
	}
	fmt.Println("Collected tags:", len(allCountMap), "for year:", year)

	MIs := make(map[string]float32, 0)
	for tag := range countMap {
		if allCountMap[tag] == 0 {
			continue
		}
		if allCountMap[tag] < allLimit {
			continue
		}
		if countMap[tag] == allCountMap[tag] {
			continue
		}
		mi := float32(countMap[tag]) / float32(allCountMap[tag])
		MIs[tag] = mi
	}

	sortedTags := make([]string, 0, len(MIs))
	for tag := range MIs {
		sortedTags = append(sortedTags, tag)
	}
	sort.SliceStable(sortedTags, func(i, j int) bool {
		return MIs[sortedTags[i]] > MIs[sortedTags[j]]
	})
	fmt.Println("Sorted By MI")

	additionalInfo := ""
	if year != 0 {
		additionalInfo = " " + strconv.Itoa(year)
	}
	if limit != 0 {
		additionalInfo += " そのタグの音MADの投稿数が" + strconv.Itoa(limit) + "件以上"
	}
	if allLimit != 0 {
		additionalInfo += " タグの投稿数が" + strconv.Itoa(allLimit) + "件以上"
	}
	additionalFileName := ""
	if limit != 0 || allLimit != 0 {
		additionalFileName += "_" + strconv.Itoa(limit)
		additionalFileName += "_" + strconv.Itoa(allLimit)
	}

	bar := chart.CreateBarChart(MIs, sortedTags, tc)
	bar.Title = opts.Title{Title: "音MAD タグ 相互情報量" + additionalInfo}
	fbar, err := os.Create("docs/" + strconv.Itoa(year) + "/mibar" + additionalFileName + ".html")
	if err != nil {
		return err
	}
	bar.Render(fbar)

	wc := chart.CreateWordCloud(MIs, sortedTags, 0, tc)
	wc.Title = opts.Title{Title: "音MAD タグ 相互情報量" + additionalInfo}
	fwc, err := os.Create("docs/" + strconv.Itoa(year) + "/miwc" + additionalFileName + ".html")
	if err != nil {
		return err
	}
	wc.Render(fwc)

	// pie := chart.CreatePieChart(MIs, sortedTags, 75, tc)
	// pie.Title = opts.Title{Title: "音MAD タグ 相互情報量" + additionalInfo}
	// fpie, err := os.Create("docs/" + strconv.Itoa(year) + "/mipie" + additionalFileName + ".html")
	// if err != nil {
	// 	return err
	// }
	// pie.Render(fpie)

	err = tc.SaveToFile()
	if err != nil {
		return err
	}
	fmt.Println("End")

	return nil
}

func main() {
	f, err := os.Create("cpu" + strconv.Itoa(int(time.Now().Unix())) + ".pprof")
	if err != nil {
		panic(err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

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

	err = createCharts(videos, tc)
	if err != nil {
		panic(err)
	}

	err = createMICharts(videos, tc, 0, 0)
	if err != nil {
		panic(err)
	}

	err = createMICharts(videos, tc, 100, 0)
	if err != nil {
		panic(err)
	}

	err = createMICharts(videos, tc, 0, 1000)
	if err != nil {
		panic(err)
	}

	for year := 2007; year <= time.Now().Year(); year++ {
		err = createChartsForYear(videos, year, tc)
		if err != nil {
			panic(err)
		}
		err = createMIChartsForYear(videos, year, tc, 0, 0)
		if err != nil {
			panic(err)
		}
		err = createMIChartsForYear(videos, year, tc, 100, 0)
		if err != nil {
			panic(err)
		}
		err = createMIChartsForYear(videos, year, tc, 0, 1000)
		if err != nil {
			panic(err)
		}
	}
	if err := tc.SaveToFile(); err != nil {
		panic(err)
	}
}
