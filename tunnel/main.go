// REFERENCE
// Go generate RSA https://gist.github.com/sdorra/1c95de8cb80da31610d2ad767cd6f251
// Go Resty https://github.com/go-resty/resty
package main

import (
	"fmt"
	"net/http"
	"time"

	tunnel "github.com/labstack/tunnel-client"
	"github.com/qiniu/log"
	resty "gopkg.in/resty.v1"
)

func main() {
	c := &tunnel.Configuration{
		Host:       "labstack.me:22",
		RemoteHost: "0.0.0.0",
		RemotePort: 8000,
		Channel:    make(chan int),
	}
	c.TargetHost = "127.0.0.1"
	c.TargetPort = 6174

	// Ref: https://github.com/labstack/tunnel-client/blob/master/cmd/root.go
	res, err := resty.R().
		SetAuthToken("hello world").
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "labstack/tunnel").
		Get("http://httpbin.org/get")
	if err != nil {
		log.Fatalf("request err: %v", err)
	} else if res.StatusCode() != http.StatusOK {
		log.Fatalf("request status code not 200, receive %d", res.StatusCode())
	}
	fmt.Printf("Response Body: %v", res.String())

CREATE:
	go tunnel.Create(c)
	event := <-c.Channel
	if event == tunnel.EventReconnect {
		log.Info("trying to reconnect")
		time.Sleep(1 * time.Second)
		goto CREATE
	}
}
