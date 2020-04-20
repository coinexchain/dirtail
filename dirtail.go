package dirtail

import (
	"io"
	"os"
	"log"
	"time"
	"fmt"
)

type DirTail struct {
	dirName string
	filePrefix string
	fileSuffix string
	fileNum  uint32
	fileOffset uint32

	stopReq chan bool
	stopAck chan bool
}

func NewDirTail(dirName string,
	filePrefix string,
	fileSuffix string,
	fileNum  uint32,
	fileOffset uint32) *DirTail {
	return &DirTail{
		dirName: dirName,
		filePrefix: filePrefix,
		fileSuffix: fileSuffix,
		fileNum: fileNum,
		fileOffset: fileOffset,
		stopReq: make(chan bool, 1),
		stopAck: make(chan bool, 1),
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (dt *DirTail) Start(interval int64, consumeFunc func(line string, fileNum uint32, offset uint32)) {
	go func() {
		for {
			stop := dt.consume(consumeFunc)
			if stop {
				break
			}
			time.Sleep(time.Duration(interval * int64(time.Millisecond)))
		}
		dt.stopAck <- true
	}()
}

func (dt *DirTail) Stop() {
	dt.stopReq <- true
	<-dt.stopAck
}

func getc(f *os.File) (byte, error) {
	b := make([]byte, 1)
	_, err := f.Read(b)
	return b[0], err
}

func readLine(f *os.File) ([]byte, error) {
	buf := make([]byte, 0, 64*1024)
	for {
		b, err := getc(f)
		if err != nil {
			return nil, err
		}
		if b == byte('\r') {
			b, err := getc(f)
			if b != byte('\n') {
				panic("Not \\r\\n!")
			}
			if err != nil {
				return nil, err
			}
			break
		}
		buf = append(buf, b)
	}
	return buf, nil
}

func (dt *DirTail) consume(consumeFunc func(line string, fileNum uint32, offset uint32)) bool {
	for {
		path := fmt.Sprintf("%s/%s%d%s",dt.dirName, dt.filePrefix, dt.fileNum, dt.fileSuffix)
		file, err := os.Open(path)
		if err != nil {
			break
		}
		log.Printf("found file %s offset %d\n", path, dt.fileOffset)

		file.Seek(int64(dt.fileOffset), 0)

		for {
			line, err := readLine(file)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
				break
			}
			dt.fileOffset = dt.fileOffset + uint32(len(line)) + 2 // "\r\n"
			consumeFunc(string(line), dt.fileNum, dt.fileOffset)
			select {
			case <-dt.stopReq:
				log.Printf("Now Stop DirTail\n")
				file.Close()
				return true
			default:
			}
		}

		select {
		case <-dt.stopReq:
			log.Printf("Now Stop DirTail\n")
			file.Close()
			return true
		default:
		}

		file.Close()

		path = fmt.Sprintf("%s/%s%d%s",dt.dirName, dt.filePrefix, dt.fileNum + 1, dt.fileSuffix)
		if fileExists(path) {
			dt.fileNum++
			dt.fileOffset = 0
		} else {
			break
		}
	}
	return false
}


