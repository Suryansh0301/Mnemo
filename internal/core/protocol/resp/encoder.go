package resp

import (
	"fmt"
	"strconv"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

var encoderHandler map[enums.RespType]func(value common.RespValue) []byte

func init() {
	encoderHandler = make(map[enums.RespType]func(value common.RespValue) []byte)

	encoderHandler[enums.SimpleStringRespType] = func(value common.RespValue) []byte {
		response := fmt.Sprintf("+%s\r\n", value.Str)
		return []byte(response)
	}

	encoderHandler[enums.IntRespType] = func(value common.RespValue) []byte {
		response := fmt.Sprintf(":%s\r\n", strconv.FormatInt(value.Int, 10))
		return []byte(response)
	}

	encoderHandler[enums.BulkStringRespType] = func(value common.RespValue) []byte {
		if value.IsNull {
			return []byte("$-1\r\n")
		}
		response := fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(value.Str)), value.Str)
		return []byte(response)
	}

	encoderHandler[enums.ErrorRespType] = func(value common.RespValue) []byte {
		response := fmt.Sprintf("-%s\r\n", value.Str)
		return []byte(response)
	}
}

func Encoder(resp common.RespValue) []byte {
	handler, exists := encoderHandler[resp.Type]
	if !exists {
		return []byte("-ERR internal error\r\n")
	}
	return handler(resp)
}
