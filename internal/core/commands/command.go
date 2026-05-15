package commands

import (
	"math"
	"strconv"

	"github.com/suryansh0301/Mnemo/internal/core/common"
	"github.com/suryansh0301/Mnemo/internal/enums"
)

type Command struct {
	Name string
	Args []string
}

var commandsHandler map[enums.CommandName]func(Command, map[string]string) common.RespValue

func init() {
	commandsHandler = make(map[enums.CommandName]func(Command, map[string]string) common.RespValue)
	commandsHandler[enums.PingCommandName] = HandlerPing
	commandsHandler[enums.EchoCommandName] = HandlerEcho
	commandsHandler[enums.SetCommandName] = HandlerSet
	commandsHandler[enums.GetCommandName] = HandlerGet
	commandsHandler[enums.IncrCommandName] = HandlerIncr
	commandsHandler[enums.DeleteCommandName] = HandlerDel
}

func CommandHandler(commandName string) func(Command, map[string]string) common.RespValue {
	handler, exists := commandsHandler[enums.StringToCommandName(commandName)]
	if !exists {
		return nil
	}
	return handler
}

func HandlerPing(command Command, _ map[string]string) common.RespValue {
	if len(command.Args) != 0 {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  common.WrongNumberOfArgumentsError(command.Name),
		}
	}
	return common.RespValue{
		Type: enums.SimpleStringRespType,
		Str:  "PONG",
	}
}

func HandlerEcho(command Command, _ map[string]string) common.RespValue {
	if len(command.Args) != 1 {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  common.WrongNumberOfArgumentsError(command.Name),
		}
	}

	return common.RespValue{
		Type: enums.BulkStringRespType,
		Str:  command.Args[0],
	}

}

func HandlerSet(command Command, store map[string]string) common.RespValue {
	if len(command.Args) != 2 {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  common.WrongNumberOfArgumentsError(command.Name),
		}
	}
	store[command.Args[0]] = command.Args[1]
	return common.RespValue{
		Type: enums.SimpleStringRespType,
		Str:  "OK",
	}
}

func HandlerGet(command Command, store map[string]string) common.RespValue {
	if len(command.Args) != 1 {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  common.WrongNumberOfArgumentsError(command.Name),
		}
	}
	value, exists := store[command.Args[0]]
	if !exists {
		return common.RespValue{
			Type:   enums.BulkStringRespType,
			IsNull: true,
		}
	}
	return common.RespValue{
		Type: enums.BulkStringRespType,
		Str:  value,
	}
}

func HandlerIncr(command Command, store map[string]string) common.RespValue {
	if len(command.Args) != 1 {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  common.WrongNumberOfArgumentsError(command.Name),
		}
	}
	value, exists := store[command.Args[0]]
	if !exists {
		value = "0"
	}

	integer, err := strconv.Atoi(value)
	if err != nil {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  "ERR value is not an integer or out of range",
		}
	}

	if integer == math.MaxInt64 {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  "ERR value is not an integer or out of range",
		}
	}

	integer = integer + 1
	store[command.Args[0]] = strconv.Itoa(integer)
	return common.RespValue{
		Type: enums.IntRespType,
		Int:  int64(integer),
	}
}

func HandlerDel(command Command, store map[string]string) common.RespValue {
	if len(command.Args) != 1 {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  common.WrongNumberOfArgumentsError(command.Name),
		}
	}
	_, exists := store[command.Args[0]]
	if !exists {
		return common.RespValue{
			Type: enums.IntRespType,
			Int:  0,
		}
	}
	delete(store, command.Args[0])
	return common.RespValue{
		Type: enums.IntRespType,
		Int:  1,
	}
}
