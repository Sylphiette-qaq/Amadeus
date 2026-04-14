package basetools

import (
	"Amadeus/internal/skill"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

const (
	maxBashOutputBytes   = 32 * 1024
	defaultBashTimeout   = 15 * time.Second
	maxBashTimeoutSecond = 60
)

func Load(cfg skill.Config) []einotool.InvokableTool {
	return []einotool.InvokableTool{
		GetBashTool(),
		GetLoadSkillTool(cfg),
	}
}

func GetBashTool() einotool.InvokableTool {
	info := &schema.ToolInfo{
		Name: "bash",
		Desc: "执行本地 bash 命令。适用于查看文件、搜索代码、列目录、运行只读检查等更基础的系统操作。参数 command 为要执行的 bash 命令；workdir 可选，用于指定执行目录；timeout_seconds 可选，默认 15 秒，最大 60 秒。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {
				Desc:     "要执行的 bash 命令",
				Type:     schema.String,
				Required: true,
			},
			"workdir": {
				Desc: "命令执行目录，支持相对路径或绝对路径；默认当前工作目录",
				Type: schema.String,
			},
			"timeout_seconds": {
				Desc: "命令超时时间，单位秒；默认 15，最大 60",
				Type: schema.Integer,
			},
		}),
	}

	return toolutils.NewTool(info, runBash)
}

func runBash(ctx context.Context, params map[string]interface{}) (string, error) {
	command, err := getRequiredString(params, "command")
	if err != nil {
		return "", err
	}

	workdir, err := getOptionalString(params, "workdir")
	if err != nil {
		return "", err
	}

	if workdir != "" {
		workdir, err = filepath.Abs(workdir)
		if err != nil {
			return "", fmt.Errorf("解析 workdir 失败: %w", err)
		}
	}

	timeout, err := getTimeout(params)
	if err != nil {
		return "", err
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "bash", "-lc", command)
	if workdir != "" {
		cmd.Dir = workdir
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	timedOut := runCtx.Err() == context.DeadlineExceeded

	stdoutText, stdoutTruncated := truncateOutput(stdout.String(), maxBashOutputBytes)
	stderrText, stderrTruncated := truncateOutput(stderr.String(), maxBashOutputBytes)

	lines := []string{
		fmt.Sprintf("command: %s", command),
	}
	if cmd.Dir != "" {
		lines = append(lines, fmt.Sprintf("workdir: %s", cmd.Dir))
	}
	if runErr == nil {
		lines = append(lines, "exit_code: 0")
	} else if exitErr, ok := runErr.(*exec.ExitError); ok {
		lines = append(lines, fmt.Sprintf("exit_code: %d", exitErr.ExitCode()))
	} else {
		lines = append(lines, "exit_code: unknown")
	}

	lines = append(lines, "stdout:")
	if stdoutText == "" {
		lines = append(lines, "(empty)")
	} else {
		lines = append(lines, stdoutText)
	}
	if stdoutTruncated {
		lines = append(lines, fmt.Sprintf("... stdout truncated to first %d bytes", maxBashOutputBytes))
	}

	lines = append(lines, "stderr:")
	if stderrText == "" {
		lines = append(lines, "(empty)")
	} else {
		lines = append(lines, stderrText)
	}
	if stderrTruncated {
		lines = append(lines, fmt.Sprintf("... stderr truncated to first %d bytes", maxBashOutputBytes))
	}

	if timedOut {
		return strings.Join(lines, "\n"), fmt.Errorf("bash command timed out after %s", timeout)
	}
	if runErr != nil {
		return strings.Join(lines, "\n"), fmt.Errorf("bash command failed: %w", runErr)
	}

	return strings.Join(lines, "\n"), nil
}

func getRequiredString(params map[string]interface{}, key string) (string, error) {
	value, ok := params[key]
	if !ok {
		return "", fmt.Errorf("参数%s缺失", key)
	}

	strValue, ok := value.(string)
	if !ok || strings.TrimSpace(strValue) == "" {
		return "", fmt.Errorf("参数%s必须是非空字符串", key)
	}

	return strValue, nil
}

func getOptionalString(params map[string]interface{}, key string) (string, error) {
	value, ok := params[key]
	if !ok {
		return "", nil
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("参数%s必须是字符串", key)
	}

	return strings.TrimSpace(strValue), nil
}

func getTimeout(params map[string]interface{}) (time.Duration, error) {
	value, ok := params["timeout_seconds"]
	if !ok {
		return defaultBashTimeout, nil
	}

	timeoutSeconds, err := asInt64(value)
	if err != nil {
		return 0, fmt.Errorf("参数timeout_seconds必须是整数")
	}
	if timeoutSeconds <= 0 {
		return 0, fmt.Errorf("参数timeout_seconds必须大于0")
	}
	if timeoutSeconds > maxBashTimeoutSecond {
		return 0, fmt.Errorf("参数timeout_seconds不能超过%d", maxBashTimeoutSecond)
	}

	return time.Duration(timeoutSeconds) * time.Second, nil
}

func asInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float32:
		if float32(int64(v)) != v {
			return 0, fmt.Errorf("not integer")
		}
		return int64(v), nil
	case float64:
		if float64(int64(v)) != v {
			return 0, fmt.Errorf("not integer")
		}
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unsupported type %T", value)
	}
}

func truncateOutput(text string, maxBytes int) (string, bool) {
	data := []byte(text)
	if len(data) <= maxBytes {
		return text, false
	}

	return string(data[:maxBytes]), true
}
