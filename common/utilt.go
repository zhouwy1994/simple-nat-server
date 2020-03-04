package common

import (
	"math/rand"
	"net"
	"strconv"
	"time"
)

func GetAvailableListener() (listener *net.Listener,port int) {
	for {
		port := rand.Intn(1110) + 3334
		listener, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
		if err != nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		return &listener,port
	}
}
