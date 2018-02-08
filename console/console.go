package console

import (
	"bufio"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/slog"
	"os"
	//"runtime"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func Run() {
	for {
		read_line, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		read_line = strings.TrimSuffix(read_line[:len(read_line)-1], "\r")
		contents := strings.Fields(read_line)

		if len(contents) < 2 {
			continue
		}

		module_name := contents[0]
		command := contents[1]
		args := contents[2:]
		m := module.FindModule(module_name)
		if m != nil {
			m.RpcCall(command, args)
		} else {
			slog.LogWarning("console", "module [%v] not found!", module_name)
		}
	}
}
