package prompt

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

func GetMessage(ctx context.Context, chatHistory []*schema.Message, userQuestion string) (result []*schema.Message, err error) {
	messages := make([]*schema.Message, 0)

	messages = append(messages, schema.SystemMessage("你是一个人工智能助手，名称是Amadeus。你需要用语气平淡，内容简洁且专业的语气回答问题。"))

	messages = append(messages, chatHistory...)

	messages = append(messages, schema.UserMessage(userQuestion))

	return messages, nil
}
