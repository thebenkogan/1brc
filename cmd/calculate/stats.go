package main

import "encoding/json"

type Stats struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Sum   float64 `json:"sum"`
	Count int     `json:"count"`
}

// map of city name to stats, not thread safe
type StatsStore struct {
	stats map[string]*Stats
}

func NewStatsStore() *StatsStore {
	return &StatsStore{
		stats: make(map[string]*Stats),
	}
}

func MergeStores(stores []*StatsStore) *StatsStore {
	merged := NewStatsStore()
	for _, store := range stores {
		for city, stats := range store.stats {
			if mergedCityStats, ok := merged.stats[city]; ok {
				mergedCityStats.Min = min(mergedCityStats.Min, stats.Min)
				mergedCityStats.Max = max(mergedCityStats.Max, stats.Max)
				mergedCityStats.Sum += stats.Sum
				mergedCityStats.Count += stats.Count
			} else {
				merged.stats[city] = stats
			}
		}
	}
	return merged
}

func (s *StatsStore) recordMeasurement(city string, measurement float64) {
	if cityStats, ok := s.stats[city]; ok {
		cityStats.Min = min(cityStats.Min, measurement)
		cityStats.Max = max(cityStats.Max, measurement)
		cityStats.Sum += measurement
		cityStats.Count++
	} else {
		s.stats[city] = &Stats{
			Min:   measurement,
			Max:   measurement,
			Sum:   measurement,
			Count: 1,
		}
	}
}

type StatsJson struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
	Count int     `json:"count"`
}

func (s *StatsStore) MarshalJSON() ([]byte, error) {
	jsonStore := make(map[string]StatsJson)
	for city, stats := range s.stats {
		jsonStore[city] = StatsJson{
			Min:   stats.Min,
			Max:   stats.Max,
			Mean:  stats.Sum / float64(stats.Count),
			Count: stats.Count,
		}
	}
	return json.Marshal(jsonStore)
}
