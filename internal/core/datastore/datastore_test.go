package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suryansh0301/mini-redis/internal/core/commands"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

func makeCommand(name string, args ...string) commands.Command {
	return commands.Command{Name: name, Args: args}
}

func TestExecuteUnknownCommand(t *testing.T) {
	exec := NewExecutor()
	resp := exec.Execute(makeCommand("INVALID"))
	assert.Equal(t, enums.ErrorRespType, resp.Type)

	expected := "ERR unknown command 'INVALID'"
	assert.Equal(t, expected, resp.Str)
}

func TestExecutePing(t *testing.T) {
	exec := NewExecutor()
	resp := exec.Execute(makeCommand("PING"))
	assert.Equal(t, enums.SimpleStringRespType, resp.Type)
	assert.Equal(t, "PONG", resp.Str)
}

func TestExecuteSetGet(t *testing.T) {
	exec := NewExecutor()

	// SET
	setResp := exec.Execute(makeCommand("SET", "foo", "bar"))
	assert.Equal(t, enums.SimpleStringRespType, setResp.Type)
	assert.Equal(t, "OK", setResp.Str)

	// GET existing
	getResp := exec.Execute(makeCommand("GET", "foo"))
	assert.Equal(t, enums.BulkStringRespType, getResp.Type)
	assert.Equal(t, "bar", getResp.Str)

	// GET missing
	getMissing := exec.Execute(makeCommand("GET", "missing"))
	assert.True(t, getMissing.IsNull)
}

func TestExecuteIncr(t *testing.T) {
	exec := NewExecutor()

	// INCR missing key — should start from 0
	resp := exec.Execute(makeCommand("INCR", "counter"))
	assert.Equal(t, enums.IntRespType, resp.Type)
	assert.Equal(t, int64(1), resp.Int)

	// check if key is present or not
	resp = exec.Execute(makeCommand("GET", "counter"))
	assert.Equal(t, "1", resp.Str)

	// INCR again
	resp = exec.Execute(makeCommand("INCR", "counter"))
	assert.Equal(t, int64(2), resp.Int)

	// INCR non integer
	exec.Execute(makeCommand("SET", "foo", "bar"))
	resp = exec.Execute(makeCommand("INCR", "foo"))
	assert.Equal(t, enums.ErrorRespType, resp.Type)
}

func TestExecuteDel(t *testing.T) {
	exec := NewExecutor()

	exec.Execute(makeCommand("SET", "foo", "bar"))

	// DEL existing
	resp := exec.Execute(makeCommand("DEL", "foo"))
	assert.Equal(t, enums.IntRespType, resp.Type)
	assert.Equal(t, int64(1), resp.Int)

	// Verify deleted
	getResp := exec.Execute(makeCommand("GET", "foo"))
	assert.True(t, getResp.IsNull)

	// DEL missing
	resp = exec.Execute(makeCommand("DEL", "foo"))
	assert.Equal(t, int64(0), resp.Int)
}

func TestExecutorStateIsolation(t *testing.T) {
	// Two executors should have independent datastores
	exec1 := NewExecutor()
	exec2 := NewExecutor()

	exec1.Execute(makeCommand("SET", "foo", "bar"))

	resp := exec2.Execute(makeCommand("GET", "foo"))
	assert.True(t, resp.IsNull)
}

func TestExecuteOverwrite(t *testing.T) {
	exec := NewExecutor()
	exec.Execute(makeCommand("SET", "foo", "bar"))
	exec.Execute(makeCommand("SET", "foo", "baz"))

	resp := exec.Execute(makeCommand("GET", "foo"))
	assert.Equal(t, "baz", resp.Str)
}
