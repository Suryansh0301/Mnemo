package resp

import (
	"strconv"

	"github.com/suryansh0301/Mnemo/internal/core/common"
	"github.com/suryansh0301/Mnemo/internal/enums"
)

type ParseResp struct {
	statusCode    enums.StatusCode
	Resp          *common.RespValue
	bytesConsumed int
	err           error
}

func (p *ParseResp) BytesConsumed() int {
	return p.bytesConsumed
}

func (p *ParseResp) Error() error {
	return p.err
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

func checkBuffer(typeByte byte, index int, buffer []byte) ParseResp {
	switch typeByte {

	case '+':
		return parseSimpleString(buffer, index)
	case '-':
		return parseError(buffer, index)
	case ':':
		return parseInteger(buffer, index)
	case '$':
		return parseBulkString(buffer, index)
	case '*':
		return parseArray(buffer, index)
	default:
		return getParseErrorResp(common.ProtocolError("invalid RESP type"))
	}
}

func parseSimpleString(buffer []byte, index int) ParseResp {
	for i := range index {
		if buffer[i] == '\r' || buffer[i] == '\n' {
			return getParseErrorResp(common.ProtocolError("invalid simple string"))
		}
	}

	return ParseResp{
		statusCode: enums.SuccessStatusCode,
		Resp: &common.RespValue{
			Type: enums.SimpleStringRespType,
			Str:  string(buffer[:index]),
		},
		bytesConsumed: 1 + index + 2,
	}
}

func parseBulkString(buffer []byte, index int) ParseResp {
	if index == 0 {
		return getParseErrorResp(common.ProtocolError("invalid bulk length"))
	}

	length64, err := strconv.ParseInt(string(buffer[:index]), 10, 64)
	if err != nil {
		return getParseErrorResp(common.ProtocolError("invalid bulk length"))
	}

	// Null bulk string
	if length64 == -1 {
		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			Resp: &common.RespValue{
				Type:   enums.BulkStringRespType,
				IsNull: true,
			},
			bytesConsumed: 1 + index + 2,
		}
	}

	if length64 < 0 {
		return getParseErrorResp(common.ProtocolError("invalid bulk length"))
	}

	if length64 > int64(len(buffer)) {
		return getParseNeedMoreDataResp()
	}

	length := int(length64)
	payloadStart := index + 2
	required := payloadStart + length + 2

	if len(buffer) < required {
		return getParseNeedMoreDataResp()
	}

	if buffer[payloadStart+length] != '\r' ||
		buffer[payloadStart+length+1] != '\n' {
		return getParseErrorResp(common.ProtocolError("invalid bulk string"))
	}

	return ParseResp{
		statusCode: enums.SuccessStatusCode,
		Resp: &common.RespValue{
			Type: enums.BulkStringRespType,
			Str:  string(buffer[payloadStart : payloadStart+length]),
		},
		bytesConsumed: 1 + index + 2 + length + 2,
	}
}

func parseArray(buffer []byte, index int) ParseResp {
	if index == 0 {
		return getParseErrorResp(common.ProtocolError("invalid array length"))
	}

	length64, err := strconv.ParseInt(string(buffer[:index]), 10, 64)
	if err != nil {
		return getParseErrorResp(common.ProtocolError("invalid array length"))
	}

	if length64 == -1 {
		return ParseResp{
			statusCode: enums.SuccessStatusCode,
			Resp: &common.RespValue{
				Type:   enums.ArrayRespType,
				IsNull: true,
			},
			bytesConsumed: 1 + index + 2,
		}
	}

	if length64 < 0 {
		return getParseErrorResp(common.ProtocolError("invalid array length"))
	}

	if length64 > int64(len(buffer)) {
		return getParseNeedMoreDataResp()
	}

	length := int(length64)

	totalConsumed := 1 + index + 2
	cursor := index + 2

	values := make([]*common.RespValue, 0, length)

	for range length {
		response := Parse(buffer[cursor:])

		if response.statusCode == enums.ErrorStatusCode {
			return getParseErrorResp(response.err)
		}

		if response.statusCode == enums.NeedMoreDataStatusCode {
			return getParseNeedMoreDataResp()
		}

		values = append(values, response.Resp)
		cursor += response.bytesConsumed
		totalConsumed += response.bytesConsumed
	}

	return ParseResp{
		statusCode: enums.SuccessStatusCode,
		Resp: &common.RespValue{
			Type:  enums.ArrayRespType,
			Array: values,
		},
		bytesConsumed: totalConsumed,
	}
}

func parseInteger(buffer []byte, index int) ParseResp {
	if index == 0 {
		return getParseErrorResp(common.ProtocolError("invalid integer"))
	}

	val, err := strconv.ParseInt(string(buffer[:index]), 10, 64)
	if err != nil {
		return getParseErrorResp(common.ProtocolError("invalid integer"))
	}

	return ParseResp{
		statusCode: enums.SuccessStatusCode,
		Resp: &common.RespValue{
			Type: enums.IntRespType,
			Int:  val,
		},
		bytesConsumed: 1 + index + 2,
	}
}

func parseError(buffer []byte, index int) ParseResp {
	for i := range index {
		if buffer[i] == '\r' || buffer[i] == '\n' {
			return getParseErrorResp(common.ProtocolError("invalid error"))
		}
	}

	return ParseResp{
		statusCode: enums.SuccessStatusCode,
		Resp: &common.RespValue{
			Type: enums.ErrorRespType,
			Str:  string(buffer[:index]),
		},
		bytesConsumed: 1 + index + 2,
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
