package console

import (
	"bufio"
	"os"
	//"runtime/debug"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func Run() {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSuffix(line[:len(line)-1], "\r")
		//debug.FreeOSMemory()
	}
}
