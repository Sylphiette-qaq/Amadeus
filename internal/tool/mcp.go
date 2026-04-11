package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type ToolsConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

func LoadToolsConfig(filePath string) (*ToolsConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config ToolsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	resolveConfigEnv(&config)

	return &config, nil
}

func McpClientFromConfig(ctx context.Context, serverName string, config MCPServerConfig) (client.MCPClient, error) {
	envs := make([]string, 0, len(config.Env))
	for key, value := range config.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	cli, err := client.NewStdioMCPClient(config.Command, envs, config.Args...)
	if err != nil {
		return nil, fmt.Errorf("创建MCP客户端失败: %v", err)
	}

	_, err = cli.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("初始化MCP客户端失败: %v", err)
	}

	return cli, nil
}

func CreateMcpClientsFromConfig(ctx context.Context, configPath string) ([]client.MCPClient, error) {
	config, err := LoadToolsConfig(configPath)
	if err != nil {
		return nil, err
	}

	clients := make([]client.MCPClient, 0, len(config.MCPServers))
	for serverName, serverConfig := range config.MCPServers {
		cli, err := McpClientFromConfig(ctx, serverName, serverConfig)
		if err != nil {
			return nil, fmt.Errorf("创建服务器 %s 的客户端失败: %v", serverName, err)
		}
		clients = append(clients, cli)
	}

	return clients, nil
}

func resolveConfigEnv(config *ToolsConfig) {
	for serverName, serverConfig := range config.MCPServers {
		if len(serverConfig.Env) == 0 {
			continue
		}

		resolvedEnv := make(map[string]string, len(serverConfig.Env))
		for key, value := range serverConfig.Env {
			resolvedEnv[key] = resolveEnvPlaceholder(value)
		}

		serverConfig.Env = resolvedEnv
		config.MCPServers[serverName] = serverConfig
	}
}

func resolveEnvPlaceholder(value string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envKey := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
		return os.Getenv(envKey)
	}

	return value
}
