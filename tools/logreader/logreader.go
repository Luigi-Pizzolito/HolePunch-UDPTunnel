package main

import (
	"fmt"
	"flag"
	"os"
	
	"k8s.io/utils/inotify"
)

var (
	watchFile string
)

func init() {
	flag.StringVar(&watchFile, "f", "file", "File to watch for changes.")
	flag.Parse()
	// fmt.Println(watchFile)
	if len(os.Args[1:]) != 2 {
		fmt.Fprintf(os.Stderr, "missing required -f argument/flag\n")
		os.Exit(2)
	}
}

func main() {
	var filePos int64

	file, err := os.Open(watchFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	filePos = fileInfo.Size()

	watcher, err := inotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	err = watcher.Watch(watchFile)
	if err != nil {
		panic(err)
	}
	for {
		select {
		case ev := <-watcher.Event:
			// fmt.Println("event:", ev.Mask)

			switch ev.Mask {
			case 16:
				changes := ""
				changes, filePos, err = readFileFromPosition(file, filePos)
				if err != nil {
					panic(err)
				}
				fmt.Print(changes)
			}

		case err := <-watcher.Error:
			fmt.Println("error:", err)
		}
	}
}

func readFileFromPosition(file *os.File, startPos int64) (string, int64, error) {
	// file, err := os.Open(filename)
	// if err != nil {
	// 	return "", 0, err
	// }
	// defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", 0, err
	}

	fileSize := fileInfo.Size()
	if startPos < 0 || startPos > fileSize {
		return "", 0, fmt.Errorf("invalid start position")
	}
	if startPos == fileSize {
		return "", startPos, nil
	}

	_, seekErr := file.Seek(startPos, 0)
	if seekErr != nil {
		return "", 0, seekErr
	}

	bufferSize := int(fileSize - startPos)
	content := make([]byte, bufferSize)

	bytesRead, readErr := file.Read(content)
	if readErr != nil {
		return "", 0, readErr
	}

	endPos := startPos + int64(bytesRead)

	return string(content), endPos, nil
}