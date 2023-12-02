package Req

import (
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Connect struct {
	req *http.Request
	res *http.Response
}

// *http.Request 슬라이스를 받아 고루틴으로 요청 실행 후, Connect 구조체에 Request, Response를 채워 반환
func Req(reqs []*http.Request, header string) []Connect {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Recovering from panic:", r)
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())
	wg := new(sync.WaitGroup)

	// Request에 대한 Respone 쌍을 담을 Connect 구조체 슬라이스, mill로 초기화
	conns := make([]Connect, len(reqs))

	// 인증 값 넣을 헤더가 있을 경우 대입, 없을 경우 추가
	headerSet := strings.Split(header, ":")
	headerName := strings.Trim(headerSet[0], " ")
	headerData := strings.Trim(headerSet[1], " ")

	for i := 0; i < len(reqs); i++ {
		if _, ok := reqs[i].Header[headerName]; ok {
			reqs[i].Header[headerName] = []string{headerData}
		} else {
			reqs[i].Header.Add(headerName, headerData)
		}
	}

	// HTTP 연결 수 제한 설정
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 10
	t.MaxConnsPerHost = 5
	t.MaxIdleConnsPerHost = 10

	// 클라이언트 생성
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: t,
	}

	for i := 0; i < len(conns); i++ {
		wg.Add(1)
		go func(i int) {
			resp, err := client.Do(reqs[i])
			if err != nil {
				panic(err)
			}
			resp.Body.Close()

			conns[i] = Connect{reqs[i], resp}
			wg.Done()
		}(i)
	}

	wg.Wait()
	return conns
}
