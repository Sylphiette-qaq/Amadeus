package memory

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino/schema"
)

const contextFilePath = "./checkpoints/context.txt"

func SaveMessage(role schema.RoleType, content string) {
	// M1 仍沿用文本落盘，先保证可恢复会话；结构化存储在 M2 再替换。
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

		// 读取历史时对坏行做容错，避免单条损坏导致整个会话无法恢复。
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

func ClearContext() error {
	return os.Remove(contextFilePath)
}
