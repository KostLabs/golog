package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type metrics struct {
	ns     float64
	bytes  float64
	allocs float64
}

var loggerOrder = []string{"Golog", "Slog", "Zerolog", "Zap", "Apex", "Logrus"}
var scenarioOrder = []string{"Simple", "WithFields", "WithLargeFields", "WithExtraLargeFields"}

func main() {
	results, raw, err := runBenchmarks()
	if err != nil {
		fmt.Fprintln(os.Stderr, raw)
		fmt.Fprintf(os.Stderr, "benchviz: %v\n", err)
		os.Exit(1)
	}

	html, err := buildHTML(results)
	if err != nil {
		fmt.Fprintf(os.Stderr, "benchviz: build html: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join("docs", "index.html")
	if err := os.WriteFile(outPath, []byte(html), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "benchviz: write %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Updated %s from latest benchmark run.\n", outPath)
}

func runBenchmarks() (map[string]map[string]metrics, string, error) {
	cmd := exec.Command("go", "test", "-bench", "^BenchmarkAllLoggers", "-benchmem", "-run", "^$", "-count=1")
	cmd.Dir = filepath.Join("benchmarks")
	out, err := cmd.CombinedOutput()
	raw := string(out)
	if err != nil {
		return nil, raw, fmt.Errorf("benchmark command failed: %w", err)
	}

	results := make(map[string]map[string]metrics)
	lineRE := regexp.MustCompile(`^BenchmarkAllLoggers([A-Za-z]+)/(Golog|Slog|Zerolog|Zap|Apex|Logrus)-\d+\s+\d+\s+([0-9.]+)\s+ns/op\s+([0-9.]+)\s+B/op\s+([0-9.]+)\s+allocs/op$`)

	for _, line := range strings.Split(raw, "\n") {
		m := lineRE.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) != 6 {
			continue
		}
		scenario := m[1]
		logger := m[2]
		ns, _ := strconv.ParseFloat(m[3], 64)
		b, _ := strconv.ParseFloat(m[4], 64)
		a, _ := strconv.ParseFloat(m[5], 64)

		if results[scenario] == nil {
			results[scenario] = make(map[string]metrics)
		}
		results[scenario][logger] = metrics{ns: ns, bytes: b, allocs: a}
	}

	for _, scenario := range scenarioOrder {
		if results[scenario] == nil {
			return nil, raw, fmt.Errorf("missing scenario %q in benchmark output", scenario)
		}
		for _, logger := range loggerOrder {
			if _, ok := results[scenario][logger]; !ok {
				return nil, raw, fmt.Errorf("missing logger %q for scenario %q in benchmark output", logger, scenario)
			}
		}
	}

	return results, raw, nil
}

func buildHTML(results map[string]map[string]metrics) (string, error) {
	cpuOption, err := buildOption(
		"Logger Performance - Execution Time",
		"Lower is better (nanoseconds per operation)",
		"Time (ns/op)",
		results,
		func(m metrics) float64 { return m.ns },
	)
	if err != nil {
		return "", err
	}
	memOption, err := buildOption(
		"Logger Performance - Memory Usage",
		"Lower is better (bytes per operation)",
		"Memory (B/op)",
		results,
		func(m metrics) float64 { return m.bytes },
	)
	if err != nil {
		return "", err
	}
	allocOption, err := buildOption(
		"Logger Performance - Allocations",
		"Lower is better (allocations per operation)",
		"Allocations (allocs/op)",
		results,
		func(m metrics) float64 { return m.allocs },
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Logger Benchmark Visualization</title>
  <script src="https://go-echarts.github.io/go-echarts-assets/assets/echarts.min.js"></script>
  <script src="https://go-echarts.github.io/go-echarts-assets/assets/themes/westeros.js"></script>
</head>
<body>
  <style>
    .container { display: flex; justify-content: center; align-items: center; }
    .item { margin: auto; }
  </style>

  <div class="container"><div class="item" id="cpu" style="width:1200px;height:700px;"></div></div>
  <div class="container"><div class="item" id="mem" style="width:1200px;height:700px;"></div></div>
  <div class="container"><div class="item" id="alloc" style="width:1200px;height:700px;"></div></div>

  <script>
    "use strict";
    const cpu = echarts.init(document.getElementById("cpu"), "westeros", {renderer: "canvas"});
    const mem = echarts.init(document.getElementById("mem"), "westeros", {renderer: "canvas"});
    const alloc = echarts.init(document.getElementById("alloc"), "westeros", {renderer: "canvas"});

    cpu.setOption(%s);
    mem.setOption(%s);
    alloc.setOption(%s);
  </script>
</body>
</html>
`, cpuOption, memOption, allocOption), nil
}

func buildOption(
	title string,
	subtitle string,
	yAxisName string,
	results map[string]map[string]metrics,
	metricSelector func(metrics) float64,
) (string, error) {
	type series struct {
		Name string           `json:"name"`
		Type string           `json:"type"`
		Data []map[string]any `json:"data"`
	}

	seriesValues := make([]series, 0, len(scenarioOrder))
	for _, scenario := range scenarioOrder {
		data := make([]map[string]any, 0, len(loggerOrder))
		for _, logger := range loggerOrder {
			data = append(data, map[string]any{"value": metricSelector(results[scenario][logger])})
		}
		seriesValues = append(seriesValues, series{Name: scenario, Type: "bar", Data: data})
	}

	option := map[string]any{
		"grid":   []map[string]any{{"left": "80px", "top": "120px", "right": "80px", "bottom": "80px"}},
		"legend": map[string]any{"show": true, "top": "80px"},
		"series": seriesValues,
		"title":  map[string]any{"text": title, "subtext": subtitle, "top": "20px"},
		"toolbox": map[string]any{},
		"tooltip": map[string]any{"show": true},
		"xAxis":   []map[string]any{{"name": "Loggers", "data": loggerOrder}},
		"yAxis":   []map[string]any{{"name": yAxisName}},
	}

	b, err := json.Marshal(option)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

