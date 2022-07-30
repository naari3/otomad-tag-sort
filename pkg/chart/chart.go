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

type myOwnRender struct {
	c      interface{}
	before []func()
}

func NewMyOwnRender(c interface{}, before ...func()) render.Renderer {
	return &myOwnRender{c: c, before: before}
}

func (r *myOwnRender) Render(w io.Writer) error {
	for _, fn := range r.before {
		fn()
	}

	contents := []string{HeaderTpl, tpls.BaseTpl, tpls.ChartTpl}
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
			Height: "800px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			AxisLabel: &opts.AxisLabel{
				Rotate: -90,
			},
		}),
	)

	bar.AddJSFuncs(fmt.Sprintf(`
	const chartDiv = document.querySelector("div.container > div.item");
	chartDiv.style.width = "100%s";
	chartDiv.style.height = null;
	chartDiv.style.minHeight = "800px";
	option_%s.grid = {
		containLabel: true,
	};
	goecharts_%s.setOption(option_%s);
	window.addEventListener("resize", function() {
		goecharts_%s.resize();
	});
	`, "%", bar.ChartID, bar.ChartID, bar.ChartID, bar.ChartID))

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

func CreatePieChart(countMap map[string]int, sortedTags []string, maxItemCount int) *charts.Pie {
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
