package main

import (
	"fmt"
	"os"
)

func main() {
	filePath := "/var/tmp/file_1"
	f, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(f.Stat())
	content := make([]byte, 1)
	n, err := f.ReadAt(content, 1073741824)
	if err != nil {
		fmt.Println("readAt err is %s", err.Error())
	}
	fmt.Println("read length:%d", n)
	ret, err := f.Seek(0, 0)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(ret)
	fmt.Println(content)
}
