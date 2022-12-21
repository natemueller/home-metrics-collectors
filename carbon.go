package homemetrics

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

func CleanLabel(raw string) string {
	app := regexp.MustCompile(`['’]`)
	invalid := regexp.MustCompile(`[^a-z0-9:]`)

	return invalid.ReplaceAllString(app.ReplaceAllString(strings.ToLower(raw), ""), "-")
}

var carbonConn net.Conn = nil

func SendCarbon(metricName []string, metricValue float64, metricTime time.Time) {
	formattedName := "home.sixth." + strings.Join(metricName, ".")
	packet := fmt.Sprintf("%s %f %d\n", formattedName, metricValue, metricTime.Unix())

	if carbonConn == nil {
		log.Printf("No active carbon connection, initiating new one")

		tcpAddr, err := net.ResolveTCPAddr("tcp", Config("carbon_endpoint"))
		if err != nil {
			log.Printf("Carbon DNS failed: %v", err)
			return
		}

		carbonConn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			log.Printf("Carbon connect failed: %v", err)
			carbonConn = nil
			return
		}
	}

	_, err := carbonConn.Write([]byte(packet))
	if err != nil {
		log.Printf("Carbon send failed: %v", err)
		carbonConn = nil
		return
	}
}
