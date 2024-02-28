package main

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCalculate(t *testing.T) {
	testDir := t.TempDir()
	name := "measurements"
	root, _ := os.Getwd()
	createPath := filepath.Join(root, "../../", "cmd/create_measurements/create_measurements.go")

	if err := exec.Command("go", "run", createPath, "-path", testDir, "-name", name, "-size", "100000").Run(); err != nil {
		t.Fatal(err)
	}

	dataPath := filepath.Join(testDir, name+".txt")
	expectedPath := filepath.Join(testDir, name+".json")

	stats := calculate(dataPath)
	j, _ := json.Marshal(stats)
	resultsStats := make(map[string]StatsJson)
	if err := json.Unmarshal(j, &resultsStats); err != nil {
		t.Fatal(err)
	}

	expected, err := os.Open(expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedStats := make(map[string]StatsJson)
	if err := json.NewDecoder(expected).Decode(&expectedStats); err != nil {
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
