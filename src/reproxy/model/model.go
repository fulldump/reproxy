package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var filename = ""

var data_by_id = map[int]*Entry{}
var data_by_prefix = map[string]*Entry{}
var next_id = 0

type Entry struct {
	Id          int         `json:"id"`
	Prefix      string      `json:"prefix"`
	Type        string      `json:"type"`
	TypeCustom  TypeCustom  `json:"type_custom"`
	TypeStatics TypeStatics `json:"type_statics"`
	TypeProxy   TypeProxy   `json:"type_proxy"`
}

type TypeCustom struct {
	StatusCode      int     `json:"status_code"`
	ResponseHeaders Headers `json:"response_headers"`
	Body            string  `json:"body"`
}

type TypeStatics struct {
	ResponseHeaders Headers `json:"response_headers"`
	Directory       string  `json:"directory"`
}

type TypeProxy struct {
	Url             string  `json:"url"`
	ResponseHeaders Headers `json:"response_headers"`
	ProxyHeaders    Headers `json:"proxy_headers"`
}

type Headers []Header

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func All() map[int]*Entry {

	return data_by_id
}

func GetById(id int) *Entry {
	if item, exist := data_by_id[id]; exist {
		return item
	}

	return nil
}

func GetByPrefix(prefix string) *Entry {
	if item, exist := data_by_prefix[prefix]; exist {
		return item
	}

	return nil
}

func Set(item *Entry) {
	// Check if exists
	if old_item, exist := data_by_prefix[item.Prefix]; exist {
		delete(data_by_prefix, old_item.Prefix)
		item.Id = old_item.Id
	} else {
		item.Id = next_id
		next_id++
	}

	data_by_id[item.Id] = item
	data_by_prefix[item.Prefix] = item

	if nil == item.TypeCustom.ResponseHeaders {
		item.TypeCustom.ResponseHeaders = Headers{}
	}

	if nil == item.TypeProxy.ResponseHeaders {
		item.TypeProxy.ResponseHeaders = Headers{}
	}

	if nil == item.TypeProxy.ProxyHeaders {
		item.TypeProxy.ProxyHeaders = Headers{}
	}

	if nil == item.TypeStatics.ResponseHeaders {
		item.TypeStatics.ResponseHeaders = Headers{}
	}

	save()
}

func Unset(item *Entry) {
	delete(data_by_prefix, item.Prefix)
	delete(data_by_id, item.Id)

	save()
}

func Load(f string) {
	filename = f

	d, err := ioutil.ReadFile(f)
	if nil != err {
		fmt.Println("Unable to read config file, don't worry, ReProxy is running, just configure it at /reproxy/")
		return
	}

	items := []*Entry{}
	err = json.Unmarshal(d, &items)
	if nil != err {
		fmt.Println("Config file is supposed to be a JSON")
		return
	}

	for _, item := range items {
		Set(item)
	}
}

func save() {

	fp, err := os.Create(filename)
	if err != nil {
		fmt.Println("Unable to create %v. Err: %v.", filename, err)
		return
	}
	defer fp.Close()

	data := []interface{}{}
	for _, item := range All() {
		data = append(data, item)
	}

	encoder := json.NewEncoder(fp)
	if err = encoder.Encode(data); err != nil {
		fmt.Println("Unable to encode Json file. Err: %v.", err)
		return
	}

}
