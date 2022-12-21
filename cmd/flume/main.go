package main

import (
	"log"
	"os"
	"time"

	homemetrics "github.com/natemueller/home-metrics-collectors"

	flume "github.com/russorat/flume-water-go-client"
)

const FLUMETIME = "2006-01-02 15:04:05"

func collect() {
	client := flume.NewClient(
		homemetrics.Config("flume.clientId"),
		homemetrics.Config("flume.clientSecret"),
		homemetrics.Config("flume.username"),
		homemetrics.Config("flume.password"),
	)
	devices, err := client.FetchUserDevices(flume.FlumeWaterFetchDeviceRequest{})
	if err != nil {
		log.Printf("Failed to fetch devices: %v", err)
		return
	}

	tzdata, err := os.ReadFile("/etc/localtime")
	if err != nil {
		log.Printf("Failed to read tzdata: %v", err)
		return
	}
	location, err := time.LoadLocationFromTZData(homemetrics.Config("time_zone"), tzdata)
	if err != nil {
		log.Printf("Failed to read tzdata: %v", err)
		return
	}

	startTime := (time.Now().Add(-10 * time.Minute)).Format(FLUMETIME)

	log.Printf("Query start time: %v", startTime)

	query := flume.FlumeWaterQuery{
		Bucket:        flume.FlumeWaterBucketMinute,
		SinceDatetime: startTime,
		RequestID:     "query",
	}
	results, err := client.QueryUserDevice(devices[0].ID, flume.FlumeWaterQueryRequest{
		Queries: []flume.FlumeWaterQuery{query},
	})
	if err != nil {
		log.Printf("Failed to fetch historical data: %v", err)
		return
	}

	for _, point := range results[0]["query"] {
		parsed, err := time.ParseInLocation(FLUMETIME, point.Datetime, location)
		if err != nil {
			log.Printf("Failed to parse date \"%v\": %v", point.Datetime, err)
			return
		}

		homemetrics.SendCarbon([]string{"flume", "rate", homemetrics.CleanLabel(devices[0].ID)}, point.Value, parsed)
	}
}

func main() {
	homemetrics.Main(collect, 5*time.Minute)
}
