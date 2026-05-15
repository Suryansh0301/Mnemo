package common

import (
	"fmt"

	"github.com/suryansh0301/Mnemo/internal/enums"
)

type RespValue struct {
	Type   enums.RespType
	Str    string
	Int    int64
	Array  []*RespValue
	IsNull bool
}

func WrongNumberOfArgumentsError(command string) string {
	return fmt.Sprintf("ERR wrong number of arguments for '%s' command", command)
}

func ProtocolError(errMessage string) error {
	return fmt.Errorf("protocol error: %s", errMessage)
}
