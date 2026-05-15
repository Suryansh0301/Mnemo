package main

import (
	"log/slog"
	"net"
	"sync/atomic"

	"github.com/suryansh0301/Mnemo/internal/core/datastore"
)

const (
	MaxClients = 100000
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	slog.SetLogLoggerLevel(-4)
	slog.Debug("Listening on port 6379")

	exec := startExecutor()
	var totalClients atomic.Int64

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		accepted := false
	inner:
		for {
			current := totalClients.Load()
			if current >= MaxClients {
				conn.Write([]byte("-ERR max number of clients reached\r\n"))
				conn.Close()
				break inner
			}
			if totalClients.CompareAndSwap(current, current+1) {
				accepted = true
				break
			}
		}

		if !accepted {
			continue
		}

		client := newClient(conn)
		go client.handleConnection(exec, &totalClients)
	}
}

func startExecutor() *datastore.Executor {
	exec := datastore.NewExecutor()

	go func() {
		for value := range exec.ExecutorChan {
			response := exec.Execute(value.Command)

			select {
			case value.ResponseChan <- response:
			default:
				slog.Debug("response channel full, dropping response")
			}
		}
	}()

	return exec
}
