package main

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	root, _ := os.Getwd()
	dataPath := filepath.Join(root, "data", "measurements.txt")
	file, err := os.Open(dataPath)
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
		go worker(dataPath, offset, amount, storesChan)
	}

	stores := make([]*StatsStore, numWorkers)
	for i := range numWorkers {
		stores[i] = <-storesChan
	}

	merged := MergeStores(stores)

	resultsFile, err := os.Create(filepath.Join(root, "data", "results.json"))
	if err != nil {
		panic(err)
	}
	defer resultsFile.Close()

	err = json.NewEncoder(resultsFile).Encode(merged)
	if err != nil {
		panic(err)
	}
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
