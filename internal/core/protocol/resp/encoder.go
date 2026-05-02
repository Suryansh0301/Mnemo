package resp

import (
	"fmt"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

/*

type RespValue struct {
	Type   enums.RespType
	Str    string
	Int    int64
	Array  []*RespValue
	IsNull bool
}

type RespType int

const (
	SimpleStringRespType RespType = iota
	BulkStringRespType
	IntRespType
	ArrayRespType
	ErrorRespType
)

*/

func Encoder(resp common.RespValue) []byte {
	if resp.Type == enums.SimpleStringRespType {
		response := fmt.Sprintf("+" + resp.Str + "\r\n")
		return []byte(response)
	} else if resp.Type == enums.IntRespType {

	} else if resp.Type == enums.BulkStringRespType {

	} else if resp.Type == enums.ErrorRespType {
	}
	return []byte{}
}
