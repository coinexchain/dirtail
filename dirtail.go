package dirtail

import (
	"os"
	"bufio"
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

func (dt *DirTail) consume(consumeFunc func(line string, fileNum uint32, offset uint32)) bool {
	for {
		path := fmt.Sprintf("%s/%s%d%s",dt.dirName, dt.filePrefix, dt.fileNum, dt.fileSuffix)
		file, err := os.Open(path)
		if err != nil {
			break
		}
		log.Printf("found file %s offset %d\n", path, dt.fileOffset)

		file.Seek(int64(dt.fileOffset), 0)

		scanner := bufio.NewScanner(file)
		var buf [128]byte
		scanner.Buffer(buf[:], 1024 * 1024 * 64)
		for scanner.Scan() {
			line := scanner.Text()
			dt.fileOffset = dt.fileOffset + uint32(len(line)) + 2 // "\r\n"
			consumeFunc(line, dt.fileNum, dt.fileOffset)
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

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
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


