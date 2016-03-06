package api

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/fulldump/golax"

	"reproxy/config"
	"reproxy/files"
)

func NewReproxy() *golax.Api {
	a := golax.NewApi()

	reproxy := a.Root.
		Interceptor(golax.InterceptorError).
		Interceptor(golax.InterceptorLog).
		Node("reproxy")

	reproxy.
		Node("config").
		Method("GET", list_config).
		Method("POST", create_config).
		Node("{config_id}").
		Interceptor(interceptor_config_id).
		Method("DELETE", delete_config).
		Method("PUT", put_config).
		Method("GET", get_config)

	files.Build(reproxy) // Add static files

	a.Root.
		Node("{{*}}").
		Method("*", func(c *golax.Context) {

		for _, e := range config.All() {
			if strings.HasPrefix(c.Request.URL.Path, e.Prefix) {
				director := func(r *http.Request) {
					r = c.Request
					r.URL.Scheme = "http"
					r.URL.Host = e.Url
					for _, header := range e.Headers {
						c.Request.Header.Add(header.Key, header.Value)
					}
				}
				proxy := &httputil.ReverseProxy{Director: director}
				proxy.ServeHTTP(c.Response, c.Request)
				break
			}
		}
	})

	return a
}

func list_config(c *golax.Context) {

	data := []interface{}{}

	for _, item := range config.All() {
		data = append(data, item)
	}

	json.NewEncoder(c.Response).Encode(data)
}

func get_config(c *golax.Context) {
	item := get_item(c)
	json.NewEncoder(c.Response).Encode(item)
}

func delete_config(c *golax.Context) {
	item := get_item(c)
	config.Unset(item)
}

func put_config(c *golax.Context) {
	item := &config.Entry{}
	err := json.NewDecoder(c.Request.Body).Decode(item)

	if nil != err {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}

	config.Set(item)
}

func create_config(c *golax.Context) {

	item := &config.Entry{}

	json.NewDecoder(c.Request.Body).Decode(item)

	if "" == item.Prefix {
		c.Error(400, "Attribute 'prefix' is mandatory and must not be empty")
		return
	}

	stored_item := config.GetByPrefix(item.Prefix)
	if nil != stored_item {
		c.Error(http.StatusConflict, "Item already exists")
		return
	}

	config.Set(item)

	json.NewEncoder(c.Response).Encode(item)
}

var interceptor_config_id = &golax.Interceptor{
	Before: func(c *golax.Context) {
		id, noint_err := strconv.Atoi(c.Parameter)
		if nil != noint_err {
			c.Error(400, "Identifier is not valid")
			return
		}

		item := config.GetById(id)
		if nil == item {
			c.Error(404, "Element '"+strconv.Itoa(id)+"' does not exist")
			return
		}

		c.Set("item", item)
	},
}

func get_item(c *golax.Context) *config.Entry {
	if item, exist := c.Get("item"); exist {
		return item.(*config.Entry)
	}

	return nil
}
