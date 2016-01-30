package ticketswap

import (
	"errors"
	"fmt"
	"strings"
)

// Supported commands
type CommandType int

// The supported command types
const (
	tCmdHelp CommandType = iota
	tCmdStartWatch
	tCmdStopWatch
	tCmdList
)

// the map of supported commands
var supportedCommands = map[string]CommandType{"/help": tCmdHelp, "/startwatch": tCmdStartWatch, "/stopwatch": tCmdStopWatch, "/list": tCmdList}

// Command defines a command that is supported by the Bot.
type Command struct {
	CommandType CommandType
	Argv        []string
}

// NewCommand creates a Command from its string representation. Returns an error if any.
func NewCommand(cmd string) (*Command, error) {
	msg := strings.Split(cmd, " ")
	cmdType, ok := supportedCommands[msg[0]]
	if !ok {
		return nil, fmt.Errorf("The command is not supported by this bot!")
	}
	var argv []string
	if len(msg) > 1 {
		argv = msg[1:]
	}
	if (cmdType == tCmdStartWatch || cmdType == tCmdStopWatch) && len(argv) != 1 {
		return nil, errors.New("The number of aguments is wrong!")
	}
	return &Command{cmdType, argv}, nil
}
