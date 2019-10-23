package dirtail

import (
	"fmt"
	"os"
	"testing"
	"time"

	//"github.com/stretchr/testify/require"
)

func Test1(t *testing.T) {
	os.MkdirAll("./test", os.ModePerm)

	//result := make([]string, 0, 100)

	os.Remove("./test/myfile1.log")
	os.Remove("./test/myfile2.log")
	f, err := os.OpenFile("./test/myfile1.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	for i:=0; i<10; i++ {
		f.WriteString("01234567890123456789012345678901234567\r\n")
		f.WriteString("0AB34567890123456789012345678901234567\r\n")
		f.WriteString("01AB4567890123456789012345678901234567\r\n")
		f.WriteString("012AB567890123456789012345678901234567\r\n")
	}
	f.Sync()

	dt := NewDirTail("./test", "myfile", ".log", 1, 0)
	var savedFileNum uint32
	var savedOffset uint32
	dt.Start(500, func(line string, fileNum uint32, offset uint32) {
		savedFileNum = fileNum
		savedOffset = offset
		fmt.Printf("|%d|%d|%s\n", fileNum, offset, line)
	})

	time.Sleep(3 * time.Second)
	for i:=0; i<10; i++ {
		f.WriteString("01234567890123456789012345678901234567\r\n")
		f.WriteString("0CD34567890123456789012345678901234567\r\n")
		f.WriteString("01CD4567890123456789012345678901234567\r\n")
		f.WriteString("012CD567890123456789012345678901234567\r\n")
	}
	f.Sync()

	dt.Stop()

	time.Sleep(3 * time.Second)
	for i:=0; i<10; i++ {
		f.WriteString("01234567890123456789012345678901234567\r\n")
		f.WriteString("0EF34567890123456789012345678901234567\r\n")
		f.WriteString("01EF4567890123456789012345678901234567\r\n")
		f.WriteString("012EF567890123456789012345678901234567\r\n")
	}
	f.Close()

	f, _ = os.OpenFile("./test/myfile2.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	for i:=0; i<10; i++ {
		f.WriteString("01234567890123456789012345678901234567\r\n")
		f.WriteString("0GH34567890123456789012345678901234567\r\n")
		f.WriteString("01GH4567890123456789012345678901234567\r\n")
		f.WriteString("012GH567890123456789012345678901234567\r\n")
	}
	f.Close()

	dt = NewDirTail("./test", "myfile", ".log", savedFileNum, savedOffset)
	dt.Start(500, func(line string, fileNum uint32, offset uint32) {
		savedFileNum = fileNum
		savedOffset = offset
		fmt.Printf("|%d|%d|%s\n", fileNum, offset, line)
	})

	time.Sleep(3 * time.Second)
	dt.Stop()
}

