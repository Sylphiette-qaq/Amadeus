package basetools

import (
	"context"

	"Amadeus/internal/memory"

	einotool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

const (
	defaultSearchTopK = 5
	maxSearchTopK     = 20
	minSearchTopK     = 1
)

func GetSearchMemoryTool(idx *memory.Indexer) einotool.InvokableTool {
	info := &schema.ToolInfo{
		Name: "search_memory",
		Desc: "在历史对话记忆中进行语义搜索，检索跨 session 的相关历史消息片段。当需要回忆之前讨论过的内容时调用此工具。返回按相关度排序的历史消息列表，每条包含来源 session、轮次、角色和原文。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Desc:     "用于语义检索的查询文本，描述你想回忆的内容",
				Type:     schema.String,
				Required: true,
			},
			"top_k": {
				Desc:     "返回的最相关记录数量，范围 1–20，默认 5",
				Type:     schema.Integer,
				Required: false,
			},
		}),
	}

	return toolutils.NewTool(info, func(ctx context.Context, params map[string]interface{}) (string, error) {
		query, err := getRequiredString(params, "query")
		if err != nil {
			return "", err
		}

		topK := defaultSearchTopK
		if raw, ok := params["top_k"]; ok && raw != nil {
			if v, ok := toInt(raw); ok {
				topK = v
			}
		}
		if topK < minSearchTopK {
			topK = minSearchTopK
		}
		if topK > maxSearchTopK {
			topK = maxSearchTopK
		}

		return idx.Search(ctx, query, topK)
	})
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	}
	return 0, false
}
