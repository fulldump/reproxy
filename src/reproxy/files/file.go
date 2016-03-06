package files

import (
	"net/http"
	"strings"

	"github.com/fulldump/golax"
)

func Build(node *golax.Node) {

	node.Method("GET", func(c *golax.Context) {
		if !strings.HasSuffix(c.Request.URL.Path, "/") {
			http.Redirect(c.Response, c.Request, c.Request.URL.Path+"/", 302)
			return
		}
		readfile(c, "index.html")
	}).
		Node("static").Node("{{*}}").Method("GET", func(c *golax.Context) {
		readfile(c, c.Parameter)
	})

}
