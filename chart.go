package main

import (
	"fmt"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func createBarChart(countMap map[string]int, sortedTags []string) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "音MAD タグ分布"}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    true,
			Trigger: "axis",
			AxisPointer: &opts.AxisPointer{
				Type: "shadow",
			},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:     "inside",
			Start:    0,
			End:      1,
			Throttle: 0,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "650px",
		}),
	)

	bar.AddJSFuncs(fmt.Sprintf(`
	const chartDiv = document.querySelector("div.container > div.item");
	chartDiv.style.width = "100%s";
	chartDiv.style.height = null;
	chartDiv.style.minHeight = "650px";
	window.addEventListener("resize", function() {
		goecharts_%s.resize();
	});
	`, "%", bar.ChartID))

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
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "650px",
		}),
	)

	wc.AddJSFuncs(fmt.Sprintf(`
	const chartDiv = document.querySelector("div.container > div.item");
	chartDiv.style.width = "100%s";
	chartDiv.style.height = null;
	chartDiv.style.minHeight = "650px";
	window.addEventListener("resize", function() {
		goecharts_%s.resize();
	});
	`, "%", wc.ChartID))

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

func createPieChart(countMap map[string]int, sortedTags []string, maxItemCount int) *charts.Pie {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "音MAD タグ分布"}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "1200px",
		}),
	)

	pie.AddJSFuncs(fmt.Sprintf(`
	const chartDiv = document.querySelector("div.container > div.item");
	chartDiv.style.width = "100%s";
	chartDiv.style.height = null;
	chartDiv.style.minHeight = "1200px";
	window.addEventListener("resize", function() {
		goecharts_%s.resize();
	});
	`, "%", pie.ChartID))

	items := make([]opts.PieData, 0)
	otherCount := 0
	currentCount := 0

	for _, tag := range sortedTags {
		if currentCount > maxItemCount {
			otherCount++
			continue
		}
		items = append(items, opts.PieData{Name: tag, Value: countMap[tag]})
		currentCount++
	}
	items = append(items, opts.PieData{Name: "", Value: otherCount})

	pie.AddSeries("タグ", items).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
		)

	return pie
}
