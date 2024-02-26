package main

import (
	"bufio"
	"flag"
	"fmt"
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
	var size int
	flag.IntVar(&size, "size", 100, "number of measurements in the dataset")
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

	root, _ := os.Getwd()
	outDir := filepath.Join(root, "data", "measurements.txt")
	file, err := os.Create(outDir)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bar := progressbar.Default(int64(size))

	writer := bufio.NewWriter(file)
	wroteBytes := 0
	for range size {
		city := cityAverages[rand.IntN(len(cityAverages))]
		measurement := city.RandomMeasurement()
		n, err := writer.WriteString(fmt.Sprintf("%s;%.1f\n", city.City, measurement))
		if err != nil {
			panic(err)
		}
		wroteBytes += n
		_ = bar.Add(1)
	}
	bar.Close()

	err = writer.Flush()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Wrote %d bytes to %s", wroteBytes, outDir)
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
