package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

const contextFilePath = "./checkpoints/context.txt"

func SaveMessage(role schema.RoleType, content string) {
	if err := os.MkdirAll("./checkpoints", 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}

	f, err := os.OpenFile(contextFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("保存上下文失败: %v\n", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s: %s\n", timestamp, role, content)
	if _, err := f.WriteString(line); err != nil {
		fmt.Printf("写入上下文失败: %v\n", err)
	}
}

func LoadContext() []*schema.Message {
	file, err := os.Open(contextFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		fmt.Printf("读取上下文失败: %v\n", err)
		return nil
	}
	defer file.Close()

	var messages []*schema.Message
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		msg := parseContextLine(line)
		if msg != nil {
			messages = append(messages, msg)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("扫描上下文失败: %v\n", err)
	}

	return messages
}

func parseContextLine(line string) *schema.Message {
	idx1 := strings.Index(line, "]")
	if idx1 == -1 {
		return nil
	}

	rest := strings.TrimSpace(line[idx1+1:])
	idx2 := strings.Index(rest, ":")
	if idx2 == -1 {
		return nil
	}

	roleStr := strings.TrimSpace(rest[:idx2])
	content := strings.TrimSpace(rest[idx2+1:])

	var role schema.RoleType
	switch roleStr {
	case "user":
		role = schema.User
	case "assistant":
		role = schema.Assistant
	case "system":
		role = schema.System
	default:
		return nil
	}

	return &schema.Message{
		Role:    role,
		Content: content,
	}
}

func ClearContext() error {
	return os.Remove(contextFilePath)
}
