package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var config struct {
	url  string
	user string
	pass string
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

func checkRedirect(v string) bool {
	return strings.Contains(v, "HTTP/1.1 302 Found")
}

func checkLogin(s []string) bool {
	re := regexp.MustCompile(`(?m)Location: [\S]+\/wp-admin\/`)
	for _, v := range s {
		if re.MatchString(v) {
			return true
		}
	}
	return false
}

func main() {
	flag.StringVar(&config.url, "url", "http://192.168.144.139", "URL Request")
	flag.StringVar(&config.user, "user", "test", "Username")
	flag.StringVar(&config.pass, "password", "test123QWE@AD", "Password")
	flag.Parse()
	domain := getDomain(config.url)
	conn, err := net.Dial("tcp", domain+":80")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	postBody := fmt.Sprintf("log=%s&pwd=%s&wp-submit=Log+In", url.QueryEscape(config.user), url.QueryEscape(config.pass))
	lengthBody := strconv.Itoa(len(postBody))
	sendRaw := "POST /wp-login.php HTTP/1.1\r\n" +
		"Host: " + domain + "\r\n" +
		"User-Agent: VCS/1.0\r\n" +
		"Accept: */*\r\n" +
		"Content-Length: " + lengthBody + "\r\n" +
		"Content-Type: application/x-www-form-urlencoded\r\n" +
		"\r\n" +
		postBody
	conn.Write([]byte(sendRaw))
	res, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}
	raw := string(res)
	splitRaw := strings.Split(raw, "\r\n")
	if checkRedirect(splitRaw[0]) && checkLogin(splitRaw) {
		fmt.Printf("User %s login successfully", config.user)
	} else {
		fmt.Printf("User %s login failed", config.user)
	}
}
