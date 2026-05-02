package datastore

import (
	"github.com/suryansh0301/mini-redis/internal/core/commands"
)

type Executor struct {
	dataStore map[string]string
}

func NewExecutor() *Executor {
	dataStore := make(map[string]string)
	return &Executor{dataStore}
}

func (e *Executor) Execute(command commands.Command) {
	handler := commands.CommandHandler(command.Name)
	handler(command, e.dataStore)
	return
}
