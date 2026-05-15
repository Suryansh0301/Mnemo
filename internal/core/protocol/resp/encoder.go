package resp

import (
	"strconv"
	"sync"

	"github.com/suryansh0301/Mnemo/internal/core/common"
	"github.com/suryansh0301/Mnemo/internal/enums"
)

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 256)
		return &b
	},
}

var encoderHandler map[enums.RespType]func(value common.RespValue) []byte

func init() {
	encoderHandler = make(map[enums.RespType]func(value common.RespValue) []byte)

	encoderHandler[enums.SimpleStringRespType] = func(value common.RespValue) []byte {
		bufPtr := bufPool.Get().(*[]byte)
		buf := (*bufPtr)[:0]

		buf = append(buf, '+')
		buf = append(buf, value.Str...)
		buf = append(buf, '\r', '\n')

		result := make([]byte, len(buf))
		copy(result, buf)

		*bufPtr = buf
		bufPool.Put(bufPtr)

		return result
	}

	encoderHandler[enums.IntRespType] = func(value common.RespValue) []byte {
		bufPtr := bufPool.Get().(*[]byte)
		buf := (*bufPtr)[:0]

		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, value.Int, 10)
		buf = append(buf, '\r', '\n')

		result := make([]byte, len(buf))
		copy(result, buf)

		*bufPtr = buf
		bufPool.Put(bufPtr)

		return result
	}

	encoderHandler[enums.BulkStringRespType] = func(value common.RespValue) []byte {
		if value.IsNull {
			return []byte("$-1\r\n")
		}

		bufPtr := bufPool.Get().(*[]byte)
		buf := (*bufPtr)[:0]

		buf = append(buf, '$')
		buf = strconv.AppendInt(buf, int64(len(value.Str)), 10)
		buf = append(buf, '\r', '\n')
		buf = append(buf, []byte(value.Str)...)
		buf = append(buf, '\r', '\n')

		result := make([]byte, len(buf))
		copy(result, buf)

		*bufPtr = buf
		bufPool.Put(bufPtr)

		return result
	}

	encoderHandler[enums.ErrorRespType] = func(value common.RespValue) []byte {
		bufPtr := bufPool.Get().(*[]byte)
		buf := (*bufPtr)[:0]

		buf = append(buf, '-')
		buf = append(buf, []byte(value.Str)...)
		buf = append(buf, '\r', '\n')

		result := make([]byte, len(buf))
		copy(result, buf)

		*bufPtr = buf
		bufPool.Put(bufPtr)

		return result
	}
}

func Encoder(resp common.RespValue) []byte {
	handler, exists := encoderHandler[resp.Type]
	if !exists {
		return []byte("-ERR internal error\r\n")
	}
	return handler(resp)
}
