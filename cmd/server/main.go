package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/core/datastore"
	parser "github.com/suryansh0301/mini-redis/internal/core/protocol/resp"
)

type client struct {
	Reader       *bufio.Reader
	Writer       *bufio.Writer
	ParserBuffer []byte
	ReadBuffer   []byte
	WriteBuffer  []byte
}

func newClient(connection net.Conn) *client {
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)

	return &client{
		Reader:       reader,
		Writer:       writer,
		ParserBuffer: make([]byte, 0, 4096),
		ReadBuffer:   make([]byte, 1024, 4096),
		WriteBuffer:  make([]byte, 1024, 4096),
	}
}

func (c *client) read() (int, error) {
	n, err := c.Reader.Read(c.ReadBuffer)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c *client) appendParseBuffer(n int) {
	c.ParserBuffer = append(c.ParserBuffer, c.ReadBuffer[:n]...)
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	slog.SetLogLoggerLevel(-4)
	slog.Debug("Listening on port 6379")

	exec := datastore.NewExecutor()
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		client := newClient(conn)
		go client.handleConnection(conn, exec)
	}
}

func (c *client) handleConnection(conn net.Conn, exec *datastore.Executor) {

	defer conn.Close()

	for {
		n, err := c.read()
		if err != nil {
			return
		}

		c.appendParseBuffer(n)

		for len(c.ParserBuffer) > 0 {
			response := parser.Parse(c.ParserBuffer)
			if response.Error() != nil {
				// we receive an error response
				common.ProtocolError(response.Error().Error())
				return
			}

			if response.BytesConsumed() == 0 {
				// we need more data hence we break and wait for the next read
				break
			}

			c.ParserBuffer = c.ParserBuffer[response.BytesConsumed():]
			value, err := parser.Decoder(response)
			if err != nil {
				common.ProtocolError(err.Error())
				return
			}

			resp := exec.Execute(value)
			fmt.Println(resp)
			byteResp := parser.Encoder(resp)
			c.WriteBuffer = append(c.WriteBuffer, byteResp...)
			n, err = c.Writer.Write(c.WriteBuffer)
			if err != nil {
				common.ProtocolError(err.Error())
				return
			}
			if n > 0 {
				err = c.Writer.Flush()
				if err != nil {
					common.ProtocolError(err.Error())
					return
				}
			}

		}
	}
}
