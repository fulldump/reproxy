package main

import (
	"flag"
	"net/http"

	"github.com/fulldump/golax"

	"reproxy/api"
	"reproxy/config"
)

func main() {

	flag.Parse()
	filename := flag.String("config", "config.json", "Configuration file")
	config.Load(*filename)

	proxy := api.NewReproxy()
	Serve(proxy)

}

func Serve(a *golax.Api) {
	s := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: a,
	}

	s.ListenAndServe()
}
