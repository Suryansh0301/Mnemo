package datastore

import (
	"fmt"

	"github.com/suryansh0301/mini-redis/internal/core/commands"
	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

type Executor struct {
	dataStore map[string]string
}

func NewExecutor() *Executor {
	dataStore := make(map[string]string)
	return &Executor{dataStore}
}

func (e *Executor) Execute(command commands.Command) common.RespValue {
	handler := commands.CommandHandler(command.Name)
	if handler == nil {
		return common.RespValue{
			Type:  enums.ErrorRespType,
			Error: fmt.Errorf("unknown command %s", command.Name),
		}
	}
	return handler(command, e.dataStore)
}
