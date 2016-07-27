package configuration

import (
	"fmt"
	"os"
	"reproxy/constants"

	"github.com/fulldump/goconfig"
)

type Config struct {
	Address  string `Default address to listen`
	Filename string `Default configuration filename to save configuration`
	Endpoint string `Configuration endpoint for reproxy`
	Version  bool   `Show version`
}

var c *Config = nil

func Get() *Config {
	if nil == c {
		c = &Config{
			Address:  "0.0.0.0:8000",
			Filename: "reproxy.json",
			Endpoint: "reproxy",
		}
		goconfig.Read(c)

		if c.Version {
			fmt.Println("Version: " + constants.VERSION)
			fmt.Println("Build date: " + constants.BUILD_DATE)
			os.Exit(0)
		}

	}

	return c
}
