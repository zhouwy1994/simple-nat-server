package client

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"
)

func Start() {
	serverAddr := "112.124.109.39"
	localAddr := "127.0.0.1"

	for {
		conn, err := net.Dial("tcp4", serverAddr+":3333")
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}

		conn.Write([]byte("register\r\n22\r\n"))

		for {
			data, _, err := bufio.NewReader(conn).ReadLine()
			if err != nil {
				break
			}

			index := strings.LastIndex(string(data), "-")
			port, err := strconv.Atoi(string(data)[index+1:])
			if err != nil {
				continue
			}

			destConn, err := net.Dial("tcp4", localAddr+":"+strconv.Itoa(port))
			if err != nil {
				continue
			}

			srcConn, err := net.Dial("tcp4", serverAddr+":3334")
			if err != nil {
				destConn.Close()
				continue
			}

			srcConn.Write([]byte(fmt.Sprintf("dataer\r\n%s\r\n", string(data))))

			trr := io.TeeReader(destConn, srcConn)
			trw := io.TeeReader(srcConn, destConn)

			go func() {
				defer destConn.Close()
				ioutil.ReadAll(trr)
			}()

			go func() {
				defer srcConn.Close()
				ioutil.ReadAll(trw)
			}()
		}

		conn.Close()
	}
}