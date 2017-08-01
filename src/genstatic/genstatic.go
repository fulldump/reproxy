package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	dir := ""
	flag.StringVar(&dir, "dir", "", "MANDATORY")

	pkg := ""
	flag.StringVar(&pkg, "package", "", "MANDATORY")

	flag.Parse()

	if "" == dir {
		fmt.Println("Argument `dir` is mandatory")
		return
	}

	if "" == pkg {
		fmt.Println("Argument `package` is mandatory")
		return
	}

	dir = strings.TrimRight(dir, "/") + "/"

	generate(dir, pkg)
}

func generate(dir, pkg string) {
	fmt.Println("package " + pkg)
	fmt.Println("")
	fmt.Println("var Files = map[string]string{")

	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {

		if !f.IsDir() {
			key := strings.TrimPrefix(path, dir)
			value := read_base64(path)
			fmt.Println("\t\"" + key + "\": `" + value + "`,")
		}

		return nil
	})

	fmt.Println("}")

	if nil != err {
		panic(err)
	}
}

func read_bytes(filename string) []byte {
	bytes, err := ioutil.ReadFile(filename)
	if nil != err {
		panic(err)
	}

	return bytes
}

func read_string(filename string) string {
	bytes := read_bytes(filename)
	return string(bytes)
}

func read_base64(filename string) string {
	bytes := read_bytes(filename)

	return base64.StdEncoding.EncodeToString(bytes)
}
