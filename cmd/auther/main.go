package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

type connect struct {
	req *http.Request
	res *http.Response
}

type Items struct {
	XMLName     xml.Name `xml:"items"`
	Text        string   `xml:",chardata"`
	BurpVersion string   `xml:"burpVersion,attr"`
	ExportTime  string   `xml:"exportTime,attr"`
	Item        []struct {
		Text string `xml:",chardata"`
		Time string `xml:"time"`
		URL  string `xml:"url"`
		Host struct {
			Text string `xml:",chardata"`
			Ip   string `xml:"ip,attr"`
		} `xml:"host"`
		Port      string `xml:"port"`
		Protocol  string `xml:"protocol"`
		Method    string `xml:"method"`
		Path      string `xml:"path"`
		Extension string `xml:"extension"`
		Request   struct {
			Text   string `xml:",chardata"`
			Base64 string `xml:"base64,attr"`
		} `xml:"request"`
		Status         string `xml:"status"`
		Responselength string `xml:"responselength"`
		Mimetype       string `xml:"mimetype"`
		Response       struct {
			Text   string `xml:",chardata"`
			Base64 string `xml:"base64,attr"`
		} `xml:"response"`
		Comment string `xml:"comment"`
	} `xml:"item"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//var wg = new(sync.WaitGroup)

	argFile := flag.String("f", "", "Add XML file containing Burp Suite history")
	authInput := flag.String("header", "foo: bar", "Add header to auth testing.")
	flag.Parse()

	//Burp suite 히스토리 파일 파싱
	connects := parseBurpHistoryXML(*argFile)

	// requestApi 함수로부터 데이터를 받을 채널
	ch := make(chan connect, len(connects))

	// 파싱하여 얻은 Request 정보로 실제 요청을 수행 (고루틴)
	// for idx, conn := range connects { //range로 고루틴을 실행하면 마지막 인자만 실행됨..왜 그렇지?
	for i := 0; i < len(connects); i++ {
		if connects[i].res == nil {
			go requestApi(connects[i], *authInput, ch)
			connects[i] = <-ch
		}
	}

	for i := 0; i < len(connects); i++ {
		if connects[i].res != nil {
			fmt.Printf("[%s] (%s) %s\n", connects[i].res.Status, connects[i].req.Method, connects[i].req.URL)
		}
	}

	fmt.Printf("%d request are processed!\n", len(connects))

}

// Burp suite 히스토리 파일 파싱
func parseBurpHistoryXML(file string) []connect {
	// Burp History에서 추출한 Request를 저장할 connect 구조체 슬라이스
	connects := []connect{}

	xmlFile, err := os.Open(file)
	if err != nil {
		log.Fatal("[!] Input correct file path, Please.")
		os.Exit(0)
	}

	defer xmlFile.Close()

	fmt.Printf("[*] Successfully Opened %s\n", file)

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Fatal("[!] Cannot read XML.")
		os.Exit(0)
	}

	// Burp History XML이 파싱되어 저장될 구조체
	var items Items
	xml.Unmarshal(byteValue, &items)

	fmt.Println("[*] URLs in file")
	//파싱된 XML에서 Request가 있는 item 요소만 추출
	for i := 0; i < len(items.Item); i++ {
		data, err := base64.StdEncoding.DecodeString(items.Item[i].Request.Text)

		if err != nil {
			log.Fatal("[!] Cannot decode contents in file.")
			os.Exit(0)
		}

		// []byte를 bufio.Reader로 형 변환
		reqData := bufio.NewReader(bytes.NewReader(data))

		//ReadRequest로 HTTP Request 패킷을 파싱하여 requst 객체로 리턴
		req, err := http.ReadRequest(reqData)
		if err != nil {
			log.Fatal("[!] Cannot parse request in file.")
			os.Exit(0)
		}

		req.RequestURI = ""

		req.URL, err = url.Parse(items.Item[i].URL)
		if err != nil {
			log.Fatal("[!] Cannot parse request in file.")
			os.Exit(0)
		}

		connects = append(connects, connect{req, nil})

		fmt.Printf("    (%s) %s\n", connects[i].req.Method, req.URL)
	}
	fmt.Println("-------------------------------------------")

	return connects
}

// connet{http.Request, Respone} 구조체를 받아 요청을 실행한 후 Response를 채워 반환, 고루틴으로 실행됨
func requestApi(conn connect, auth string, ch chan connect) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovering from panic:", r)
			ch <- conn
		}
	}()

	headerSet := strings.Split(auth, ":")
	conn.req.Header.Add(strings.Trim(headerSet[0], " "), strings.Trim(headerSet[1], " "))

	// HTTP 연결 수 제한 설정
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 10
	t.MaxConnsPerHost = 5
	t.MaxIdleConnsPerHost = 10

	// 제한 적용하여 클라이언트 생성
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: t,
	}

	resp, err := client.Do(conn.req)
	if err != nil {
		panic(err)
	}

	resp.Body.Close()

	conn.res = resp
	ch <- conn
}
