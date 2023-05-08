package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheeyeo/AWS_S3_PIPES/files"
	"github.com/schollz/progressbar/v3"
)

// GenerateSlice returns a data slice
func GenerateSlice(start, end, step int) []int {
	if step <= 0 || end < start {
		return []int{}
	}
	s := make([]int, 0, 1+(end-start)/step)
	for start <= end {
		s = append(s, start)
		start += step
	}
	return s
}

// UploadFile tries to open a given pipefile and copies the orig file to it
func UploadFile(filePath string, pipeFile *os.File) error {
	orig, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer orig.Close()

	fileSize, err := files.GetLocalFileSize(filePath)
	if err != nil {
		return err
	}

	uploadMsg := fmt.Sprintf("Uploading %s", filepath.Base(filePath))
	bar := progressbar.DefaultBytes(
		fileSize,
		uploadMsg,
	)

	_, err = io.Copy(io.MultiWriter(pipeFile, bar), orig)
	if err != nil {
		return err
	}

	defer pipeFile.Close()
	return nil
}

func main() {
	var pipe string
	var fileName string

	flag.StringVar(&pipe, "p", "", "Pipe name.")
	flag.StringVar(&fileName, "f", "", "File to upload")
	flag.Parse()

	pipeFile, err := os.OpenFile(pipe, os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	fmt.Printf("FILE: %v\n", fileName)
	if len(fileName) == 0 {
		data := GenerateSlice(1, 20000, 1)
		for i, x := range data {
			str := fmt.Sprintf("IDX: %d, VAL: %d\n", i, x)
			d := []byte(str)
			pipeFile.Write(d)
		}
	} else {
		err = UploadFile(fileName, pipeFile)
		if err != nil {
			panic(err)
		}
	}

	defer pipeFile.Close()
}
