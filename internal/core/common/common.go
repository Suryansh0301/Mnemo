package common

import "github.com/suryansh0301/mini-redis/internal/enums"

type RespValue struct {
	Type   enums.RespType
	Str    string
	Int    int64
	Array  []*RespValue
	Error  error
	Bool   bool
	IsNull bool
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
