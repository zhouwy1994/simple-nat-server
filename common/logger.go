package common

import (
	"io"
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	file,err := os.OpenFile("nat-server.log", os.O_CREATE | os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	wr := io.MultiWriter(os.Stderr, file)
	Logger = log.New(wr, "net-server ", log.Ldate | log.Ltime | log.Lshortfile)
}