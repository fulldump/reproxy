package main

import (
	"fmt"
	"net/http"

	"github.com/fulldump/golax"

	"reproxy/api"
	"reproxy/config"
	"reproxy/model"
)

func main() {

	fmt.Println(config.Banner)

	model.Load(config.Filename)

	proxy := api.NewReproxy()
	Serve(proxy, config.Address)

}

func Serve(a *golax.Api, address string) {
	s := &http.Server{
		Addr:    address,
		Handler: a,
	}

	fmt.Println("Listening at " + address)

	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
	}
}
