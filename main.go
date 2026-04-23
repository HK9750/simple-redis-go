package main

import (
	"bufio"
	"fmt"
	"net"
	"redis/resp"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on :6379")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Handle client connection
	resp := resp.NewResp(bufio.NewReader(conn))
	fmt.Println(resp.Read())
	conn.Write([]byte("+OK\r\n"))
}
