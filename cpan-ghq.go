package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile("^([^0-9\\s\\[][^\\s\\[]*)?(\\[[0-9]+\\])?$")

func jsonpath(b []byte, jp string, t interface{}) error {
	var v interface{}
	err := json.Unmarshal([]byte(b), &v)
	if err != nil {
		return err
	}
	if jp == "" {
		return errors.New("invalid path")
	}
	for _, token := range strings.Split(jp, "/") {
		sl := re.FindAllStringSubmatch(token, -1)
		if len(sl) == 0 {
			return errors.New("invalid path")
		}
		ss := sl[0]
		if ss[1] != "" {
			v = v.(map[string]interface{})[ss[1]]
		}
		if ss[2] != "" {
			i, err := strconv.Atoi(ss[2][1 : len(ss[2])-1])
			if err != nil {
				return errors.New("invalid path")
			}
			v = v.([]interface{})[i]
		}
	}
	rt := reflect.ValueOf(t).Elem()
	rv := reflect.ValueOf(v)
	rt.Set(rv)
	return nil
}

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
		err = jsonpath(body, "/release/_source/resources/repository/url", &s)
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
			println(url)
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
			println(url)
		}
	}
}
