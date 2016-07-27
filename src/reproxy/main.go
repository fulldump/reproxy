package main

import (
	"fmt"
	"net/http"

	"github.com/fulldump/golax"

	"reproxy/api"
	"reproxy/configuration"
	"reproxy/constants"
	"reproxy/model"
)

func main() {
	fmt.Println(constants.BANNER)

	c := configuration.Get()

	model.Load(c.Filename)

	proxy := api.NewReproxy(c.Endpoint)
	Serve(proxy, c.Address)
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
