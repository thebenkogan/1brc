package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/html"
)

const DATA_URL string = "https://en.wikipedia.org/wiki/List_of_cities_by_average_temperature"

func main() {
	root, _ := os.Getwd()

	var size int
	var path string
	var name string
	flag.IntVar(&size, "size", 100, "number of measurements in the dataset")
	flag.StringVar(&path, "path", filepath.Join(root, "data"), "directory to write the measurements and stats")
	flag.StringVar(&name, "name", "measurements", "name of the data and stats files")
	flag.Parse()

	res, err := http.Get(DATA_URL)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	dom, err := html.Parse(res.Body)
	if err != nil {
		panic(err)
	}

	tableBodies := getElementNodes(dom, "tbody")
	cityAverages := make([]CityAverage, 0)
	for _, tableBody := range tableBodies {
		cityAverages = append(cityAverages, getCityAverages(tableBody)...)
	}

	fmt.Println("Found averages for", len(cityAverages), "cities.")

	dataFilePath := filepath.Join(path, fmt.Sprintf("%s.txt", name))
	file, err := os.Create(dataFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bar := progressbar.Default(int64(size))
	onGenOne := func() {
		_ = bar.Add(1)
	}

	generation, err := GenerateMeasurements(file, cityAverages, size, onGenOne)
	bar.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Wrote %d bytes to %s", generation.TotalBytes, dataFilePath)

	statsPath := filepath.Join(path, fmt.Sprintf("%s.json", name))
	file, err = os.Create(statsPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(generation.CityStats)
	if err != nil {
		panic(err)
	}
}

type Stats struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
	Count int     `json:"count"`
}

type Generation struct {
	CityStats  map[string]*Stats
	TotalBytes int
}

func GenerateMeasurements(output io.Writer, cities []CityAverage, size int, onGenOne func()) (*Generation, error) {
	statsMap := make(map[string]*Stats)
	writer := bufio.NewWriter(output)
	wroteBytes := 0

	for i := range size {
		city := cities[rand.IntN(len(cities))]
		measurement := city.RandomMeasurement()
		measurement = float64(int(measurement*10)) / 10 // single digit precision
		line := fmt.Sprintf("%s;%.1f\n", city.City, measurement)
		if i == size-1 {
			line = strings.TrimSuffix(line, "\n")
		}
		n, err := writer.WriteString(line)
		if err != nil {
			return nil, err
		}
		onGenOne()
		wroteBytes += n

		if cityStats, ok := statsMap[city.City]; ok {
			cityStats.Min = min(cityStats.Min, measurement)
			cityStats.Max = max(cityStats.Max, measurement)
			cityStats.Count++
			N := float64(cityStats.Count)
			cityStats.Mean = cityStats.Mean*(N-1)/N + measurement/N
		} else {
			statsMap[city.City] = &Stats{
				Min:   measurement,
				Max:   measurement,
				Mean:  measurement,
				Count: 1,
			}
		}

	}

	err := writer.Flush()
	if err != nil {
		return nil, err
	}

	return &Generation{
		CityStats:  statsMap,
		TotalBytes: wroteBytes,
	}, nil
}

type CityAverage struct {
	City    string
	Average float64
}

func (c CityAverage) RandomMeasurement() float64 {
	return rand.NormFloat64()*5 + c.Average
}

func getElementNodes(root *html.Node, name string) []*html.Node {
	var elements []*html.Node
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == name {
			elements = append(elements, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)
	return elements
}

func getInnerText(node *html.Node) string {
	var text string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)
	return text
}

func getCityAverages(tableBody *html.Node) []CityAverage {
	cityAverages := make([]CityAverage, 0)

	for tr := tableBody.FirstChild; tr != nil; tr = tr.NextSibling {
		tds := getElementNodes(tr, "td")
		if len(tds) > 0 {
			city := getInnerText(tds[1])
			yearAvgStr := getInnerText(tds[len(tds)-2])
			split := strings.Split(yearAvgStr, "(")
			float, _ := strconv.ParseFloat(split[1][:len(split[1])-2], 64)
			cityAverages = append(cityAverages, CityAverage{
				City:    strings.TrimSuffix(city, "\n"),
				Average: float,
			})
		}
	}

	return cityAverages
}
