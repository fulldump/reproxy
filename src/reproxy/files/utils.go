package files

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fulldump/golax"
)

var contenttypes = map[string]string{

	// Web
	".css":  "text/css; charset=utf-8",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".js":   "text/javascript; charset=UTF-8",
	".json": "application/json; charset=UTF-8",

	// Images
	".ico":  "image/x-icon",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".svg":  "image/svg+xml",

	// Fonts:
	".eot":   "application/vnd.ms-fontobject",
	".otf":   "application/x-font-opentype",
	".sfnt":  "application/font-sfnt",
	".ttf":   "application/x-font-ttf",
	".woff":  "application/font-woff",
	".woff2": "application/font-woff2",
}

func addcontenttype(c *golax.Context, filename string) {
	ext := strings.ToLower(filepath.Ext(filename))

	value, exists := contenttypes[ext]
	if exists {
		c.Response.Header().Add("Content-Type", value)
	}
}

func readfile(c *golax.Context, filename string) {
	encoded, exists := Files[filename]
	if !exists {
		c.Error(404, "File '"+filename+"' not found")
		return
	}

	content, err := base64.StdEncoding.DecodeString(encoded)
	if nil != err {
		c.Error(500, "Unexpected error: "+err.Error())
	}

	addcontenttype(c, filename)

	fmt.Fprint(c.Response, string(content))
}

/**
 * Only for develop purposes
 */
func readfile_dev(c *golax.Context, filename string) {
	PREFIX := "static/"

	bytes, err := ioutil.ReadFile(PREFIX + filename)
	if nil != err {
		c.Error(404, "File '"+filename+"' not found")
		return
	}

	addcontenttype(c, filename)

	fmt.Fprint(c.Response, string(bytes))
}
