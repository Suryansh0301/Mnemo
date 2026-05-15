package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/suryansh0301/Mnemo/internal/core/common"
	"github.com/suryansh0301/Mnemo/internal/core/datastore"
	parser "github.com/suryansh0301/Mnemo/internal/core/protocol/resp"
	"github.com/suryansh0301/Mnemo/internal/enums"
)

const (
	ReadTimeout  = 30 * time.Second
	WriteTimeout = 30 * time.Second
)

type client struct {
	reader         *bufio.Reader
	writer         *bufio.Writer
	responseChan   chan common.RespValue
	pendingRequest sync.WaitGroup
	parserBuffer   []byte
	readBuffer     []byte
	conn           net.Conn
}

func newClient(connection net.Conn) *client {
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)

	return &client{
		reader:       reader,
		writer:       writer,
		responseChan: make(chan common.RespValue, 256),
		parserBuffer: make([]byte, 0, 4096),
		readBuffer:   make([]byte, 4096),
		conn:         connection,
	}
}

func (c *client) handleConnection(exec *datastore.Executor, totalClients *atomic.Int64) {
	_, cancel := context.WithCancel(context.Background())

	defer c.conn.Close()
	defer func() {
		c.drainRequests()
		close(c.responseChan)
		totalClients.Add(-1)
	}()
	defer cancel()

	go c.handleWrites(cancel)
	c.handleReads(exec)
}

func (c *client) handleWrites(cancel context.CancelFunc) {
	for resp := range c.responseChan {
		byteResp := parser.Encoder(resp)
		c.handleWrite(cancel, byteResp)

	}
}

func (c *client) handleWrite(cancel context.CancelFunc, byteResp []byte) {
	var err error

	defer func() {
		c.decreasePendingRequest()
		if err != nil {
			slog.Info("encountered error while writing", "error", err.Error())
			c.setReadDeadline(-1)
			cancel()
		}
	}()

	_, err = c.writer.Write(byteResp)
	if err != nil {
		return
	}

	if len(c.responseChan) == 0 {
		c.setWriteDeadline(WriteTimeout)
		err = c.writer.Flush()
		if err != nil {
			return
		}
	}

}

func (c *client) handleReads(exec *datastore.Executor) {
	for {
		c.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
		n, err := c.read()
		if err != nil {
			if err != io.EOF && !errors.Is(err, os.ErrDeadlineExceeded) {
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

			if len(c.parserBuffer) == 0 {
				c.parserBuffer = c.parserBuffer[:0]
			}

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

func (c *client) setReadDeadline(duration time.Duration) error {
	err := c.conn.SetReadDeadline(time.Now().Add(duration))
	return err
}

func (c *client) setWriteDeadline(duration time.Duration) error {
	err := c.conn.SetWriteDeadline(time.Now().Add(duration))
	return err
}
