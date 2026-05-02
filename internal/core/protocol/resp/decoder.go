package resp

import (
	"fmt"
	"strings"

	"github.com/suryansh0301/mini-redis/internal/enums"
)

type Command struct {
	Name string
	Args []string
}

func Decoder(parsedResp ParseResp) (Command, error) {
	if !parsedResp.resp.IsType(enums.RespTypeArray) {
		err := fmt.Errorf("expected array response, got %+v", parsedResp.resp.Type)
		return Command{}, err
	}

	if parsedResp.resp.IsNull || parsedResp.resp.IsEmpty() {
		err := fmt.Errorf("invalid array response")
		return Command{}, err

	}

	commandNameRespValue := parsedResp.resp.Array[0]
	if !commandNameRespValue.IsType(enums.RespTypeString) {
		err := fmt.Errorf("expected command name, got %+v", commandNameRespValue.Type)
		return Command{}, err
	}

	if commandNameRespValue.IsNull || commandNameRespValue.IsEmpty() {
		err := fmt.Errorf("invalid command name")
		return Command{}, err
	}
	commandName := strings.ToUpper(commandNameRespValue.Str)

	commandArgsParsedResp := parsedResp.resp.Array[1:]

	args := make([]string, 0, len(commandArgsParsedResp))
	for i := 0; i < len(commandArgsParsedResp); i++ {
		if !commandArgsParsedResp[i].IsType(enums.RespTypeString) {
			err := fmt.Errorf("expected command args, got %+v", commandArgsParsedResp[i].Type)
			return Command{}, err
		}

		if commandArgsParsedResp[i].IsNull {
			err := fmt.Errorf("invalid command args")
			return Command{}, err
		}
		args = append(args, commandArgsParsedResp[i].Str)
	}

	return Command{
		Name: commandName,
		Args: args,
	}, nil
}
