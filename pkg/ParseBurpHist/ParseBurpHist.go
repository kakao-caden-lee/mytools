package ParseBurpHist

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Burp History XML이 파싱되어 저장될 구조체
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

// Burp suite 히스토리 파일 파싱
func ParseBurpHistory(file string) []*http.Request {
	// Burp History에서 추출한 Request를 저장할 슬라이스
	reqs := []*http.Request{}

	// Burp History XML이 파싱되어 저장될 구조체
	var items Items

	xmlFile, err := os.Open(file)
	if err != nil {
		log.Fatal("[!] Input correct file path, Please.")
		panic(err)
	}

	defer xmlFile.Close()
	fmt.Printf("[*] Successfully Opened %s\n", file)

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Fatal("[!] Cannot read file.")
		panic(err)
	}

	xml.Unmarshal(byteValue, &items)

	fmt.Println("[*] URLs in file ")
	//파싱된 XML에서 Request가 있는 item 요소만 추출
	for i := 0; i < len(items.Item); i++ {
		data, err := base64.StdEncoding.DecodeString(items.Item[i].Request.Text)
		if err != nil {
			log.Fatal("[!] Cannot decode contents in file.")
			panic(err)
		}

		// readRequest 함수 인자로 넣기위해 []byte를 bufio.Reader로 형 변환
		reqData := bufio.NewReader(bytes.NewReader(data))

		//ReadRequest로 HTTP Request 패킷을 파싱하여 requst 객체로 리턴
		req, err := http.ReadRequest(reqData)
		if err != nil {
			log.Fatal("[!] Cannot parse request.")
			panic(err)
		}

		// req (http.Request) 에 URL 경로 값이 아닌, Full URL을 대입
		// req.RequestURI 값이 있을 경우, http.Request.URL 보다 우선해서 쓰이므로 빈 값으로 초기화
		req.RequestURI = ""

		req.URL, err = url.Parse(items.Item[i].URL)
		if err != nil {
			log.Fatal("[!] Cannot parse request in file.")
			panic(err)
		}

		reqs = append(reqs, req)

		fmt.Printf("    (%s) %s\n", reqs[i].Method, req.URL)
	}

	return reqs
}
