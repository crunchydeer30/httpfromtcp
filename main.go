package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var ErrSomethingWentWrong = errors.New("something went wrong")

func main() {
	filepath := "messages.txt"

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("could not open %s: %s\n", filepath, err)
	}

	ch := getLinesChannel(file)

	for line := range ch {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer f.Close()

		currentLine := ""

		for {
			buf := make([]byte, 8)
			n, err := f.Read(buf)

			if n > 0 {
				parts := strings.Split(string(buf[:n]), "\n")
				for i := 0; i < len(parts)-1; i++ {
					lines <- currentLine + parts[i]
					currentLine = ""
				}
				currentLine += parts[len(parts)-1]
			}

			if err == io.EOF {
				if currentLine != "" {
					lines <- currentLine
				}
				break
			}

			if err != nil {
				break
			}
		}
	}()

	return lines
}
