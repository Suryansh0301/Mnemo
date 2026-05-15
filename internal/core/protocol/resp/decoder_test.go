package resp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suryansh0301/Mnemo/internal/core/common"
	"github.com/suryansh0301/Mnemo/internal/enums"
)

// Helper to build a ParseResp from a RespValue directly
func makeParseResp(v *common.RespValue) ParseResp {
	return ParseResp{
		Resp:          v,
		bytesConsumed: 1,
	}
}

func TestDecoderPing(t *testing.T) {
	input := makeParseResp(&common.RespValue{
		Type: enums.ArrayRespType,
		Array: []*common.RespValue{
			{
				Type: enums.BulkStringRespType,
				Str:  "ping",
			},
		},
	})

	cmd, err := Decoder(input)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "PING", cmd.Name)
	assert.Equal(t, 0, len(cmd.Args))
}

func TestDecoderSet(t *testing.T) {
	input := makeParseResp(&common.RespValue{
		Type: enums.ArrayRespType,
		Array: []*common.RespValue{
			{
				Type: enums.BulkStringRespType,
				Str:  "SET",
			},
			{
				Type: enums.BulkStringRespType,
				Str:  "foo",
			},
			{
				Type: enums.BulkStringRespType,
				Str:  "bar",
			},
		},
	})

	cmd, err := Decoder(input)

	assert.NoError(t, err)
	assert.Equal(t, "SET", cmd.Name)
	assert.Equal(t, 2, len(cmd.Args))
	assert.Equal(t, "foo", cmd.Args[0])
	assert.Equal(t, "bar", cmd.Args[1])
}

func TestDecoderPassesThroughExtraArgs(t *testing.T) {
	// Decoder will not validate arg count
	input := makeParseResp(&common.RespValue{
		Type: enums.ArrayRespType,
		Array: []*common.RespValue{
			{
				Type: enums.BulkStringRespType,
				Str:  "SET",
			},
			{
				Type: enums.BulkStringRespType,
				Str:  "foo",
			},
			{
				Type: enums.BulkStringRespType,
				Str:  "bar",
			},
			{
				Type: enums.BulkStringRespType,
				Str:  "extra",
			},
		},
	})

	cmd, err := Decoder(input)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(cmd.Args)) // decoder passes all args through
}

func TestDecoderCaseInsensitive(t *testing.T) {
	// Redis commands are case insensitive
	tests := []string{"ping", "PING", "Ping", "pInG"}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			input := makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{Type: enums.BulkStringRespType, Str: name},
				},
			})
			cmd, err := Decoder(input)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "PING", cmd.Name)
		})
	}
}

func TestDecoderErrors(t *testing.T) {
	tests := []struct {
		name  string
		input ParseResp
	}{
		{
			name:  "nil resp",
			input: ParseResp{Resp: nil},
		},
		{
			name: "not an array",
			input: makeParseResp(&common.RespValue{
				Type: enums.SimpleStringRespType,
				Str:  "PING",
			}),
		},
		{
			name: "null array",
			input: makeParseResp(&common.RespValue{
				Type:   enums.ArrayRespType,
				IsNull: true,
			}),
		},
		{
			name: "empty array",
			input: makeParseResp(&common.RespValue{
				Type:  enums.ArrayRespType,
				Array: []*common.RespValue{},
			}),
		},
		{
			name: "non bulk string command name",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{
						Type: enums.SimpleStringRespType,
						Str:  "PING",
					},
				},
			}),
		},
		{
			name: "null command name",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{
						Type:   enums.BulkStringRespType,
						IsNull: true,
					},
				},
			}),
		},
		{
			name: "null argument",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{
						Type: enums.BulkStringRespType,
						Str:  "GET",
					},
					{
						Type:   enums.BulkStringRespType,
						IsNull: true,
					},
				},
			}),
		},
		{
			name: "non bulk string argument",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{
						Type: enums.SimpleStringRespType,
						Str:  "foo",
					},
					{
						Type: enums.BulkStringRespType,
						Str:  "GET",
					},
				},
			}),
		},
		{
			name: "empty command name",
			input: makeParseResp(&common.RespValue{
				Type: enums.ArrayRespType,
				Array: []*common.RespValue{
					{
						Type: enums.BulkStringRespType,
						Str:  "",
					},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decoder(tt.input)
			assert.Error(t, err)
		})
	}
}
