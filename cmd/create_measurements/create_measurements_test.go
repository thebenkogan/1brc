package main

import (
	"bytes"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	size := 500

	sampleData := []CityAverage{
		{"New York", 12.3},
		{"Los Angeles", 18.2},
		{"Chicago", 9.8},
	}

	buf := new(bytes.Buffer)
	generation, err := GenerateMeasurements(buf, sampleData, size, func() {})
	if err != nil {
		t.Fatal(err)
	}

	data := strings.Split(buf.String(), "\n")

	if len(data) != size {
		t.Errorf("Expected %d lines, got %d", size, len(data))
	}

	samples := make(map[string][]float64)
	for _, city := range sampleData {
		samples[city.City] = make([]float64, 0)
	}

	t.Run("Measurement format", func(t *testing.T) {
		for _, line := range data {
			parts := strings.Split(line, ";")
			if len(parts) != 2 {
				t.Errorf("Expected 2 parts, got %d", len(parts))
			}

			city := parts[0]
			measurement, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				t.Errorf("Expected a float measurement, got %s", parts[1])
			}

			if _, ok := samples[city]; !ok {
				t.Errorf("Unexpected city %s", city)
			}

			samples[city] = append(samples[city], measurement)
		}
	})

	cityStats := generation.CityStats
	for _, city := range sampleData {
		t.Run(city.City, func(t *testing.T) {
			citySamples := samples[city.City]
			stats := cityStats[city.City]

			if len(citySamples) != stats.Count {
				t.Errorf("Expected %d samples for %s, got %d", stats.Count, city.City, len(citySamples))
			}

			assertClose(t, "min", slices.Min(citySamples), stats.Min)
			assertClose(t, "max", slices.Max(citySamples), stats.Max)
			assertClose(t, "mean", getAverage(t, citySamples), stats.Mean)

			cityIdx := slices.IndexFunc(sampleData, func(c CityAverage) bool {
				return c.City == city.City
			})
			targetMean := sampleData[cityIdx].Average

			if math.Abs(targetMean-stats.Mean) > 1 {
				t.Errorf("Expected mean to be close to %f, got %f", targetMean, stats.Mean)
			}
		})
	}
}

func assertClose(t *testing.T, name string, expected, got float64) {
	t.Helper()
	tolerance := 1e-9
	if math.Abs(expected-got) > tolerance {
		t.Errorf("Expected %f %s, got %f", expected, name, got)
	}
}

func getAverage(t *testing.T, samples []float64) float64 {
	t.Helper()
	var sum float64
	for _, s := range samples {
		sum += s
	}
	return sum / float64(len(samples))
}
