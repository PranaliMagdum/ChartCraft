package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type GradeData struct {
	Headers []string
	Rows    [][]string
}

func main() {
	gradeData, err := readCSV("data.csv")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rootHandler(w, r, gradeData)
	})

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func readCSV(filename string) (GradeData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return GradeData{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return GradeData{}, err
	}

	if len(records) < 2 {
		return GradeData{}, fmt.Errorf("CSV file is empty or has insufficient data")
	}

	return GradeData{
		Headers: records[0],
		Rows:    records[1:],
	}, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request, data GradeData) {
	if r.URL.Path == "/" {
		displayOptions(w, data.Headers)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}

	chartType := parts[1]
	column, err := strconv.Atoi(parts[2])
	if err != nil || column < 0 || column >= len(data.Headers) {
		http.Error(w, "Invalid column selection", http.StatusBadRequest)
		return
	}

	renderChart(w, chartType, data, column)
}

func displayOptions(w http.ResponseWriter, headers []string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Select a column to visualize:</h1>")
	for i, header := range headers {
		fmt.Fprintf(w, "<p>%s:</p>", header)
		fmt.Fprintf(w, "<ul>")
		fmt.Fprintf(w, "<li><a href='/pie/%d'>Pie Chart</a></li>", i)
		fmt.Fprintf(w, "<li><a href='/bar/%d'>Bar Chart</a></li>", i)
		fmt.Fprintf(w, "<li><a href='/line/%d'>Line Chart</a></li>", i)
		fmt.Fprintf(w, "</ul>")
	}
}

func renderChart(w http.ResponseWriter, chartType string, data GradeData, column int) {
	counts := make(map[string]int)
	for _, row := range data.Rows {
		if column < len(row) {
			counts[row[column]]++
		}
	}

	title := data.Headers[column] + " Distribution"

	switch chartType {
	case "pie":
		createPieChart(counts, title, w)
	case "bar":
		createBarChart(counts, title, w)
	case "line":
		createLineChart(counts, title, w)
	default:
		http.Error(w, "Invalid chart type", http.StatusBadRequest)
		return
	}

}

func createPieChart(data map[string]int, title string, w http.ResponseWriter) {
	pie := charts.NewPie()
	pie.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))

	items := make([]opts.PieData, 0, len(data))
	for label, value := range data {
		items = append(items, opts.PieData{Name: label, Value: value})
	}
	pie.AddSeries(title, items)

	err := pie.Render(w)
	if err != nil {
		http.Error(w, "Error rendering chart", http.StatusInternalServerError)
	}
}

func createBarChart(data map[string]int, title string, w http.ResponseWriter) {
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))

	var labels []string
	var values []opts.BarData
	for label, value := range data {
		labels = append(labels, label)
		values = append(values, opts.BarData{Value: value})
	}
	bar.SetXAxis(labels).AddSeries(title, values)

	err := bar.Render(w)
	if err != nil {
		http.Error(w, "Error rendering chart", http.StatusInternalServerError)
	}
}

func createLineChart(data map[string]int, title string, w http.ResponseWriter) {
	line := charts.NewLine()
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))

	var labels []string
	var values []opts.LineData
	for label, value := range data {
		labels = append(labels, label)
		values = append(values, opts.LineData{Value: value})
	}
	line.SetXAxis(labels).AddSeries(title, values)

	err := line.Render(w)
	if err != nil {
		http.Error(w, "Error rendering chart", http.StatusInternalServerError)
	}
}
