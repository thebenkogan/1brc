package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestCalculate(t *testing.T) {
	resultsPath := filepath.Join("/home/benkogan/code/1brc", "data", "results.json")
	results, err := os.Open(resultsPath)
	if err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join("/home/benkogan/code/1brc", "data", "measurements.json")
	expected, err := os.Open(expectedPath)
	if err != nil {
		t.Fatal(err)
	}

	expectedStats := make(map[string]StatsJson)
	if err := json.NewDecoder(expected).Decode(&expectedStats); err != nil {
		t.Fatal(err)
	}

	resultsStats := make(map[string]StatsJson)
	if err := json.NewDecoder(results).Decode(&resultsStats); err != nil {
		t.Fatal(err)
	}

	if len(expectedStats) != len(resultsStats) {
		t.Errorf("Expected %d stats, got %d", len(expectedStats), len(resultsStats))
	}

	for city, expected := range expectedStats {
		t.Run(city, func(t *testing.T) {
			resultsStats, ok := resultsStats[city]
			if !ok {
				t.Errorf("Missing stats for %s", city)
			}
			assertClose(t, "min", expected.Min, resultsStats.Min)
			assertClose(t, "max", expected.Max, resultsStats.Max)
			assertClose(t, "mean", expected.Mean, resultsStats.Mean)
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
