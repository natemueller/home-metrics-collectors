package main

import (
	"log"
	"time"

	homemetrics "github.com/natemueller/home-metrics-collectors"

	"github.com/lrosenman/ambient"
)

func collect() {
	key := ambient.NewKey(homemetrics.Config("ambient_weather.applicationKey"), homemetrics.Config("ambient_weather.apiKey"))
	dr, err := ambient.Device(key)
	if err != nil {
		log.Printf("Failed to fetch device: %v", err)
		return
	}

	switch dr.HTTPResponseCode {
	case 200:
	default:
		{
			log.Printf("HTTPResponseCode=%d fetching device list", dr.HTTPResponseCode)
			return
		}
	}

	for z := range dr.DeviceRecord {
		time.Sleep(1 * time.Second) // API RateLimit

		ar, err := ambient.DeviceMac(key, dr.DeviceRecord[z].Macaddress, time.Now(), 1)
		if err != nil {
			log.Printf("Failed to fetch latest measurements: %v", err)
			return
		}

		switch ar.HTTPResponseCode {
		case 200:
		default:
			{
				log.Printf("HTTPResponseCode=%d fetching latest measurements", dr.HTTPResponseCode)
				return
			}
		}

		for label, value := range ar.RecordFields[0] {
			floatValue, ok := value.(float64)
			if !ok {
				continue
			}

			homemetrics.SendCarbon([]string{"ambient", label, homemetrics.CleanLabel(dr.DeviceRecord[z].Info.Location)}, floatValue, time.Now())
		}
	}
}

func main() {
	homemetrics.Main(collect, time.Minute)
}
