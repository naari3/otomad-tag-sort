package chart

import (
	"bytes"
	"fmt"
	"io"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	tpls "github.com/go-echarts/go-echarts/v2/templates"
)

var HeaderTpl = `
{{ define "header" }}
<head>
    <meta charset="utf-8">
    <title>{{ .PageTitle }}</title>
	<style>body {margin: 0;}</style>
{{- range .JSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CustomizedJSAssets.Values }}
    <script src="{{ . }}"></script>
{{- end }}
{{- range .CSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
{{- range .CustomizedCSSAssets.Values }}
    <link href="{{ . }}" rel="stylesheet">
{{- end }}
</head>
{{ end }}
`

var ChartTpl = `
{{- define "chart" }}
<!DOCTYPE html>
<html>
    {{- template "header" . }}
<body>
    {{- template "base" . }}
<style>
    .container {display: flex;justify-content: center;align-items: center;}
    .item {margin: auto;}
</style>
</body>
</html>
{{ end }}
`

type noMarginRender struct {
	c      interface{}
	before []func()
}

func NewNoMarginRender(c interface{}, before ...func()) render.Renderer {
	return &noMarginRender{c: c, before: before}
}

func (r *noMarginRender) Render(w io.Writer) error {
	for _, fn := range r.before {
		fn()
	}

	contents := []string{HeaderTpl, tpls.BaseTpl, ChartTpl}
	tpl := render.MustTemplate("chart", contents)

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "chart", r.c); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	return err
}

func CreateBarChart(countMap map[string]int, sortedTags []string) *charts.Bar {
	bar := charts.NewBar()
	bar.Renderer = NewNoMarginRender(bar, bar.Validate)
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
			Height: "100vh",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			AxisLabel: &opts.AxisLabel{
				Rotate: -90,
			},
		}),
	)

	bar.AddJSFuncs(fmt.Sprintf(`
	let option = option_%s;
	let echart = goecharts_%s;
	const chartDiv = document.querySelector("div.container > div.item");
	option.grid = {
		containLabel: true,
	};
	echart.setOption(option);
	window.addEventListener("resize", function() {
		echart.resize();
	});
	echart.on("click", (params) => {
		if (params.componentType === "series") {
			const tag = encodeURIComponent(params.name);
			const url = "https://www.nicovideo.jp/tag/%s+"+tag+"?sort=h&order=d";
			window.open(url, '_blank');
		}
	});
	`, bar.ChartID, bar.ChartID, "%E9%9F%B3MAD"))

	items := make([]opts.BarData, 0)
	for _, tag := range sortedTags {
		items = append(items, opts.BarData{Value: countMap[tag]})
	}

	bar.SetXAxis(sortedTags).
		AddSeries("タグ", items)

	return bar
}

func CreateWordCloud(countMap map[string]int, sortedTags []string, min int) *charts.WordCloud {
	wc := charts.NewWordCloud()
	wc.Renderer = NewNoMarginRender(wc, wc.Validate)
	wc.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "音MAD タグ分布"}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "100vh",
		}),
	)

	wc.AddJSFuncs(fmt.Sprintf(`
	let option = option_%s;
	let echart = goecharts_%s;
	const chartDiv = document.querySelector("div.container > div.item");
	option.grid = {
		containLabel: true,
	};
	echart.setOption(option);
	window.addEventListener("resize", function() {
		echart.resize();
	});
	echart.on("click", (params) => {
		if (params.componentType === "series") {
			const tag = encodeURIComponent(params.name);
			const url = "https://www.nicovideo.jp/tag/%s+"+tag+"?sort=h&order=d";
			window.open(url, '_blank');
		}
	});
	`, wc.ChartID, wc.ChartID, "%E9%9F%B3MAD"))

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
					SizeRange: []float32{10, 80},
				},
			),
		)

	return wc
}

func CreatePieChart(countMap map[string]int, sortedTags []string, maxItemCount int) *charts.Pie {
	pie := charts.NewPie()
	pie.Renderer = NewNoMarginRender(pie, pie.Validate)
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "音MAD タグ分布"}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    true,
			Trigger: "item",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "100vh",
		}),
	)

	pie.AddJSFuncs(fmt.Sprintf(`
	let option = option_%s;
	let echart = goecharts_%s;
	const chartDiv = document.querySelector("div.container > div.item");
	option.grid = {
		containLabel: true,
	};
	echart.setOption(option);
	window.addEventListener("resize", function() {
		echart.resize();
	});
	echart.on("click", (params) => {
		if (params.componentType === "series") {
			const tag = encodeURIComponent(params.name);
			const url = "https://www.nicovideo.jp/tag/%s+"+tag+"?sort=h&order=d";
			window.open(url, '_blank');
		}
	});
	`, pie.ChartID, pie.ChartID, "%E9%9F%B3MAD"))

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
