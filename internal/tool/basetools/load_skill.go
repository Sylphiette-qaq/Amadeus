package basetools

import (
	"Amadeus/internal/skill"
	"context"

	einotool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

func GetLoadSkillTool(cfg skill.Config) einotool.InvokableTool {
	info := &schema.ToolInfo{
		Name: "load_skill",
		Desc: "按 skill 名称加载完整的 SKILL.md 指令，并把该 skill 持久激活到当前会话上下文中。加载成功后，本轮后续推理和后续轮次都会继续携带这个 skill。一次只加载一个 skill，不要批量加载全部 skill。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name": {
				Desc:     "skill 名称，必须与 agent.md 中注册的 name 一致",
				Type:     schema.String,
				Required: true,
			},
		}),
	}

	return toolutils.NewTool(info, func(ctx context.Context, params map[string]interface{}) (skill.Document, error) {
		name, err := getRequiredString(params, "name")
		if err != nil {
			return skill.Document{}, err
		}

		return skill.LoadSkillContent(cfg, name)
	})
}
