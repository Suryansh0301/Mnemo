package command

type Command struct {
	Name string
	Args []string
}

var commandHandler map[string]func(Command) RespValue
