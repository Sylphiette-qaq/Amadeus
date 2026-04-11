package tool

import (
	"context"
	"fmt"

	einotool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

func GetCalculatorTool() einotool.InvokableTool {
	info := &schema.ToolInfo{
		Name: "add",
		Desc: "当用户请求进行数学计算、加法运算或询问两个数字的和时，必须调用此工具。支持整数和浮点数计算。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"num1": {
				Desc:     "第一个数字",
				Type:     schema.Number,
				Required: true,
			},
			"num2": {
				Desc:     "第二个数字",
				Type:     schema.Number,
				Required: true,
			},
		}),
	}

	return toolutils.NewTool(info, calculatorFunc)
}

func calculatorFunc(_ context.Context, params map[string]interface{}) (string, error) {
	num1Val, ok := params["num1"]
	if !ok {
		return "", fmt.Errorf("参数num1缺失")
	}

	num2Val, ok := params["num2"]
	if !ok {
		return "", fmt.Errorf("参数num2缺失")
	}

	sum := num1Val.(float64) + num2Val.(float64)
	return fmt.Sprintf("计算结果：%f + %f = %f", num1Val, num2Val, sum), nil
}
