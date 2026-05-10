package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// OneBot 11 的 message 字段可能是字符串或消息段数组，用自定义类型兼容两种格式。
type messageField string

func (m *messageField) UnmarshalJSON(b []byte) error {
	// 尝试字符串
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*m = messageField(s)
		return nil
	}
	// 尝试数组：提取其中 type=text 的段拼接为纯文本
	var segments []struct {
		Type string `json:"type"`
		Data struct {
			Text string `json:"text"`
			QQ   string `json:"qq"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &segments); err != nil {
		return err
	}
	var parts []string
	for _, seg := range segments {
		switch seg.Type {
		case "text":
			parts = append(parts, seg.Data.Text)
		case "at":
			parts = append(parts, fmt.Sprintf("[CQ:at,qq=%s]", seg.Data.QQ))
		}
	}
	*m = messageField(strings.Join(parts, ""))
	return nil
}

// OneBot 11 事件结构体（仅解析需要的字段）。
type oneBotEvent struct {
	PostType    string       `json:"post_type"`
	MessageType string       `json:"message_type"`
	UserID      int64        `json:"user_id"`
	GroupID     int64        `json:"group_id"`
	Message     messageField `json:"message"`
}

type chatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

type napCatSendMsg struct {
	UserID  int64  `json:"user_id,omitempty"`
	GroupID int64  `json:"group_id,omitempty"`
	Message string `json:"message"`
}

var cqCodeRegexp = regexp.MustCompile(`\[CQ:[^\]]+\]`)

// stripCQCodes 剥离消息中所有 CQ 码，返回纯文本。
func stripCQCodes(msg string) string {
	return strings.TrimSpace(cqCodeRegexp.ReplaceAllString(msg, ""))
}

// containsAtBot 检查消息是否包含 @bot 的 CQ 码。
func containsAtBot(msg, botID string) bool {
	return strings.Contains(msg, fmt.Sprintf("[CQ:at,qq=%s]", botID))
}

// callAgentChat 调用 Amadeus Agent Server /chat 接口。
func callAgentChat(agentURL, conversationID, message string) (string, error) {
	reqBody, err := json.Marshal(chatRequest{
		ConversationID: conversationID,
		Message:        message,
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(agentURL+"/chat", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("call agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("agent returned %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode agent response: %w", err)
	}
	return result.Reply, nil
}

// sendNapCatMsg 通过 NapCat HTTP API 发送消息。
func sendNapCatMsg(napCatURL, token, endpoint string, payload napCatSendMsg) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, napCatURL+"/"+endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send to napcat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("napcat returned %d", resp.StatusCode)
	}
	return nil
}

func main() {
	_ = godotenv.Load()

	napCatURL := os.Getenv("NAPCAT_API_URL")
	if napCatURL == "" {
		napCatURL = "http://127.0.0.1:3000"
	}
	napCatToken := os.Getenv("NAPCAT_TOKEN")
	botID := os.Getenv("QQ_BOT_ID")
	agentURL := os.Getenv("AGENT_SERVER_URL")
	if agentURL == "" {
		agentURL = "http://localhost:9000"
	}

	if botID == "" {
		log.Fatal("QQ_BOT_ID 环境变量未设置")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}

		var event oneBotEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			log.Printf("解析事件失败: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("收到事件: post_type=%s message_type=%s user_id=%d group_id=%d message=%q",
			event.PostType, event.MessageType, event.UserID, event.GroupID, string(event.Message))

		// 只处理消息事件。
		if event.PostType != "message" {
			log.Printf("忽略非消息事件: post_type=%s", event.PostType)
			w.WriteHeader(http.StatusOK)
			return
		}

		var conversationID, textMessage string

		switch event.MessageType {
		case "private":
			conversationID = fmt.Sprintf("private:%d", event.UserID)
			textMessage = strings.TrimSpace(string(event.Message))

		case "group":
			if !containsAtBot(string(event.Message), botID) {
				log.Printf("群消息未@bot，忽略: group_id=%d", event.GroupID)
				w.WriteHeader(http.StatusOK)
				return
			}
			conversationID = fmt.Sprintf("group:%d", event.GroupID)
			textMessage = fmt.Sprintf("[用户 %d]: %s", event.UserID, stripCQCodes(string(event.Message)))

		default:
			log.Printf("忽略未知消息类型: %s", event.MessageType)
			w.WriteHeader(http.StatusOK)
			return
		}

		if textMessage == "" {
			log.Printf("[%s] 消息文本为空，忽略", conversationID)
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Printf("[%s] 处理消息: %q", conversationID, textMessage)
		w.WriteHeader(http.StatusOK)

		// 异步处理，避免 NapCat webhook 超时。
		go func() {
			reply, err := callAgentChat(agentURL, conversationID, textMessage)
			if err != nil {
				log.Printf("[%s] agent error: %v", conversationID, err)
				return
			}
			log.Printf("[%s] agent 回复: %q", conversationID, reply)

			var sendErr error
			if event.MessageType == "private" {
				sendErr = sendNapCatMsg(napCatURL, napCatToken, "send_private_msg", napCatSendMsg{
					UserID:  event.UserID,
					Message: reply,
				})
			} else {
				sendErr = sendNapCatMsg(napCatURL, napCatToken, "send_group_msg", napCatSendMsg{
					GroupID: event.GroupID,
					Message: reply,
				})
			}
			if sendErr != nil {
				log.Printf("[%s] napcat send error: %v", conversationID, sendErr)
			} else {
				log.Printf("[%s] 消息发送成功", conversationID)
			}
		}()
	})

	addr := ":8080"
	log.Printf("QQ Adapter 启动，监听 %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
