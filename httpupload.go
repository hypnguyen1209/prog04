package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var config struct {
	url    string
	user   string
	pass   string
	image  string
	domain string
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

func getCookieJar(s []string) string {
	cookieSlice := []string{}
	re := regexp.MustCompile(`(?m)Set-Cookie: (.+?);`)
	for _, v := range s {
		if strings.HasPrefix(v, "Set-Cookie") {
			cookieSlice = append(cookieSlice, re.FindStringSubmatch(v)[1])
		}
	}
	return strings.Join(cookieSlice, "; ")
}

func readFile(src string) []byte {
	result, err := os.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func getFileName(src string) string {
	splitName := strings.Split(src, "/")
	return splitName[len(splitName)-1]
}

func getContentType(src string) string {
	splitName := strings.Split(src, "/")
	extension := strings.Split(splitName[len(splitName)-1], ".")
	return fmt.Sprintf("image/%s", extension[len(extension)-1])
}

func getWpNonce(jar string) string {
	re := regexp.MustCompile(`(?m)\"_wpnonce\":\"(.+?)\"}}`)
	conn, err := net.Dial("tcp", config.domain+":80")
	if err != nil {
		log.Println(err)
		return ""
	}
	defer conn.Close()
	sendRaw := "GET /wp-admin/upload.php HTTP/1.1\r\n" +
		"Host: " + config.domain + "\r\n" +
		"User-Agent: VCS/1.0\r\n" +
		"Accept: */*\r\n" +
		"Cookie: " + jar + "\r\n" +
		"\r\n"
	conn.Write([]byte(sendRaw))
	res, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}
	raw := string(res)
	return re.FindStringSubmatch(raw)[1]
}

func getImage(v string) string {
	re := regexp.MustCompile(`(?m)\"url\":\"(.+?)\",\"link`)
	return strings.Replace(re.FindStringSubmatch(v)[1], "\\", "", -1)
}

func isOk(v string) bool {
	return strings.Contains(v, "HTTP/1.1 200 OK")
}

func sendUpload(jar string) {
	imageRaw := readFile(config.image)
	conn, err := net.Dial("tcp", config.domain+":80")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	rawBody := "------WebKitFormBoundary\r\n" +
		"Content-Disposition: form-data; name=\"name\"\r\n\r\n" +
		getFileName(config.image) + "\r\n" +
		"------WebKitFormBoundary\r\n" +
		"Content-Disposition: form-data; name=\"action\"\r\n\r\n" +
		"upload-attachment\r\n" +
		"------WebKitFormBoundary\r\n" +
		"Content-Disposition: form-data; name=\"_wpnonce\"\r\n\r\n" +
		getWpNonce(jar) + "\r\n" +
		"------WebKitFormBoundary\r\n" +
		"Content-Disposition: form-data; name=\"async-upload\"; filename=\"" + getFileName(config.image) + "\"\r\n" +
		"Content-Type: " + getContentType(config.image) + "\r\n" +
		"\r\n" +
		string(imageRaw) + "\r\n" +
		"------WebKitFormBoundary--"
	lengthBody := len(rawBody)
	sendRaw := "POST /wp-admin/async-upload.php HTTP/1.1\r\n" +
		"Host: " + config.domain + "\r\n" +
		"User-Agent: VCS/1.0\r\n" +
		"Accept: */*\r\n" +
		"Accept-Encoding: deflate\r\n" +
		"Cookie: " + jar + "\r\n" +
		"Content-Type: multipart/form-data; boundary=----WebKitFormBoundary\r\n" +
		"Content-Length: " + strconv.Itoa(lengthBody) + "\r\n" +
		"Connection: keep-alive\r\n" +
		"\r\n" +
		rawBody
	conn.Write([]byte(sendRaw))
	res, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}
	raw := string(res)
	if isOk(strings.Split(raw, "\r\n")[0]) {
		imageLink := getImage(raw)
		fmt.Println("Upload was successful. URL:", imageLink)
	} else {
		fmt.Println("Upload failed")
	}
}

func writeLog(s string) {
	os.WriteFile("log.txt", []byte(s), 0644)
}

func main() {
	flag.StringVar(&config.url, "url", "http://192.168.144.139", "URL Request")
	flag.StringVar(&config.user, "user", "test", "Username")
	flag.StringVar(&config.pass, "password", "test123QWE@AD", "Password")
	flag.StringVar(&config.image, "local-file", "/home/hyp/Pictures/cat2.png", "path to image file in local")
	flag.Parse()
	config.domain = getDomain(config.url)
	conn, err := net.Dial("tcp", config.domain+":80")
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	postBody := fmt.Sprintf("log=%s&pwd=%s&wp-submit=Log+In", url.QueryEscape(config.user), url.QueryEscape(config.pass))
	lengthBody := strconv.Itoa(len(postBody))
	sendRaw := "POST /wp-login.php HTTP/1.1\r\n" +
		"Host: " + config.domain + "\r\n" +
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
		cookieJar := getCookieJar(splitRaw)
		sendUpload(cookieJar)

	} else {
		fmt.Printf("User %s login failed", config.user)
	}
}
