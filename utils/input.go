package utils

import "fmt"

func ReadUserInput() (string, error) {
	fmt.Print("请输入你的问题：")
	var userQuestion string
	count, err := fmt.Scanln(&userQuestion)
	if err != nil {
		return "", err
	}
	if count <= 0 {
		return "", nil
	}
	return userQuestion, nil
}
