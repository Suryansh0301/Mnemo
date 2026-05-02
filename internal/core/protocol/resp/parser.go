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

func Parse(buffer []byte) ParseResp {
	if len(buffer) == 0 {
		return getParseNeedMoreDataResp()
	}

	bufferValue, bytesConsumed := readLine(buffer)
	if bytesConsumed == 0 {
		return getParseNeedMoreDataResp()
	}

	return checkBuffer(bufferValue)

}

func checkBuffer(bufferValue []byte) ParseResp {
	if len(bufferValue) == 0 {
		return ParseResp{
			result:        constants.ResultTypeError,
			value:         "",
			bytesConsumed: 0,
			err:           fmt.Errorf(""),
		}
	}
	sign := bufferValue[0]
	switch sign {
	case '+':
		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue[1:]),
			bytesConsumed: len(bufferValue) + 2,
			err:           nil,
		}
	case '-':
		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue[1:]),
			bytesConsumed: len(bufferValue) + 2,
			err:           nil,
		}
	case ':':
		for i := range bufferValue {
			if (bufferValue[i] < '0' || bufferValue[i] > '9') && bufferValue[i] != '-' && bufferValue[i] != '+' {
				return ParseResp{
					result:        constants.ResultTypeError,
					value:         "",
					bytesConsumed: 0,
					err:           fmt.Errorf("invalid value received in simple integer : %q", bufferValue),
				}
			}
		}
		return ParseResp{
			result:        constants.ResultTypeSuccess,
			value:         string(bufferValue[1:]),
			bytesConsumed: len(bufferValue) + 2,
			err:           nil,
		}
	default:
		return ParseResp{
			result:        constants.ResultTypeError,
			value:         "",
			bytesConsumed: 0,
			err:           fmt.Errorf("invalid type received: %q", bufferValue[0]),
		}
	}
}

func readLine(buffer []byte) ([]byte, int) {
	for i := range buffer {
		if buffer[i] == '\r' {
			if i+1 < len(buffer) && buffer[i+1] == '\n' {
				return buffer[:i], i + 2
			}
		}
	}

	return nil, 0
}
