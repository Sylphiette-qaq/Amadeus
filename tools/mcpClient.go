package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPServerConfig MCP服务器配置
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// ToolsConfig 工具配置总结构
type ToolsConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// LoadToolsConfig 从JSON文件加载工具配置
// 参数:
//
//	filePath: 配置文件路径
//
// 返回:
//
//	工具配置结构
//	错误信息
func LoadToolsConfig(filePath string) (*ToolsConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config ToolsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// McpClientFromConfig 根据配置创建MCP客户端
// 参数:
//
//	ctx: 上下文
//	serverName: 服务器名称（配置文件中的key）
//	config: MCP服务器配置
//
// 返回:
//
//	MCP客户端
//	错误信息
func McpClientFromConfig(ctx context.Context, serverName string, config MCPServerConfig) (client.MCPClient, error) {
	// 构建环境变量列表
	envs := make([]string, 0, len(config.Env))
	for key, value := range config.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	// 创建MCP客户端
	cli, err := client.NewStdioMCPClient(config.Command, envs, config.Args...)
	if err != nil {
		return nil, fmt.Errorf("创建MCP客户端失败: %v", err)
	}

	// 初始化客户端
	_, err = cli.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("初始化MCP客户端失败: %v", err)
	}

	return cli, nil
}

// McpClient 创建MCP客户端（兼容旧代码）
// 参数:
//
//	ctx: 上下文
//	command: 执行命令
//	envs: 环境变量列表
//	args: 命令参数
//
// 返回:
//
//	MCP客户端
//	错误信息
func McpClient(ctx context.Context, command string, envs []string, args ...string) (client.MCPClient, error) {
	cli, err := client.NewStdioMCPClient(command, envs, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	_, err = cli.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %v", err)
	}

	return cli, nil
}

// CreateMcpClientsFromConfig 从配置文件创建所有MCP客户端
// 参数:
//
//	ctx: 上下文
//	configPath: 配置文件路径
//
// 返回:
//
//	MCP客户端列表
//	错误信息
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
