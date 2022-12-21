package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	homemetrics "github.com/natemueller/home-metrics-collectors"

	"golang.org/x/oauth2/clientcredentials"
)

type Segment struct {
	Id      string
	Name    string
	Started string
	Active  bool
}

type Location struct {
	Id   string
	Name string
}

type Device struct {
	Id         string
	DeviceType string
	Sensors    []string
	Segment    Segment
	Location   Location
}

type DeviceList struct {
	Devices []Device
	Error   string
}

func apiGet(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
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
	config := &clientcredentials.Config{
		ClientID:     homemetrics.Config("airthings.clientId"),
		ClientSecret: homemetrics.Config("airthings.clientSecret"),
		Scopes:       []string{"read:device:current_values"},
		TokenURL:     "https://accounts-api.airthings.com/v1/token",
	}
	client := config.Client(context.Background())

	var parsedDeviceList DeviceList
	resp, err := apiGet(client, "https://ext-api.airthings.com/v1/devices")
	if err != nil {
		log.Printf("Device list fetch failed: %v", err)
		return
	}
	json.Unmarshal([]byte(resp), &parsedDeviceList)
	if parsedDeviceList.Error != "" {
		log.Printf("Device list API error: %v", parsedDeviceList.Error)
		return
	}

	for i := range parsedDeviceList.Devices {
		d := parsedDeviceList.Devices[i]
		if d.DeviceType == "HUB" {
			continue
		}

		var latestSamples map[string]map[string]interface{}
		resp, err = apiGet(client, "https://ext-api.airthings.com/v1/devices/"+d.Id+"/latest-samples")
		if err != nil {
			log.Printf("Latest sample fetch failed: %v", err)
			return
		}
		json.Unmarshal([]byte(resp), &latestSamples)

		for label, value := range latestSamples["data"] {
			floatValue, ok := value.(float64)
			if !ok {
				continue
			}

			homemetrics.SendCarbon([]string{"airthings", label, homemetrics.CleanLabel(d.Segment.Name)}, floatValue, time.Now())
		}
	}
}

func main() {
	homemetrics.Main(collect, 4*time.Minute)
}
