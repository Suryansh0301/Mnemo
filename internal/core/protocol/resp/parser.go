package resp

import (
	"fmt"

	"github.com/suryansh0301/mini-redis/internal/constants"
)

type ParseResp struct {
	result        constants.ResultType
	value         string
	bytesConsumed int
	err           error
}

func getParseNeedMoreDataResp() ParseResp {
	return ParseResp{
		result:        constants.ResultTypeNeedMoreData,
		value:         "",
		bytesConsumed: 0,
		err:           nil,
	}
}

func getParseErrorResp(err error) ParseResp {
	return ParseResp{
		result:        constants.ResultTypeError,
		value:         "",
		bytesConsumed: 0,
		err:           err,
	}
}

/*
TEST CASES:-
\r\n
+OK\r\n
?\r\n
+OK\r\n+PONG\r\n
$3\r\nfoo\r\n
:-\r\n
:00001\r\n
*/

func Parse(buffer []byte) ParseResp {
	if len(buffer) == 0 {
		return getParseNeedMoreDataResp()
	}

	typeByte := buffer[0]
	bufferValue := readLine(buffer[1:])
	if bufferValue == nil {
		return getParseNeedMoreDataResp()
	}

	return checkBuffer(typeByte, bufferValue)

}

func checkBuffer(typeByte byte, bufferValue []byte) ParseResp {
	switch typeByte {
	case '+':
		for i := range bufferValue {
			if bufferValue[i] == '\r' || bufferValue[i] == '\n' {
				return getParseErrorResp(fmt.Errorf("invalid value received in simple string : %q", bufferValue))
			}
		}

		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue),
			bytesConsumed: 1 + len(bufferValue) + 2,
			err:           nil,
		}
	case '-':
		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue),
			bytesConsumed: 1 + len(bufferValue) + 2,
			err:           nil,
		}
	case ':':
		if len(bufferValue) == 0 {
			return getParseErrorResp(fmt.Errorf("empty value received in simple integer"))
		}

		if len(bufferValue) > 1 && bufferValue[0] == '0' {
			return getParseErrorResp(fmt.Errorf("invalid value received in simple integer : %q", bufferValue))
		}

		for i := range bufferValue {
			if i == 0 && (bufferValue[i] == '-' || bufferValue[i] == '+') {
				if len(bufferValue) == 1 {
					return getParseErrorResp(fmt.Errorf("invalid value received in simple string : %q", bufferValue))
				}
				continue
			}

			if bufferValue[i] < '0' || bufferValue[i] > '9' {
				return getParseErrorResp(fmt.Errorf("invalid value received in simple integer : %q", bufferValue)))
			}
		}
		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue),
			bytesConsumed: 1 + len(bufferValue) + 2,
			err:           nil,
		}
	default:
		return getParseErrorResp(fmt.Errorf("invalid type received: %q", typeByte))
	}
}

func readLine(buffer []byte) []byte {
	for i := range buffer {
		if buffer[i] == '\r' {
			if i+1 < len(buffer) && buffer[i+1] == '\n' {
				return buffer[:i]
			}
		}
	}

	return nil
}
