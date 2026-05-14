package main

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/core/datastore"
	parser "github.com/suryansh0301/mini-redis/internal/core/protocol/resp"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

type client struct {
	reader         *bufio.Reader
	writer         *bufio.Writer
	responseChan   chan common.RespValue
	pendingRequest sync.WaitGroup
	parserBuffer   []byte
	readBuffer     []byte
}

func newClient(connection net.Conn) *client {
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)

	return &client{
		reader:       reader,
		writer:       writer,
		responseChan: make(chan common.RespValue, 256),
		parserBuffer: make([]byte, 0, 4096),
		readBuffer:   make([]byte, 1024, 4096),
	}
}

func (c *client) handleConnection(conn net.Conn, exec *datastore.Executor) {
	defer conn.Close()
	defer func() {
		c.drainRequests()
		close(c.responseChan)
	}()

	go c.handleWrites()
	c.handleReads(exec)
}

func (c *client) handleWrites() {
	for resp := range c.responseChan {
		byteResp := parser.Encoder(resp)
		_, err := c.writer.Write(byteResp)
		if err != nil {
			slog.Info("encountered error while writing", "error", err.Error())
		}
		if len(c.responseChan) == 0 {
			err = c.writer.Flush()
			if err != nil {
				slog.Info("encountered error while writing", "error", err.Error())
			}
		}
		c.decreasePendingRequest()
	}
}

func (c *client) handleReads(exec *datastore.Executor) {
	for {
		n, err := c.read()
		if err != nil {
			if err != io.EOF {
				c.handleError()
			}
			return
		}

		c.appendParseBuffer(n)

		for len(c.parserBuffer) > 0 {
			response := parser.Parse(c.parserBuffer)
			if response.Error() != nil {
				// we receive an error response
				c.handleError()
				return
			}

			if response.BytesConsumed() == 0 {
				// we need more data hence we break and wait for the next read
				break
			}

			c.parserBuffer = c.parserBuffer[response.BytesConsumed():]
			value, err := parser.Decoder(response)
			if err != nil {
				c.handleError()
				return
			}

			c.increasePendingRequest()
			exec.ExecutorChan <- datastore.Value{
				Command:      value,
				ResponseChan: c.responseChan,
			}

		}
	}
}

func (c *client) read() (int, error) {
	n, err := c.reader.Read(c.readBuffer)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c *client) appendParseBuffer(n int) {
	c.parserBuffer = append(c.parserBuffer, c.readBuffer[:n]...)
}

func (c *client) handleError() {
	responseErr := common.RespValue{
		Type: enums.ErrorRespType,
		Str:  "ERR Protocol error",
	}
	c.increasePendingRequest()
	c.responseChan <- responseErr
}

func (c *client) increasePendingRequest() {
	c.pendingRequest.Add(1)
}

func (c *client) decreasePendingRequest() {
	c.pendingRequest.Done()
}

func (c *client) drainRequests() {
	c.pendingRequest.Wait()
}
