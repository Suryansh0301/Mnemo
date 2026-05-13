package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

func TestParseSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantMore bool
		consumed int
		str      string
	}{
		{
			name:     "valid OK",
			input:    "+OK\r\n",
			consumed: 5,
			str:      "OK",
		},
		{
			name:     "valid PONG",
			input:    "+PONG\r\n",
			consumed: 7,
			str:      "PONG",
		},
		{
			name:    "invalid newline",
			input:   "+OK\r\r\n",
			wantErr: true,
		},
		{
			name:     "incomplete",
			input:    "+OK",
			wantMore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, result.Error())
				return
			}
			if tt.wantMore {
				assert.NoError(t, result.Error())
				assert.Equal(t, 0, result.BytesConsumed())
				return
			}
			assert.NoError(t, result.Error())
			assert.Equal(t, tt.consumed, result.BytesConsumed())
			assert.Equal(t, tt.str, result.Resp.Str)
		})
	}
}

func TestParseBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantMore bool
		isNull   bool
		consumed int
		str      string
	}{
		{
			name:     "valid FOO",
			input:    "$3\r\nfoo\r\n",
			consumed: 9,
			str:      "foo",
		},
		{
			name:     "valid null bulk string",
			input:    "$-1\r\n",
			isNull:   true,
			consumed: 5,
		},
		{
			name:     "incomplete bulk string",
			input:    "$3\r\nfo",
			wantMore: true,
		},
		{
			name:    "non numeric length",
			input:   "$abc\r\nfoo\r\n",
			wantErr: true,
		},
		{
			name:    "negative length",
			input:   "$-5\r\nfoo\r\n",
			wantErr: true,
		},
		{
			name:    "missing terminator wrong bytes",
			input:   "$3\r\nfooXX",
			wantErr: true,
		},
		{
			name:     "missing terminator incomplete",
			input:    "$3\r\nfoo",
			wantMore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, result.Error())
				return
			}
			if tt.wantMore {
				assert.NoError(t, result.Error())
				assert.Equal(t, 0, result.BytesConsumed())
				return
			}
			if tt.isNull {
				assert.True(t, result.Resp.IsNull)
				return
			}
			assert.NoError(t, result.Error())
			assert.Equal(t, tt.consumed, result.BytesConsumed())
			assert.Equal(t, tt.str, result.Resp.Str)
		})
	}
}

func TestParseInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantMore bool
		consumed int
		integer  int64
	}{
		{
			name:     "valid OK",
			input:    ":42\r\n",
			consumed: 5,
			integer:  42,
		},
		{
			name:     "valid negative integer",
			input:    ":-1\r\n",
			consumed: 5,
			integer:  -1},
		{
			name:    "invalid integer",
			input:   ":abc\r\n",
			wantErr: true,
		},
		{
			name:     "incomplete",
			input:    ":42",
			wantMore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, result.Error())
				return
			}
			if tt.wantMore {
				assert.NoError(t, result.Error())
				assert.Equal(t, 0, result.BytesConsumed())
				return
			}
			assert.NoError(t, result.Error())
			assert.Equal(t, tt.consumed, result.BytesConsumed())
			assert.Equal(t, tt.integer, result.Resp.Int)
		})
	}
}

func TestParseArray(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantMore  bool
		isNull    bool
		consumed  int
		arrayLen  int
		firstElem string
	}{
		{
			name:      "valid two elements",
			input:     "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n",
			consumed:  24,
			arrayLen:  2,
			firstElem: "PING",
		},
		{
			name:     "null array",
			input:    "*-1\r\n",
			consumed: 5,
			isNull:   true,
		},
		{
			name:     "incomplete",
			input:    "*2\r\n$4\r\nPING\r\n",
			wantMore: true,
		},
		{
			name:    "invalid length",
			input:   "*abc\r\n",
			wantErr: true,
		},
		{
			name:    "negative length",
			input:   "*-5\r\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, result.Error())
				return
			}
			if tt.wantMore {
				assert.NoError(t, result.Error())
				assert.Equal(t, 0, result.BytesConsumed())
				return
			}
			assert.NoError(t, result.Error())
			assert.Equal(t, tt.consumed, result.BytesConsumed())
			if tt.isNull {
				assert.True(t, result.Resp.IsNull)
				return
			}
			assert.Equal(t, tt.arrayLen, len(result.Resp.Array))
			if tt.firstElem != "" {
				assert.Equal(t, tt.firstElem, result.Resp.Array[0].Str)
			}
		})
	}
}

func TestParsePipelined(t *testing.T) {
	input := "*1\r\n$4\r\nPING\r\n*1\r\n$4\r\nPING\r\n"
	buf := []byte(input)

	first := Parse(buf)
	assert.NoError(t, first.Error())
	assert.NotEqual(t, 0, first.BytesConsumed())

	buf = buf[first.BytesConsumed():]

	second := Parse(buf)
	assert.NoError(t, second.Error())
	assert.NotEqual(t, 0, second.BytesConsumed())
}

func TestParseEmptyBuffer(t *testing.T) {
	result := Parse([]byte(""))

	assert.NoError(t, result.Error())
	assert.Equal(t, 0, result.BytesConsumed())
}

func TestParseInvalidType(t *testing.T) {
	result := Parse([]byte("X invalid\r\n"))

	assert.Error(t, result.Error())
}

func FuzzParse(f *testing.F) {
	// Seed with valid inputs
	f.Add([]byte("*1\r\n$4\r\nPING\r\n"))
	f.Add([]byte("+OK\r\n"))
	f.Add([]byte("$3\r\nfoo\r\n"))
	f.Add([]byte(":42\r\n"))
	f.Add([]byte("*-1\r\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must never panic — that's the only invariant
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parser panicked on input %q: %v", data, r)
			}
		}()
		Parse(data)
	})
}

// Round trip — parse what encoder produces
func TestEncoderRoundTrip(t *testing.T) {
	values := []common.RespValue{
		{
			Type: enums.SimpleStringRespType,
			Str:  "OK",
		},
		{
			Type: enums.IntRespType,
			Int:  42,
		},
		{
			Type: enums.BulkStringRespType,
			Str:  "hello",
		},
		{
			Type: enums.ErrorRespType,
			Str:  "ERR something",
		},
		{
			Type:   enums.BulkStringRespType,
			IsNull: true,
		},
	}

	for _, v := range values {
		encoded := Encoder(v)
		parsed := Parse(encoded)

		assert.NoError(t, parsed.Error())
		assert.Equal(t, len(encoded), parsed.BytesConsumed())
	}
}
