package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strings"
)

var config struct {
	url string
}

func getDomain(url string) string {
	re := regexp.MustCompile(`(?m)^[http(s?)]+://([^/]\S+)$`)
	domain := re.FindStringSubmatch(url)[1]
	if strings.HasSuffix(domain, "/") {
		domain = domain[:len(domain)-1]
	}
	splitURL := strings.Split(domain, "/")[0]
	return splitURL
}

func getTitle(raw string) string {
	re := regexp.MustCompile(`(?m)<title>(.+?)</title>`)
	title := re.FindStringSubmatch(raw)[1]
	return title
}

func main() {
	flag.StringVar(&config.url, "url", "http://192.168.144.139", "URL Request")
	flag.Parse()
	domain := getDomain(config.url)
	conn, err := net.Dial("tcp", domain+":80")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	sendRaw := "GET / HTTP/1.1\r\n" +
		"Host: " + domain + "\r\n" +
		"User-Agent: VCS/1.0\r\n" +
		"Accept: */*\r\n" +
		"\r\n"
	conn.Write([]byte(sendRaw))
	res, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}
	raw := string(res)
	title := getTitle(raw)
	fmt.Println("Title:", title)
}
