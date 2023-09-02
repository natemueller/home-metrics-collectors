package main

import (
	"log"
	"time"

	homemetrics "github.com/natemueller/home-metrics-collectors"

	purpleair "github.com/poynting/purpleair-api-go/purpleair"
)

func collect() {
	client, err := purpleair.NewClient(
		homemetrics.Config("purple_air.apiKey"),
		"",
	)
	if err != nil {
		log.Printf("Failed to create Purple Air client: %v", err)
		return
	}

	params := map[string]string{
		"fields":        "pm2.5_10minute",
		"location_type": "0",
		"show_only":     homemetrics.Config("purple_air.sensorId"),
	}
	s, err := client.GetSensors(params)
	if err != nil {
		log.Printf("Failed to fetch sensor data: %v", err)
		return
	}

	samples := client.SensorsToSamples(s.DataTimeStamp, s.Fields, s.Data)
	if len(samples) > 0 {
		homemetrics.SendCarbon([]string{"purpleair", "pm2_5"}, toAQI(float64(samples[0].Sampledata["pm2.5_10minute"])), time.Now())
	}
}

func toAQI(pm float64) float64 {
	var pmRange = []float64{0, 12.1, 35.5, 55.5, 150.5, 250.5, 350.5, 500.4}
	var aqiRange = []float64{0, 51, 101, 151, 201, 301, 401, 500}

	for i, _ := range pmRange {
		if pm > pmRange[i+1] {
			continue
		}

		return ((pm-pmRange[i])/(pmRange[i+1]-pmRange[i]))*(aqiRange[i+1]-aqiRange[i]) + aqiRange[i]
	}

	return aqiRange[len(aqiRange)-1]
}

func main() {
	homemetrics.Main(collect, 10*time.Minute)
}
