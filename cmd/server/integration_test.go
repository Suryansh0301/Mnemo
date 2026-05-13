package main

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"strconv"

	"github.com/stretchr/testify/assert"
	"github.com/suryansh0301/mini-redis/internal/core/datastore"
)

// ── Server Setup ──────────────────────────────────────────────────

func startTestServer(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("error in listening to the port: %s", err)
	}

	exec := datastore.NewExecutor()

	go func() {
		for value := range exec.ExecutorChan {
			response := exec.Execute(value.Command)
			value.ResponseChan <- response
		}
	}()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			client := newClient(conn)
			go client.handleConnection(conn, exec)
		}
	}()

	t.Cleanup(func() {
		listener.Close()
		close(exec.ExecutorChan)
	})

	return listener.Addr().String()
}

func dial(t *testing.T, addr string) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	return conn
}

func send(t *testing.T, conn net.Conn, cmd string) string {
	t.Helper()
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	return string(buf[:n])
}

func readUntil(t *testing.T, conn net.Conn, expected string) string {
	t.Helper()
	buf := make([]byte, 512)
	var result string
	for {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		result += string(buf[:n])
		if len(result) >= len(expected) {
			return result
		}
	}
}

func TestIntegrationPing(t *testing.T) {
	addr := startTestServer(t)
	conn := dial(t, addr)
	defer conn.Close()

	resp := send(t, conn, "*1\r\n$4\r\nPING\r\n")
	assert.Equal(t, "+PONG\r\n", resp)
}

func TestIntegrationSetGet(t *testing.T) {
	addr := startTestServer(t)
	conn := dial(t, addr)
	defer conn.Close()

	setResp := send(t, conn, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	assert.Equal(t, "+OK\r\n", setResp)

	getResp := send(t, conn, "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")
	assert.Equal(t, "$3\r\nbar\r\n", getResp)
}

func TestIntegrationGetMissingKey(t *testing.T) {
	addr := startTestServer(t)
	conn := dial(t, addr)
	defer conn.Close()

	resp := send(t, conn, "*2\r\n$3\r\nGET\r\n$7\r\nmissing\r\n")
	assert.Equal(t, "$-1\r\n", resp)
}

func TestIntegrationUnknownCommand(t *testing.T) {
	addr := startTestServer(t)
	conn := dial(t, addr)
	defer conn.Close()

	resp := send(t, conn, "*1\r\n$7\r\nUNKNOWN\r\n")
	assert.Equal(t, "-ERR unknown command 'UNKNOWN'\r\n", resp)
}

func TestIntegrationIncr(t *testing.T) {
	addr := startTestServer(t)
	conn := dial(t, addr)
	defer conn.Close()

	// INCR missing key — starts at 1
	resp := send(t, conn, "*2\r\n$4\r\nINCR\r\n$7\r\ncounter\r\n")
	assert.Equal(t, ":1\r\n", resp)

	// INCR again
	resp = send(t, conn, "*2\r\n$4\r\nINCR\r\n$7\r\ncounter\r\n")
	assert.Equal(t, ":2\r\n", resp)
}

func TestIntegrationPipelining(t *testing.T) {
	addr := startTestServer(t)
	conn := dial(t, addr)
	defer conn.Close()

	pipeline := "*1\r\n$4\r\nPING\r\n*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	_, err := conn.Write([]byte(pipeline))
	if err != nil {
		t.Fatal(err)
	}

	expected := "+PONG\r\n+OK\r\n"
	resp := readUntil(t, conn, expected)
	assert.Equal(t, expected, resp)
}

func TestIntegrationProtocolError(t *testing.T) {
	addr := startTestServer(t)
	conn := dial(t, addr)
	defer conn.Close()

	resp := send(t, conn, "GARBAGE\r\n")
	assert.Equal(t, "-ERR Protocol error\r\n", resp)
}

func TestIntegrationConcurrentClients(t *testing.T) {
	addr := startTestServer(t)

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := dial(t, addr)
			defer conn.Close()

			resp := send(t, conn, "*1\r\n$4\r\nPING\r\n")
			if resp != "+PONG\r\n" {
				errors <- fmt.Errorf("expected +PONG got %q", resp)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestIntegrationConcurrentIsolation(t *testing.T) {
	addr := startTestServer(t)
	var wg sync.WaitGroup

	// Phase 1 — 25 concurrent writers
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := dial(t, addr)
			defer conn.Close()
			key := fmt.Sprintf("key%d", i)
			val := fmt.Sprintf("val%d", i)
			cmd := "*3\r\n$3\r\nSET\r\n$" + strconv.Itoa(len(key)) + "\r\n" + key + "\r\n$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
			resp := send(t, conn, cmd)
			assert.Equal(t, "+OK\r\n", resp, "SET failed for key%d", i)
		}(i)
	}
	wg.Wait()

	// Phase 2 — 25 concurrent readers
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := dial(t, addr)
			defer conn.Close()
			key := fmt.Sprintf("key%d", i)
			val := fmt.Sprintf("val%d", i)
			cmd := "*2\r\n$3\r\nGET\r\n$" + strconv.Itoa(len(key)) + "\r\n" + key + "\r\n"
			resp := send(t, conn, cmd)
			expected := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
			assert.Equal(t, expected, resp, "GET wrong value for key%d", i)
		}(i)
	}
	wg.Wait()
}

func TestIntegrationGracefulDisconnect(t *testing.T) {
	addr := startTestServer(t)

	// Connect and immediately disconnect repeatedly
	for i := 0; i < 10; i++ {
		conn := dial(t, addr)
		conn.Close()
	}

	// Server should still be responsive after abrupt disconnects
	conn := dial(t, addr)
	defer conn.Close()
	resp := send(t, conn, "*1\r\n$4\r\nPING\r\n")
	assert.Equal(t, "+PONG\r\n", resp)
}
