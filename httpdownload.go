package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
)

var config struct {
	url   string
	image string
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

func isOk(v string) bool {
	return strings.Contains(v, "HTTP/1.1 200 OK")
}

func getFileSize(rawHeader string) string {
	re := regexp.MustCompile(`(?m)Content-Length: (\d+)`)
	for _, v := range strings.Split(rawHeader, "\r\n") {
		if strings.Contains(v, "Content-Length") {
			return re.FindStringSubmatch(v)[1]
		}
	}
	return ""
}
func writeFile(s, filename string) {
	os.WriteFile(filename, []byte(s), 0644)
}

func main() {
	flag.StringVar(&config.url, "url", "http://192.168.144.139", "URL Request")
	flag.StringVar(&config.image, "remote-file", "/wp-content/uploads/2022/02/nen.png", "path to image file in local")
	flag.Parse()
	domain := getDomain(config.url)
	conn, err := net.Dial("tcp", domain+":80")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	sendRaw := "GET " + config.image + " HTTP/1.1\r\n" +
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
	splitRaw := strings.Split(raw, "\r\n")
	if isOk(splitRaw[0]) {
		splitFileName := strings.Split(config.image, "/")
		filename := splitFileName[len(splitFileName)-1]
		splitImage := strings.Split(raw, "\r\n\r\n")
		image := splitImage[1]
		writeFile(image, filename)
		fmt.Println("File", filename, "size:", getFileSize(splitImage[0]), "bytes")
	} else {
		log.Println("image error:", splitRaw[0])
	}
}
