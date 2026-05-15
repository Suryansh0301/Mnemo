package datastore

import (
	"fmt"

	"github.com/suryansh0301/Mnemo/internal/core/commands"
	"github.com/suryansh0301/Mnemo/internal/core/common"
	"github.com/suryansh0301/Mnemo/internal/enums"
)

type Executor struct {
	dataStore    map[string]string
	ExecutorChan chan Value
}

type Value struct {
	ResponseChan chan common.RespValue
	Command      commands.Command
}

func NewExecutor() *Executor {
	return &Executor{
		dataStore:    make(map[string]string),
		ExecutorChan: make(chan Value, 1024),
	}
}

func (e *Executor) Execute(command commands.Command) common.RespValue {
	handler := commands.CommandHandler(command.Name)
	if handler == nil {
		return common.RespValue{
			Type: enums.ErrorRespType,
			Str:  fmt.Sprintf("ERR unknown command '%s'", command.Name),
		}
	}
	return handler(command, e.dataStore)
}
