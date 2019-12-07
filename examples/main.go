package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/winterssy/sreq"
)

func main() {
	// setQueryParams()
	// setHeaders()
	// setCookies()
	// sendForm()
	// sendJSON()
	// uploadFiles()
	// setBasicAuth()
	// setBearerToken()
	// setProxy()
	// setContext()
	// debug()
}

func setQueryParams() {
	data, err := sreq.
		Get("http://httpbin.org/get",
			sreq.WithQuery(sreq.Params{
				"k1": "v1",
				"k2": "v2",
			}),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func setHeaders() {
	data, err := sreq.
		Get("http://httpbin.org/get",
			sreq.WithHeaders(sreq.Headers{
				"Origin":  "http://httpbin.org",
				"Referer": "http://httpbin.org",
			}),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func setCookies() {
	data, err := sreq.
		Get("http://httpbin.org/cookies",
			sreq.WithCookies(
				&http.Cookie{
					Name:  "n1",
					Value: "v1",
				},
				&http.Cookie{
					Name:  "n2",
					Value: "v2",
				},
			),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func sendForm() {
	data, err := sreq.
		Post("http://httpbin.org/post",
			sreq.WithForm(sreq.Form{
				"k1": "v1",
				"k2": "v2",
			}),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func sendJSON() {
	data, err := sreq.
		Post("http://httpbin.org/post",
			sreq.WithJSON(sreq.JSON{
				"msg": "hello world",
				"num": 2019,
			}, true),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func uploadFiles() {
	data, err := sreq.
		Post("http://httpbin.org/post",
			sreq.WithMultipart(sreq.Files{
				"image1": "./testdata/testimage1.jpg",
				"image2": "./testdata/testimage2.jpg",
			}),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func setBasicAuth() {
	data, err := sreq.
		Get("http://httpbin.org/basic-auth/admin/pass",
			sreq.WithBasicAuth("admin", "pass"),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func setBearerToken() {
	data, err := sreq.
		Get("http://httpbin.org/bearer",
			sreq.WithBearerToken("sreq"),
		).
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func setProxy() {
	client := sreq.New().SetProxyFromURL("socks5://127.0.0.1:1080")
	data, err := client.
		Get("https://api.ipify.org?format=json").
		Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func setContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := sreq.
		Get("http://httpbin.org/delay/10",
			sreq.WithContext(ctx),
		).
		Verbose(ioutil.Discard)
	fmt.Println(err)
}

func debug() {
	err := sreq.
		Get("http://httpbin.org/get").
		Verbose(os.Stdout)
	if err != nil {
		panic(err)
	}
}
