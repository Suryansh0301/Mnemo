package resp

import (
	"fmt"
	"strings"

	"github.com/suryansh0301/mini-redis/internal/core/commands"
	"github.com/suryansh0301/mini-redis/internal/enums"
)

func Decoder(parsedResp ParseResp) (command commands.Command, err error) {
	if !parsedResp.resp.IsType(enums.ArrayRespType) {
		err = fmt.Errorf("expected array response, got %+v", parsedResp.resp.Type)
		return
	}

	if parsedResp.resp.IsNull || parsedResp.resp.IsEmpty() {
		err = fmt.Errorf("invalid array response")
		return

	}

	commandNameRespValue := parsedResp.resp.Array[0]
	if !commandNameRespValue.IsType(enums.StringRespType) {
		err = fmt.Errorf("expected command name, got %+v", commandNameRespValue.Type)
		return
	}

	if commandNameRespValue.IsNull || commandNameRespValue.IsEmpty() {
		err = fmt.Errorf("invalid command name")
		return
	}
	commandName := strings.ToUpper(commandNameRespValue.Str)

	commandArgsParsedResp := parsedResp.resp.Array[1:]

	args := make([]string, 0, len(commandArgsParsedResp))
	for i := 0; i < len(commandArgsParsedResp); i++ {
		if !commandArgsParsedResp[i].IsType(enums.StringRespType) {
			err = fmt.Errorf("expected command args, got %+v", commandArgsParsedResp[i].Type)
			return
		}

		if commandArgsParsedResp[i].IsNull {
			err = fmt.Errorf("invalid command args")
			return
		}
		args = append(args, commandArgsParsedResp[i].Str)
	}

	return commands.Command{
		Name: commandName,
		Args: args,
	}, nil
}
