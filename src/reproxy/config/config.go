package config

import "flag"

var Filename = ""
var Address = ""

// Logo generated with:
// http://patorjk.com/software/taag/#p=display&f=Ivrit&t=ReProxy
var Banner = `
  ____      ____                      
 |  _ \ ___|  _ \ _ __ _____  ___   _ 
 | |_) / _ \ |_) | '__/ _ \ \/ / | | |
 |  _ <  __/  __/| | | (_) >  <| |_| |
 |_| \_\___|_|   |_|  \___/_/\_\\__, |
                                |___/ 
`

func init() {

	flag.Parse()

	flag.StringVar(&Address, "address", "0.0.0.0:8000", "Server address")
	flag.StringVar(&Filename, "filename", "config.json", "Configuration file")

}
