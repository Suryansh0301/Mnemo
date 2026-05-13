package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

func TestEncodeSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "OK",
			input:    common.RespValue{Type: enums.SimpleStringRespType, Str: "OK"},
			expected: "+OK\r\n",
		},
		{
			name:     "PONG",
			input:    common.RespValue{Type: enums.SimpleStringRespType, Str: "PONG"},
			expected: "+PONG\r\n",
		},
		{
			name:     "empty string",
			input:    common.RespValue{Type: enums.SimpleStringRespType, Str: ""},
			expected: "+\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeError(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "protocol error",
			input:    common.RespValue{Type: enums.ErrorRespType, Str: "ERR Protocol error"},
			expected: "-ERR Protocol error\r\n",
		},
		{
			name:     "unknown command",
			input:    common.RespValue{Type: enums.ErrorRespType, Str: "ERR unknown command 'FOO'"},
			expected: "-ERR unknown command 'FOO'\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "positive",
			input:    common.RespValue{Type: enums.IntRespType, Int: 42},
			expected: ":42\r\n",
		},
		{
			name:     "zero",
			input:    common.RespValue{Type: enums.IntRespType, Int: 0},
			expected: ":0\r\n",
		},
		{
			name:     "negative",
			input:    common.RespValue{Type: enums.IntRespType, Int: -1},
			expected: ":-1\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    common.RespValue
		expected string
	}{
		{
			name:     "simple value",
			input:    common.RespValue{Type: enums.BulkStringRespType, Str: "foo"},
			expected: "$3\r\nfoo\r\n",
		},
		{
			name:     "empty string",
			input:    common.RespValue{Type: enums.BulkStringRespType, Str: ""},
			expected: "$0\r\n\r\n",
		},
		{
			name:     "longer value",
			input:    common.RespValue{Type: enums.BulkStringRespType, Str: "hello world"},
			expected: "$11\r\nhello world\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(Encoder(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeBulkStringNull(t *testing.T) {
	result := string(Encoder(common.RespValue{
		Type:   enums.BulkStringRespType,
		IsNull: true,
	}))
	assert.Equal(t, "$-1\r\n", result)
}

// Empty string is NOT null
func TestEncodeBulkStringEmpty(t *testing.T) {
	result := string(Encoder(common.RespValue{
		Type: enums.BulkStringRespType,
		Str:  "",
	}))
	assert.Equal(t, "$0\r\n\r\n", result)
}

func TestEncodeUnknownType(t *testing.T) {
	result := string(Encoder(common.RespValue{Type: 99}))
	assert.Equal(t, "-ERR internal error\r\n", result)
}
