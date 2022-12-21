package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	homemetrics "github.com/natemueller/home-metrics-collectors"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var configHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	topic := regexp.MustCompile(`/`).ReplaceAllString(msg.Topic()[21:len(msg.Topic())-7], "/sensor/") + "/state"
	client.Subscribe(topic, 1, stateHandler)
}

var stateHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	parts := strings.Split(msg.Topic(), "/")
	panel := parts[0]
	sensor := parts[2]

	value, err := strconv.ParseFloat(string(msg.Payload()), 64)
	if err != nil {
		log.Printf("Invalid value from sensor %s: %s", msg.Topic(), msg.Payload())
		return
	}

	if regexp.MustCompile(`phase_._voltage`).MatchString(sensor) {
		sensor = sensor[0 : strings.Index(sensor, "voltage")-1]
		go func() {
			homemetrics.SendCarbon([]string{"vue", "vue_phase_voltage", homemetrics.CleanLabel(panel), homemetrics.CleanLabel(sensor)}, value, time.Now())
		}()
	} else if regexp.MustCompile(`phase_._power`).MatchString(sensor) {
		sensor = sensor[0 : strings.Index(sensor, "power")-1]
		go func() {
			homemetrics.SendCarbon([]string{"vue", "vue_phase_power", homemetrics.CleanLabel(panel), homemetrics.CleanLabel(sensor)}, value, time.Now())
		}()
	} else {
		go func() {
			homemetrics.SendCarbon([]string{"vue", "vue_circuit_power", homemetrics.CleanLabel(panel), homemetrics.CleanLabel(sensor)}, value, time.Now())
		}()
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	panic(err)
}

func collect() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(homemetrics.Config("vue.mqtt_endpoint"))
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("unable to connect to mqtt broker: %v", token.Error())
		return
	}

	client.Subscribe("homeassistant/#", 1, configHandler).Wait()

	for {
		time.Sleep(time.Minute)
	}
}

func main() {
	homemetrics.Main(collect, time.Minute)
}
