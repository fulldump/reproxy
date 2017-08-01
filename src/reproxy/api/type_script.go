package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reproxy/vm"

	"reproxy/model"

	"github.com/fulldump/golax"
)

func type_script(c *golax.Context, e *model.Entry) {
	defer func() {
		if r := recover(); r != nil {
			c.Error(http.StatusInternalServerError, "JSVM: "+fmt.Sprint(r))
		}
	}()

	// Get query params:
	query_params := map[string]interface{}{}
	for a, b := range c.Request.URL.Query() {
		if 1 == len(b) {
			query_params[a] = b[0]
		} else {
			query_params[a] = b
		}
	}

	v := vm.New()

	v.Set("response", map[string]interface{}{
		"error":     c.Error,
		"addHeader": c.Response.Header().Add,
		"setHeader": c.Response.Header().Set,
		"setStatus": c.Response.WriteHeader,
		"write": func(s interface{}) {
			// TODO: ensure s is type string
			fmt.Fprint(c.Response, s)
		},
		"writeJson": func(s interface{}) {
			json.NewEncoder(c.Response).Encode(s)
		},
	})
	v.Set("request", map[string]interface{}{
		"getHeaders": func() map[string]string {
			h := map[string]string{}

			for k, v := range c.Request.Header {
				h[k] = v[0]
			}

			return h
		},
		"method": c.Request.Method,
		"getQuery": func() interface{} {
			return query_params
		},
		"getBody": func() interface{} {
			bytes, _ := ioutil.ReadAll(c.Request.Body)
			return bytes
		},
		"getBodyJson": func() interface{} {
			var v interface{}
			json.NewDecoder(c.Request.Body).Decode(&v)
			return v
		},
	})

	_, v_err := v.Run(e.TypeScript.Code)

	if nil != v_err {
		c.Error(500, "JSVM: "+v_err.Error())
	}

}
