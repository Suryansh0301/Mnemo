package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suryansh0301/Mnemo/internal/core/common"
	"github.com/suryansh0301/Mnemo/internal/enums"
)

// Helpers
func makeStore(pairs ...string) map[string]string {
	store := make(map[string]string)
	for i := 0; i+1 < len(pairs); i += 2 {
		store[pairs[i]] = pairs[i+1]
	}
	return store
}

func TestPing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no args",
			args:        []string{},
			expectError: false,
		},
		{
			name:        "with args",
			args:        []string{"hello"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "PING", Args: tt.args}
			resp := HandlerPing(cmd, nil)
			if tt.expectError {
				assert.Equal(t, enums.ErrorRespType, resp.Type)
				assert.Equal(t, common.WrongNumberOfArgumentsError("PING"), resp.Str)
			} else {
				assert.Equal(t, enums.SimpleStringRespType, resp.Type)
				assert.Equal(t, "PONG", resp.Str)
			}
		})
	}
}

func TestEcho(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		expected    string
	}{
		{
			name:     "valid",
			args:     []string{"hello"},
			expected: "hello",
		},
		{
			name:        "no args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"hello", "world"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "ECHO", Args: tt.args}
			resp := HandlerEcho(cmd, nil)
			if tt.expectError {
				assert.Equal(t, enums.ErrorRespType, resp.Type)
				assert.Equal(t, common.WrongNumberOfArgumentsError("ECHO"), resp.Str)
			} else {
				assert.Equal(t, enums.BulkStringRespType, resp.Type)
				assert.Equal(t, tt.expected, resp.Str)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name: "valid",
			args: []string{"foo", "bar"},
		},
		{
			name:        "no args",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "only key",
			args:        []string{"foo"},
			expectError: true,
		},
		{
			name:        "too many args",
			args:        []string{"foo", "bar", "baz"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := makeStore()
			cmd := Command{Name: "SET", Args: tt.args}
			resp := HandlerSet(cmd, store)
			if tt.expectError {
				assert.Equal(t, enums.ErrorRespType, resp.Type)
				assert.Equal(t, common.WrongNumberOfArgumentsError("SET"), resp.Str)
			} else {
				assert.Equal(t, enums.SimpleStringRespType, resp.Type)
				assert.Equal(t, "OK", resp.Str)
				assert.Equal(t, tt.args[1], store[tt.args[0]])
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]string
		args        []string
		expectNull  bool
		expectError bool
		expected    string
	}{
		{
			name:     "existing key",
			store:    makeStore("foo", "bar"),
			args:     []string{"foo"},
			expected: "bar",
		},
		{
			name:       "missing key",
			store:      makeStore(),
			args:       []string{"foo"},
			expectNull: true,
		},
		{
			name:        "no args",
			store:       makeStore(),
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			store:       makeStore(),
			args:        []string{"foo", "bar"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "GET", Args: tt.args}
			resp := HandlerGet(cmd, tt.store)
			if tt.expectError {
				assert.Equal(t, common.WrongNumberOfArgumentsError("GET"), resp.Str)
			} else if tt.expectNull {
				assert.Equal(t, enums.BulkStringRespType, resp.Type)
				assert.True(t, resp.IsNull)
			} else {
				assert.Equal(t, enums.BulkStringRespType, resp.Type)
				assert.Equal(t, tt.expected, resp.Str)
			}
		})
	}
}

func TestIncr(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]string
		args        []string
		expectError bool
		errorMsg    string
		expected    int64
	}{
		{
			name:     "existing integer key",
			store:    makeStore("counter", "5"),
			args:     []string{"counter"},
			expected: 6,
		},
		{
			name:        "max int64 overflow",
			store:       makeStore("counter", "9223372036854775807"),
			args:        []string{"counter"},
			expectError: true,
			errorMsg:    "ERR value is not an integer or out of range",
		},
		{
			name:     "missing key starts at 0",
			store:    makeStore(),
			args:     []string{"counter"},
			expected: 1,
		},
		{
			name:        "non integer value",
			store:       makeStore("foo", "bar"),
			args:        []string{"foo"},
			expectError: true,
			errorMsg:    "ERR value is not an integer or out of range",
		},
		{
			name:        "no args",
			store:       makeStore(),
			args:        []string{},
			expectError: true,
			errorMsg:    common.WrongNumberOfArgumentsError("INCR"),
		},
		{
			name:        "too many args",
			store:       makeStore(),
			args:        []string{"foo", "bar"},
			expectError: true,
			errorMsg:    common.WrongNumberOfArgumentsError("INCR"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "INCR", Args: tt.args}
			resp := HandlerIncr(cmd, tt.store)
			if tt.expectError {
				assert.Equal(t, enums.ErrorRespType, resp.Type)
				assert.Equal(t, tt.errorMsg, resp.Str)
			} else {
				assert.Equal(t, enums.IntRespType, resp.Type)
				assert.Equal(t, tt.expected, resp.Int)
			}
		})
	}
}

func TestDel(t *testing.T) {
	tests := []struct {
		name        string
		store       map[string]string
		args        []string
		expectError bool
		expected    int64
	}{
		{
			name:     "existing key",
			store:    makeStore("foo", "bar"),
			args:     []string{"foo"},
			expected: 1,
		},
		{
			name:        "multiple keys not supported yet",
			args:        []string{"foo", "bar"},
			expectError: true,
		},
		{
			name:     "missing key",
			store:    makeStore(),
			args:     []string{"foo"},
			expected: 0,
		},
		{
			name:        "no args",
			store:       makeStore(),
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many args",
			store:       makeStore(),
			args:        []string{"foo", "bar"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command{Name: "DEL", Args: tt.args}
			resp := HandlerDel(cmd, tt.store)
			if tt.expectError {
				assert.Equal(t, enums.ErrorRespType, resp.Type)
				assert.Equal(t, common.WrongNumberOfArgumentsError("DEL"), resp.Str)
			} else {
				assert.Equal(t, enums.IntRespType, resp.Type)
				assert.Equal(t, tt.expected, resp.Int)
				if tt.expected == 1 {
					if _, exists := tt.store[tt.args[0]]; exists {
						t.Fatal("key should have been deleted from store")
					}
				}
			}
		})
	}
}
