package utils

import (
	"bufio"
	"fmt"
	"os"
)

// ReadUserInput 读取用户输入的整行文本
// 返回:
//
//	用户输入的文本
//	错误信息
func ReadUserInput() (string, error) {
	fmt.Print("请输入你的问题：")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// 移除末尾的换行符
	if len(input) > 0 && input[len(input)-1] == '\n' {
		input = input[:len(input)-1]
	}
	// 移除可能的回车符（Windows系统）
	if len(input) > 0 && input[len(input)-1] == '\r' {
		input = input[:len(input)-1]
	}
	return input, nil
}
