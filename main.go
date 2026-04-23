package main

import (
	"bufio"
	"fmt"
	"net"
	"redis/handler"
	"redis/resp"
	"strings"
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
		command := strings.ToUpper(string(val.Array[0].Bulk))
		args := val.Array[1:]

		handler := handler.Handlers[command]
		if handler != nil {
			result := handler(args)
			writer.Write(result)
		} else {
			writer.Write(resp.Value{Type: "error", Str: "ERR unknown command '" + command + "'"})
		}
		bufWriter.Flush()
	}
}
