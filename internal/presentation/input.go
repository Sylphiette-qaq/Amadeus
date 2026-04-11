package presentation

import (
	"bufio"
	"fmt"
	"os"
)

func ReadUserInput() (string, error) {
	fmt.Print("请输入你的问题：")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
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
