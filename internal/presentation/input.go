package presentation

import (
	"bufio"
	"fmt"
	"os"
)

var stdinReader = bufio.NewReader(os.Stdin)

func ReadUserInput() (string, error) {
	fmt.Print("\n> ")
	input, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(input) > 0 && input[len(input)-1] == '\n' {
		input = input[:len(input)-1]
	}
	if len(input) > 0 && input[len(input)-1] == '\r' {
		input = input[:len(input)-1]
	}

	return input, nil
}
