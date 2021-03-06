package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/fulldump/golax"

	"crypto/tls"
	"reproxy/files"
	"reproxy/model"
)

var LogIncommingTraffic = false

type Priority struct {
	Prefix string
	Entry  *model.Entry
}

type ByPrefix []Priority

func (v ByPrefix) Len() int           { return len(v) }
func (v ByPrefix) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v ByPrefix) Less(i, j int) bool { return 0 < strings.Compare(v[i].Prefix, v[j].Prefix) }

func NewReproxy(endpoint string) *golax.Api {
	a := golax.NewApi()

	reproxy := a.Root.
		Interceptor(golax.InterceptorError).
		Interceptor(golax.InterceptorLog).
		Method("*", all_proxy).
		Node(endpoint)

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
		Method("*", all_proxy)

	return a
}

func all_proxy(c *golax.Context) {

	pritorities := []Priority{}

	for _, e := range model.All() {
		pritorities = append(pritorities, Priority{e.Prefix, e})
	}

	sort.Sort(ByPrefix(pritorities))

	for _, item := range pritorities {
		e := item.Entry

		if strings.HasPrefix(c.Request.URL.Path, e.Prefix) {

			var t = e.Type

			if "custom" == t {
				type_custom(c, e)
			} else if "statics" == t {
				type_statics(c, e)
			} else if "proxy" == t {
				type_proxy(c, e)
			} else if "script" == t {
				type_script(c, e)
			} else {
				// Missconfiguration
				fmt.Fprint(c.Response, `<h1>This gateway has not been configured</h1>`)
				c.Response.WriteHeader(http.StatusBadGateway)
			}

			break
		}
	}
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

	certs := []tls.Certificate{}
	if type_proxy.Key != "" && type_proxy.Cert != "" {
		cert, err := tls.X509KeyPair([]byte(type_proxy.Cert), []byte(type_proxy.Key))
		if nil == err {
			certs = append(certs, cert)
		} else {
			fmt.Println("Create cert error:", err)
		}
	}

	director := func(r *http.Request) {

		if LogIncommingTraffic {
			fmt.Println(r.Method, r.RequestURI, r.Proto)
			fmt.Println("Host:", r.Host)
			for k, values := range r.Header {
				for _, v := range values {
					fmt.Printf("%s: %s\n", k, v)
				}
			}

			fmt.Println("")

			body, _ := ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewReader(body))
			// fmt.Println(body)
			fmt.Printf("%s", body)
			fmt.Println("")
		}

		u, _ := url.Parse(type_proxy.Url)

		r.Host = u.Host

		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.URL.Path = u.Path + strings.TrimPrefix(r.URL.Path, e.Prefix)

		for _, h := range type_proxy.ProxyHeaders {
			r.Header.Add(h.Key, h.Value)
			if "Host" == h.Key {
				r.Host = h.Value
			}
		}
	}

	for _, h := range type_proxy.ResponseHeaders {
		c.Response.Header().Set(h.Key, h.Value)
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				Certificates:       certs,
			},
		},
	}
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
