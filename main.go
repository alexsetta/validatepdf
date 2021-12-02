package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type result struct {
	err  error
	file string
}

var pdftotext string

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: pdf2txt <threads> <path_pdftotext.exe> <pdf_dir>")
		os.Exit(1)
	}

	threads, _ := strconv.ParseInt(os.Args[1], 10, 32)

	pdftotext = os.Args[2] + "/pdftotext.exe"
	if _, err := os.Stat(pdftotext); os.IsNotExist(err) {
		log.Fatal("pdftotext.exe not found")
	}

	files, err := getFiles(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	fmt.Println("Files:", len(files))
	fmt.Println("Threads:", threads)

	sem := make(chan struct{}, threads)
	c := make(chan result)

	for _, file := range files {
		sem <- struct{}{}
		go func(file string) {
			var res result
			res.file = file
			res.err = checkPdf(file)
			<-sem
			c <- res
		}(file)
	}

	for range files {
		res := <-c
		if res.err != nil {
			log.Println(res.file)
		}
	}

	fmt.Println("time:", time.Since(start))
}

// get all files in directory
func getFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, file := range files {
		paths = append(paths, dir+"/"+file.Name())
	}
	return paths, nil
}

func checkPdf(path string) error {
	cmd := exec.Command(pdftotext, path, "nul")
	stdErr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	if len(stdErr) > 0 {
		return fmt.Errorf("%s", stdErr)
	}

	return nil
}
