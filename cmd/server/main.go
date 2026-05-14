package main

import (
	"log/slog"
	"net"

	"github.com/suryansh0301/mini-redis/internal/core/datastore"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	slog.SetLogLoggerLevel(-4)
	slog.Debug("Listening on port 6379")

	exec := datastore.NewExecutor()

	go func() {
		for value := range exec.ExecutorChan {
			response := exec.Execute(value.Command)

			value.ResponseChan <- response
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		client := newClient(conn)
		go client.handleConnection(conn, exec)
	}
}
