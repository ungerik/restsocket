package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	addr = flag.String("addr", ":80", "")
)

type TCP struct {
	Name         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Socket       *net.TCPConn
}

func (tcp *TCP) Register() {
	if !strings.HasPrefix(tcp.Name, "/") {
		tcp.Name = "/" + tcp.Name
	}
	if !strings.HasSuffix(tcp.Name, "/") {
		tcp.Name += "/"
	}

	http.HandleFunc(tcp.Name+"read/byte.bin", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.ReadByteBin(response, request)
	})
	http.HandleFunc(tcp.Name+"read/byte.txt", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.ReadByteTxt(response, request)
	})
	http.HandleFunc(tcp.Name+"read/bytes.bin", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.ReadBytesBin(response, request)
	})
	http.HandleFunc(tcp.Name+"read/base64.txt", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.ReadBase64Txt(response, request)
	})
	http.HandleFunc(tcp.Name+"read/array.json", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.ReadArrayJSON(response, request)
	})
	http.HandleFunc(tcp.Name+"read/text.json", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.ReadTextJSON(response, request)
	})

	http.HandleFunc(tcp.Name+"write/bytes.bin", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.WriteBytesBin(response, request)
	})
	http.HandleFunc(tcp.Name+"write/base64.txt", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tcp.WriteBytesBin(response, request)
	})

}

func (tcp *TCP) read() ([]byte, error) {
	buf := make([]byte, 1<<16)
	n, err := tcp.Socket.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

type Error struct {
	Error error
}

func (tcp *TCP) respondWithError(response http.ResponseWriter, err error) {
	response.WriteHeader(500)
	response.Write([]byte(err.Error()))
	// data, _ := json.Marshal(Error{err})
	// response.Header().Set("Content-Type", "application/json")
	// response.Header().Set("Content-Length", strconv.Itoa(len(data)))
	// response.Write(data)
}

func (tcp *TCP) ReadByteBin(response http.ResponseWriter, request *http.Request) {
	data := make([]byte, 1)
	n, err := tcp.Socket.Read(data)
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	data = data[:n]
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Content-Length", strconv.Itoa(len(data)))
	response.Write(data)
}

func (tcp *TCP) ReadByteTxt(response http.ResponseWriter, request *http.Request) {
	data := make([]byte, 1)
	n, err := tcp.Socket.Read(data)
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	var txt string
	if n > 0 {
		txt = strconv.Itoa(n)
	}
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Content-Length", strconv.Itoa(len(txt)))
	response.Write([]byte(txt))
}

func (tcp *TCP) ReadBytesBin(response http.ResponseWriter, request *http.Request) {
	data, err := tcp.read()
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	response.Header().Set("Content-Type", "application/octet-stream")
	response.Header().Set("Content-Length", strconv.Itoa(len(data)))
	response.Write(data)
}

func (tcp *TCP) ReadBase64Txt(response http.ResponseWriter, request *http.Request) {
	data, err := tcp.read()
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	encoder := base64.NewEncoder(base64.StdEncoding, response)
	response.Header().Set("Content-Type", "text/plain")
	response.Header().Set("Content-Length", strconv.Itoa(base64.StdEncoding.EncodedLen(len(data))))
	encoder.Write(data)
}

func (tcp *TCP) ReadArrayJSON(response http.ResponseWriter, request *http.Request) {
	data, err := tcp.read()
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, b := range data {
		if i > 0 {
			buf.WriteByte(',')
			fmt.Fprintf(&buf, "%d", b)
		}
	}
	buf.WriteByte(']')
	data = buf.Bytes()
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Content-Length", strconv.Itoa(len(data)))
	response.Write(data)
}

type Text struct {
	Text string
}

func (tcp *TCP) ReadTextJSON(response http.ResponseWriter, request *http.Request) {
	data, err := tcp.read()
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	data, _ = json.Marshal(Text{string(data)})
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Content-Length", strconv.Itoa(len(data)))
	response.Write(data)
}

func (tcp *TCP) write(data []byte, response http.ResponseWriter) {
	n, err := tcp.Socket.Write(data)
	fmt.Fprint(response, n)
	if err != nil {
		tcp.respondWithError(response, err)
	}
}

func (tcp *TCP) WriteBytesBin(response http.ResponseWriter, request *http.Request) {
	data, err := ioutil.ReadAll(request.Body)
	request.Body.Close()
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	tcp.write(data, response)
}

func (tcp *TCP) WriteBase64Txt(response http.ResponseWriter, request *http.Request) {
	data, err := ioutil.ReadAll(request.Body)
	request.Body.Close()
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	data, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		tcp.respondWithError(response, err)
		return
	}
	tcp.write(data, response)
}

func main() {
	http.ListenAndServe(*addr, nil)
}
