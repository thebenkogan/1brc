package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	root, _ := os.Getwd()
	var dataPath string
	var resultsPath string
	flag.StringVar(&dataPath, "source", filepath.Join(root, "data", "measurements.txt"), "path to the data file")
	flag.StringVar(&resultsPath, "out", filepath.Join(root, "data", "results.json"), "path to store results")

	stats := calculate(dataPath)

	resultsFile, err := os.Create(resultsPath)
	if err != nil {
		panic(err)
	}
	defer resultsFile.Close()

	err = json.NewEncoder(resultsFile).Encode(stats)
	if err != nil {
		panic(err)
	}
}

func calculate(path string) *StatsStore {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}
	totalBytes := stat.Size()
	file.Close()

	numWorkers := runtime.NumCPU()
	amountPerWorker := totalBytes / int64(numWorkers)
	storesChan := make(chan *StatsStore)
	for i := range numWorkers {
		offset := totalBytes / int64(numWorkers) * int64(i)
		amount := amountPerWorker
		if i == numWorkers-1 {
			// last worker should read up to end of file, this fixes any division rounding errors in amountPerWorker
			amount = math.MaxInt64
		}
		go worker(path, offset, amount, storesChan)
	}

	stores := make([]*StatsStore, numWorkers)
	for i := range numWorkers {
		stores[i] = <-storesChan
	}

	return MergeStores(stores)
}

func worker(path string, offset int64, toRead int64, storesChan chan<- *StatsStore) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if _, err := file.Seek(offset, 0); err != nil {
		panic(err)
	}

	var numBytes int64

	reader := bufio.NewReader(file)
	// offset might be in the middle of a line, so move to the next line
	if offset > 0 {
		b, err := reader.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		numBytes += int64(len(b))
	}

	scanner := bufio.NewScanner(reader)
	store := NewStatsStore()
	for scanner.Scan() {
		b := scanner.Bytes()
		line := string(b)
		split := strings.Split(line, ";")
		city := split[0]
		measurement, err := strconv.ParseFloat(split[1], 64)
		if err != nil {
			panic(err)
		}
		store.recordMeasurement(city, measurement)
		numBytes += int64(len(b)) + 1 // +1 for the discarded newline
		if numBytes > toRead {
			break
		}
	}

	storesChan <- store
}
