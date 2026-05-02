package commands

import (
	"fmt"
	"strconv"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
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
		err := fmt.Errorf("arguments present in ping command")
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: err,
		}
	}
	return common.RespValue{
		Type: enums.StringRespType,
		Str:  "PONG",
	}
}

func HandlerEcho(command Command, _ map[string]string) common.RespValue {
	if len(command.Args) != 1 {
		err := fmt.Errorf("echo command requires exactly one argument")
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: err,
		}
	}

	return common.RespValue{
		Type: enums.StringRespType,
		Str:  command.Args[0],
	}

}

func HandlerSet(command Command, store map[string]string) common.RespValue {
	if len(command.Args) != 2 {
		err := fmt.Errorf("set command requires exactly two argument")
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: err,
		}
	}
	store[command.Args[0]] = command.Args[1]
	return common.RespValue{
		Type: enums.StringRespType,
		Str:  "OK",
	}
}

func HandlerGet(command Command, store map[string]string) common.RespValue {
	if len(command.Args) != 1 {
		err := fmt.Errorf("get command requires exactly one argument")
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: err,
		}
	}
	value, exists := store[command.Args[0]]
	if !exists {
		return common.RespValue{
			Type:   enums.StringRespType,
			IsNull: true,
		}
	}
	return common.RespValue{
		Type: enums.StringRespType,
		Str:  value,
	}
}

func HandlerIncr(command Command, store map[string]string) common.RespValue {
	if len(command.Args) != 1 {
		err := fmt.Errorf("incr command requires exactly one argument")
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: err,
		}
	}
	value, exists := store[command.Args[0]]
	if !exists {
		value = "0"
	}
	integer, err := strconv.Atoi(value)
	if err != nil {
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: err,
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
		err := fmt.Errorf("del command requires exactly one argument")
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: err,
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
