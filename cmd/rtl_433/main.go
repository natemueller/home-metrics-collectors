package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	homemetrics "github.com/natemueller/home-metrics-collectors"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var stateHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	parts := strings.Split(msg.Topic(), "/")
	device_id := parts[2]
	metric := parts[3]
	value, err := strconv.ParseFloat(string(msg.Payload()), 64)
	if err != nil {
		return
	}

	if homemetrics.HasConfig("rtl_433.id_map." + device_id) {
		go func() {
			homemetrics.SendCarbon([]string{
				"rtl_433",
				homemetrics.CleanLabel(homemetrics.Config("rtl_433.id_map." + device_id)),
				homemetrics.CleanLabel(metric),
			}, value, time.Now())
		}()
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	panic(err)
}

func collect() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(homemetrics.Config("rtl_433.mqtt_endpoint"))
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("unable to connect to mqtt broker: %v", token.Error())
		return
	}

	client.Subscribe("rtl_433/devices/#", 1, stateHandler).Wait()

	for {
		time.Sleep(time.Minute)
	}
}

func main() {
	homemetrics.Main(collect, time.Minute)
}
