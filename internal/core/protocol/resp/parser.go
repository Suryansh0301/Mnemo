package resp

import (
	"fmt"
	"strconv"

	"github.com/suryansh0301/mini-redis/internal/enums"
)

type RespValue struct {
	Type   enums.RespType
	Str    string
	Int    int64
	Array  []*RespValue
	Error  error
	IsNull bool
}

type ParseResp struct {
	statusCode    enums.StatusCode
	resp          *RespValue
	bytesConsumed int
	err           error
}

func (r *RespValue) IsType(t enums.RespType) bool {
	return r.Type == t
}

func (r *RespValue) IsEmpty() bool {
	switch r.Type {
	case enums.ArrayRespType:
		return len(r.Array) == 0
	case enums.StringRespType:
		return len(r.Str) == 0
	default:
		// as for other cases currently there is no value considered to be empty
		return false
	}
}

func getParseNeedMoreDataResp() ParseResp {
	return ParseResp{
		statusCode:    enums.NeedMoreDataStatusCode,
		bytesConsumed: 0,
	}
}

func getParseErrorResp(err error) ParseResp {
	return ParseResp{
		statusCode:    enums.ErrorStatusCode,
		bytesConsumed: 0,
		err:           err,
	}
}

func Parse(buffer []byte) ParseResp {
	if len(buffer) == 0 {
		return getParseNeedMoreDataResp()
	}

	typeByte := buffer[0]
	index := readLine(buffer[1:])
	if index == -1 {
		return getParseNeedMoreDataResp()
	}

	return checkBuffer(typeByte, index, buffer[1:])
}

func checkBuffer(typeByte byte, index int, bufferValue []byte) ParseResp {
	switch typeByte {

	case '+':
		for i := 0; i < index; i++ {
			if bufferValue[i] == '\r' || bufferValue[i] == '\n' {
				return getParseErrorResp(fmt.Errorf("invalid simple string"))
			}
		}
		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &RespValue{
				Type: enums.StringRespType,
				Str:  string(bufferValue[:index]),
			},
			bytesConsumed: 1 + index + 2,
		}

	case '-':
		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &RespValue{
				Type:  enums.ErrorRespType,
				Error: fmt.Errorf(string(bufferValue[:index])),
			},
			bytesConsumed: 1 + index + 2,
		}

	case ':':
		if index == 0 {
			return getParseErrorResp(fmt.Errorf("empty integer"))
		}

		start := 0
		sign := ""

		if bufferValue[0] == '+' || bufferValue[0] == '-' {
			if index == 1 {
				return getParseErrorResp(fmt.Errorf("invalid integer"))
			}
			sign = string(bufferValue[0])
			start = 1
		}

		if index-start > 1 && bufferValue[start] == '0' {
			return getParseErrorResp(fmt.Errorf("invalid integer"))
		}

		for i := start; i < index; i++ {
			if bufferValue[i] < '0' || bufferValue[i] > '9' {
				return getParseErrorResp(fmt.Errorf("invalid integer"))
			}
		}

		val, err := strconv.Atoi(string(bufferValue[start:index]))
		if err != nil {
			return getParseErrorResp(fmt.Errorf("invalid integer"))
		}
		if sign == "-" {
			val = val * -1
		}

		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &RespValue{
				Type: enums.IntRespType,
				Int:  int64(val),
			},
			bytesConsumed: 1 + index + 2,
		}

	case '$':
		if index == 0 {
			return getParseErrorResp(fmt.Errorf("empty bulk length"))
		}

		// null bulk string
		if index == 2 && string(bufferValue[:index]) == "-1" {
			return ParseResp{
				statusCode: enums.SuccessStatusCode,
				resp: &RespValue{
					Type:   enums.StringRespType,
					IsNull: true,
				},
				bytesConsumed: 1 + index + 2,
			}
		}

		// reject "-<anything else>"
		if bufferValue[0] == '-' {
			return getParseErrorResp(fmt.Errorf("invalid bulk length"))
		}

		if index > 1 && bufferValue[0] == '0' {
			return getParseErrorResp(fmt.Errorf("invalid bulk length"))
		}

		length := 0
		for i := 0; i < index; i++ {
			if bufferValue[i] < '0' || bufferValue[i] > '9' {
				return getParseErrorResp(fmt.Errorf("invalid bulk length"))
			}
			length = length*10 + int(bufferValue[i]-'0')
		}

		payloadStart := index + 2
		required := payloadStart + length + 2

		if len(bufferValue) < required {
			return getParseNeedMoreDataResp()
		}

		// empty bulk string
		if length == 0 {
			if bufferValue[payloadStart] != '\r' || bufferValue[payloadStart+1] != '\n' {
				return getParseErrorResp(fmt.Errorf("invalid bulk string"))
			}
			return ParseResp{
				statusCode: enums.SuccessStatusCode,
				resp: &RespValue{
					Type: enums.StringRespType,
				},
				bytesConsumed: 1 + index + 4,
			}
		}

		// trailing CRLF check
		if bufferValue[payloadStart+length] != '\r' ||
			bufferValue[payloadStart+length+1] != '\n' {
			return getParseErrorResp(fmt.Errorf("invalid bulk string"))
		}

		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &RespValue{
				Type: enums.StringRespType,
				Str:  string(bufferValue[payloadStart : payloadStart+length]),
			},
			bytesConsumed: 1 + index + 2 + length + 2,
		}

	case '*':
		if index == 0 {
			return getParseErrorResp(fmt.Errorf("empty array length"))
		}
		if bufferValue[0] == '0' {
			if len(bufferValue[:index]) > 1 {
				return getParseErrorResp(fmt.Errorf("invalid array length"))
			}
			return ParseResp{
				statusCode: enums.SuccessStatusCode,
				resp: &RespValue{
					Type:  enums.ArrayRespType,
					Array: []*RespValue{},
				},
				bytesConsumed: 1 + index + 2,
			}
		}

		if bufferValue[0] == '-' {
			if len(bufferValue[:index]) == 1 {
				return getParseErrorResp(fmt.Errorf("invalid array length"))
			}
			if bufferValue[1] == '1' && len(bufferValue[:index]) == 2 {
				return ParseResp{
					statusCode: enums.SuccessStatusCode,
					resp: &RespValue{
						Type:   enums.ArrayRespType,
						IsNull: true,
					},
					bytesConsumed: 1 + index + 2,
				}
			}

			return getParseErrorResp(fmt.Errorf("invalid array length"))
		}

		length := 0
		for i := 0; i < index; i++ {
			if bufferValue[i] < '0' || bufferValue[i] > '9' {
				return getParseErrorResp(fmt.Errorf("invalid array length"))
			}
			length = length*10 + int(bufferValue[i]-'0')
		}

		totalConsumed := 1 + index + 2
		cursor := index + 2
		respValue := make([]*RespValue, 0, length)
		for i := 0; i < length; i++ {
			response := Parse(bufferValue[cursor:])
			if response.statusCode == enums.ErrorStatusCode {
				return getParseErrorResp(response.err)
			}
			if response.statusCode == enums.NeedMoreDataStatusCode {
				return getParseNeedMoreDataResp()
			}
			totalConsumed += response.bytesConsumed
			cursor += response.bytesConsumed
			respValue = append(respValue, response.resp)
		}

		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			resp: &RespValue{
				Type:  enums.ArrayRespType,
				Array: respValue,
			},
			bytesConsumed: totalConsumed,
		}

	default:
		return getParseErrorResp(fmt.Errorf("invalid type byte"))
	}
}

func readLine(buffer []byte) int {
	for i := 0; i+1 < len(buffer); i++ {
		if buffer[i] == '\r' && buffer[i+1] == '\n' {
			return i
		}
	}
	return -1
}
