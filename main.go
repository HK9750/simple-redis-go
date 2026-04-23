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

	reader := resp.NewReader(bufio.NewReader(conn))
	bufWriter := bufio.NewWriter(conn)
	writer := resp.NewWriter(bufWriter)

	for {
		val, err := reader.Read()
		if err != nil {
			fmt.Println("Error reading from client:", err)
			return
		}

		fmt.Printf("received: %+v\n", val)

		writer.Write(resp.Value{Type: "string", Str: "OK"})
		bufWriter.Flush()
	}
}
