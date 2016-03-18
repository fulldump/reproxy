package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/fulldump/golax"

	"reproxy/files"
	"reproxy/model"
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

		for _, e := range model.All() {

			if strings.HasPrefix(c.Request.URL.Path, e.Prefix) {

				var t = e.Type

				if "custom" == t {
					type_custom(c, e)
				} else if "statics" == t {
					type_statics(c, e)
				} else if "proxy" == t {
					type_proxy(c, e)
				} else {
					// Missconfiguration
					fmt.Fprint(c.Response, `<h1>This gateway has not been configured</h1>`)
					c.Response.WriteHeader(http.StatusBadGateway)
				}

				break
			}
		}
	})

	return a
}

func type_custom(c *golax.Context, e *model.Entry) {
	custom := e.TypeCustom

	for _, h := range custom.ResponseHeaders {
		c.Response.Header().Set(h.Key, h.Value)
	}

	c.Response.WriteHeader(custom.StatusCode)

	fmt.Fprint(c.Response, custom.Body)
}

func type_statics(c *golax.Context, e *model.Entry) {
	statics := e.TypeStatics

	for _, h := range statics.ResponseHeaders {
		c.Response.Header().Set(h.Key, h.Value)
	}

	c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, e.Prefix)

	d := http.Dir(statics.Directory)
	http.FileServer(d).ServeHTTP(c.Response, c.Request)
}

func type_proxy(c *golax.Context, e *model.Entry) {
	type_proxy := e.TypeProxy

	director := func(r *http.Request) {

		u, _ := url.Parse(type_proxy.Url)

		r.Host = u.Host

		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.URL.Path = u.Path + strings.TrimPrefix(r.URL.Path, e.Prefix)

		for _, h := range type_proxy.ProxyHeaders {
			r.Header.Add(h.Key, h.Value)
		}
	}

	for _, h := range type_proxy.ResponseHeaders {
		c.Response.Header().Set(h.Key, h.Value)
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Response, c.Request)
}

func list_config(c *golax.Context) {

	data := []interface{}{}

	for _, item := range model.All() {
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
	model.Unset(item)
}

func put_config(c *golax.Context) {
	item := &model.Entry{}
	err := json.NewDecoder(c.Request.Body).Decode(item)

	if nil != err {
		c.Error(http.StatusBadRequest, err.Error())
		return
	}

	model.Set(item)
}

func create_config(c *golax.Context) {

	item := &model.Entry{}

	json.NewDecoder(c.Request.Body).Decode(item)

	if "" == item.Prefix {
		c.Error(400, "Attribute 'prefix' is mandatory and must not be empty")
		return
	}

	stored_item := model.GetByPrefix(item.Prefix)
	if nil != stored_item {
		c.Error(http.StatusConflict, "Item already exists")
		return
	}

	model.Set(item)

	json.NewEncoder(c.Response).Encode(item)
}

var interceptor_config_id = &golax.Interceptor{
	Before: func(c *golax.Context) {
		id, noint_err := strconv.Atoi(c.Parameter)
		if nil != noint_err {
			c.Error(400, "Identifier is not valid")
			return
		}

		item := model.GetById(id)
		if nil == item {
			c.Error(404, "Element '"+strconv.Itoa(id)+"' does not exist")
			return
		}

		c.Set("item", item)
	},
}

func get_item(c *golax.Context) *model.Entry {
	if item, exist := c.Get("item"); exist {
		return item.(*model.Entry)
	}

	return nil
}
