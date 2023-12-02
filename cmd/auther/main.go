package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cadenqq/mytools/pkg/ParseBurpHist"
	"github.com/cadenqq/mytools/pkg/Req"
)

func main() {
	file := flag.String("f", "", "Add XML file containing Burp Suite history")
	header := flag.String("h", "", "Add header to auth testing.")
	flag.Parse()

	if file == nil {
		fmt.Fprintf(os.Stderr, "Usage : %s -f <file-path> [-h <string>]\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	//Burp suite 히스토리 파일 파싱
	reqs := ParseBurpHist.ParseBurpHistory(*file)
	fmt.Println("----------------------------------------------------------------------------------------------------------------")

	// 파싱하여 얻은 Request 정보로 실제 요청을 고루틴으로 수행
	conns := Req.ExecReq(reqs, *header)
	fmt.Println("----------------------------------------------------------------------------------------------------------------")

	// 채널로 받은 응답 값 출력
	fmt.Printf("[*] Result of requests\n")
	for i := 0; i < len(conns); i++ {
		if conns[i].Req != nil {
			fmt.Printf("[%s] (%s) %s\n", conns[i].Res.Status, conns[i].Req.Method, conns[i].Req.URL)
		}
	}

	fmt.Printf("[*] %d request are processed!\n", len(conns))
}
