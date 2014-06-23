package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/mattn/go-scan"
)

func getRepositoryUrl(module string) (repositoryUrl string, err error) {
	var s string
	url := fmt.Sprintf("http://api.metacpan.org/v0/module/%s?join=release",
		url.QueryEscape(module))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		var v interface{}
		err = json.Unmarshal(body, &v)
		if err != nil {
			return "", err
		}
		err = scan.ScanTree(v, "/distribution", &s)
		if s == "perl" {
			return "", fmt.Errorf("%q is provided by core module", module)
		}
		err = scan.ScanTree(v, "/release/_source/resources/repository/url", &s)
		if err != nil {
			return "", err
		}

		return s, nil
	} else {
		return "", nil
	}
}

func readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func main() {
	if len(os.Args) > 1 {
		for _, module := range os.Args[1:] {
			url, err := getRepositoryUrl(module)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(url)
		}
	} else {
		bio := bufio.NewReader(os.Stdin)
		for {
			line, err := readln(bio)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			url, err := getRepositoryUrl(string(line[:]))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(url)
		}
	}
}
