package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// listen 포트
const portNum string = ":8088"

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")

	//모든 헤더 콘솔에 출력
	fmt.Printf("Request at %v\n", time.Now())
	for k, v := range r.Header {
		fmt.Printf("%v: %v\n", k, v)
	}

	fmt.Println("homepage called!")
}

func Info(w http.ResponseWriter, r *http.Request) {
	dt := time.Now()
	fmt.Fprintf(w, "Info page\n")
	fmt.Fprintf(w, "Current date and time is: %s \n", dt.String())

	//모든 헤더 콘솔에 출력
	fmt.Printf("Request at %v\n", time.Now())
	for k, v := range r.Header {
		fmt.Printf("%v: %v\n", k, v)
	}

	fmt.Println("Info & Status page called!")
}

func main() {
	log.Println("Starting our simple http server.")

	//각 경로별 실행할 함수
	http.HandleFunc("/", Home)
	http.HandleFunc("/info", Info)
	http.HandleFunc("/status", Info)

	log.Println("Started on port", portNum)
	fmt.Println("To close connection CTRL+C :D")

	// Spinning up the server.
	err := http.ListenAndServe(portNum, nil)
	if err != nil {
		log.Fatal(err)
	}
}
