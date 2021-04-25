package main

import (
	"log"
	"net"
	"net/http"
)

var start byte = 0xef
var end byte = 0xff


func main() {
	http.HandleFunc("/pwd/update", UpdatePwd)
	http.HandleFunc("/pwd/dynamic", DynamicPwd)

	go func() {
		log.Println(http.ListenAndServe(":8081", nil))
	}()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Printf("listener err: %v", err)
	}
	for  {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept err: %v", err)
		}
		log.Println("connect success!")
		go ConnHandle(conn)
	}
}
