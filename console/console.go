package console

import (
	"bufio"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/slog"
	"os"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func Run() {
	go doRun()
}

func doRun() {
	for {
		readLine, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		readLine = strings.TrimSuffix(readLine[:len(readLine)-1], "\r")
		contents := strings.Fields(readLine)

		if len(contents) < 2 {
			continue
		}

		moduleName := contents[0]
		command := contents[1]
		argsStrings := contents[2:]
		args := make([]interface{}, len(argsStrings))
		for k, v := range argsStrings {
			args[k] = v
		}
		m := module.FindModule(moduleName)
		if m != nil {
			m.RpcCall(command, args...)
		} else {
			slog.LogWarning("console", "module [%v] not found!", moduleName)
		}
	}
}
