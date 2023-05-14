package homemetrics

import (
	"log"
	"time"

	arg "github.com/alexflint/go-arg"
	toml "github.com/pelletier/go-toml"
)

var config *toml.Tree

func Config(path string) string {
	val, ok := config.Get(path).(string)
	if !ok {
		log.Fatalf("Invalid format for %s: expected string got %v", path, config.Get(path))
	}

	return val
}

func HasConfig(path string) bool {
	return config.Has(path)
}

func Main(collect func(), rate time.Duration) {
	var args struct {
		Config string `default:"/etc/home-metrics.toml"`
	}
	arg.MustParse(&args)

	c, err := toml.LoadFile(args.Config)
	if err != nil {
		log.Fatalf("Can not read config file: %v", err)
	} else {
		config = c
	}

	for {
		start := time.Now()
		collect()

		time.Sleep(rate - time.Since(start))
	}
}
