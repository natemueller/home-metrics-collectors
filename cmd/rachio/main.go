package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	homemetrics "github.com/natemueller/home-metrics-collectors"
)

func apiGet(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+homemetrics.Config("rachio.apiKey"))

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func collect() {
	resp, err := apiGet("https://api.rach.io/1/public/person/" + homemetrics.Config("rachio.personId"))
	if err != nil {
		log.Printf("Person info fetch failed: %v", err)
		return
	}

	var personInfo map[string]interface{}
	json.Unmarshal([]byte(resp), &personInfo)

	devices, ok := personInfo["devices"].([]interface{})
	if !ok {
		log.Printf("Invalid format for device list: %v", personInfo["devices"])
		return
	}

	for deviceIdx := range devices {
		device, ok := devices[deviceIdx].(map[string]interface{})
		if !ok {
			log.Printf("Invalid format for device entry: %v", devices[deviceIdx])
			return
		}

		deviceName, ok := device["name"].(string)
		if !ok {
			log.Printf("Invalid format for device name: %v", device["name"])
			return
		}

		zones, ok := device["zones"].([]interface{})
		if !ok {
			log.Printf("Invalid format for zone list: %v", device["zones"])
			return
		}

		for zoneIdx := range zones {
			zone, ok := zones[zoneIdx].(map[string]interface{})
			if !ok {
				log.Printf("Invalid format for zone entry: %v", zones[zoneIdx])
				return
			}

			zoneName, ok := zone["name"].(string)
			if !ok {
				log.Printf("Invalid format for zone name: %v", zone["name"])
				return
			}

			enabled, ok := zone["enabled"].(bool)
			if !ok {
				log.Printf("Invalid format for enabled: %v", zone["enabled"])
				return
			}

			if !enabled {
				continue
			}

			var isRunning float64
			if _, ok := zone["lastWateredDuration"]; ok {
				isRunning = 0
			} else {
				isRunning = 1
			}

			homemetrics.SendCarbon([]string{"rachio", "is-running", homemetrics.CleanLabel(deviceName), homemetrics.CleanLabel(zoneName)}, isRunning, time.Now())
		}
	}
}

func main() {
	homemetrics.Main(collect, time.Minute)
}
