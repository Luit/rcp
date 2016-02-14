// Copyright © 2016 Luit van Drongelen <luit@luit.eu>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package server // import "luit.eu/rcp/server"

import (
	"io"
	"log"
	"net"
	"strings"

	"luit.eu/resp"
)

func connect(addr string) (*net.TCPConn, error) {
	raddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, raddr)
}

func Dumb(clientConn *net.TCPConn) {
	var backendAddr = "127.0.0.1:30001"
	var backendConn *net.TCPConn
	var backend *resp.Reader
	defer func() {
		clientConn.Close()
		if backendConn != nil {
			backendConn.Close()
		}
	}()
	client := resp.NewCommandReader(clientConn)
	for {
		data, parts, err := client.Read()
		if err != nil {
			if v, ok := err.(resp.Error); ok {
				clientConn.Write(v.RESP())
			} else {
				io.WriteString(clientConn, "-ERR unexpected error reading command\r\n")
			}
			return
		}
	roundtrip:
		if backendConn == nil {
			backendConn, err = connect(backendAddr)
			if err != nil {
				io.WriteString(clientConn, "-ERR backend dial error\r\n")
				return
			}
			log.Printf("connected to server %s\n", backendAddr)
			backend = resp.NewReader(backendConn)
		}
		_ = parts // Using this later
		_, err = backendConn.Write(data)
		if err != nil {
			io.WriteString(clientConn, "-ERR backend write error\r\n")
			return
		}
		var response []byte
		response, err = backend.Read()
		if err != nil {
			if v, ok := err.(resp.Error); ok {
				clientConn.Write(v.RESP())
			} else {
				io.WriteString(clientConn, "-ERR backend read error\r\n")
			}
			return
		}
		if response[0] == '-' {
			respError := resp.ParseError(response)
			prefix := respError.Prefix()
			switch prefix {
			case "ASK":
				// TODO: Do stuff
				log.Fatal(respError.Error())
			case "MOVED":
				message := respError.Error()
				i := strings.LastIndexByte(message, ' ')
				if i == -1 {
					io.WriteString(clientConn, "-ERR unexpected -MOVED string\r\n")
					return
				}
				backendAddr = message[i+1:]
				backendConn.Close()
				backendConn = nil
				goto roundtrip
			}
		}
		_, err = clientConn.Write(response)
		if err != nil {
			return
		}
	}
}
