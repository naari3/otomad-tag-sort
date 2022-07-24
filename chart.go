package main

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func createBarChart(countMap map[string]int, sortedTags []string) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "音MAD タグ分布"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:     "slider",
			Start:    0,
			End:      1,
			Throttle: 0,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1200px",
			Height: "600px",
		}),
	)

	items := make([]opts.BarData, 0)
	for _, tag := range sortedTags {
		items = append(items, opts.BarData{Value: countMap[tag]})
	}

	bar.SetXAxis(sortedTags).
		AddSeries("タグ", items)

	return bar
}

func createWordCloud(countMap map[string]int, sortedTags []string, min int) *charts.WordCloud {
	wc := charts.NewWordCloud()
	wc.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "音MAD タグ分布"}),
	)

	items := make([]opts.WordCloudData, 0)
	for _, tag := range sortedTags {
		if countMap[tag] < min {
			break
		}
		items = append(items, opts.WordCloudData{Name: tag, Value: countMap[tag]})
	}

	wc.AddSeries("タグ", items).
		SetSeriesOptions(
			charts.WithWorldCloudChartOpts(
				opts.WordCloudChart{
					SizeRange: []float32{14, 80},
				},
			),
		)

	return wc
}

func createPieChart(countMap map[string]int, sortedTags []string, min int) *charts.Pie {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "音MAD タグ分布"}),
	)

	items := make([]opts.PieData, 0)
	for _, tag := range sortedTags {
		if countMap[tag] < min {
			break
		}
		items = append(items, opts.PieData{Name: tag, Value: countMap[tag]})
	}

	pie.AddSeries("タグ", items)
	// .SetSeriesOptions(
	// 	charts.WithPieChartOpts(
	// 		opts.PieChart{
	// 		},
	// 	),
	// )

	return pie
}
