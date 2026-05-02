package enums

type StatusCode string

const (
	SuccessStatusCode      StatusCode = "success"
	ErrorStatusCode        StatusCode = "error"
	NeedMoreDataStatusCode StatusCode = "needMoreData"
)

type RespType int

const (
	StringRespType RespType = iota
	IntRespType
	ArrayRespType
	ErrorRespType
)

type CommandName string

const (
	SetCommandName    CommandName = "set"
	GetCommandName    CommandName = "get"
	IncrCommandName   CommandName = "incr"
	PingCommandName   CommandName = "ping"
	DeleteCommandName CommandName = "del"
	EchoCommandName   CommandName = "echo"
)

var stringToCommandName = map[string]CommandName{
	"set":  SetCommandName,
	"get":  GetCommandName,
	"incr": IncrCommandName,
	"ping": PingCommandName,
	"del":  DeleteCommandName,
	"echo": EchoCommandName,
}

func StringToCommandName(commandName string) CommandName {
	return stringToCommandName[commandName]
}
