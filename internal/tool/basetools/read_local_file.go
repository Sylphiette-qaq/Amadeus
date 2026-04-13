package basetools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

const maxReadFileBytes = 32 * 1024
const maxReadDirEntries = 200

func Load() []einotool.InvokableTool {
	return []einotool.InvokableTool{
		GetReadLocalFileTool(),
		GetReadLocalDirectoryTool(),
	}
}

func GetReadLocalFileTool() einotool.InvokableTool {
	info := &schema.ToolInfo{
		Name: "read_local_file",
		Desc: "读取本地文件内容。适用于查看项目中的源码、配置或文档文件。参数 path 必须是文件路径，建议优先传项目内相对路径。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Desc:     "要读取的本地文件路径，支持相对路径或绝对路径",
				Type:     schema.String,
				Required: true,
			},
		}),
	}

	return toolutils.NewTool(info, readLocalFile)
}

func GetReadLocalDirectoryTool() einotool.InvokableTool {
	info := &schema.ToolInfo{
		Name: "read_local_directory",
		Desc: "读取本地目录内容，返回目录下的文件和子目录列表。适用于浏览项目结构、确认文件位置。参数 path 必须是目录路径；recursive 控制是否递归子目录；include_hidden 控制是否显示以 . 开头的隐藏文件或目录。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Desc:     "要读取的本地目录路径，支持相对路径或绝对路径",
				Type:     schema.String,
				Required: true,
			},
			"recursive": {
				Desc: "是否递归读取子目录，默认 false",
				Type: schema.Boolean,
			},
			"include_hidden": {
				Desc: "是否包含隐藏文件和隐藏目录，默认 false",
				Type: schema.Boolean,
			},
		}),
	}

	return toolutils.NewTool(info, readLocalDirectory)
}

func readLocalFile(_ context.Context, params map[string]interface{}) (string, error) {
	pathValue, ok := params["path"]
	if !ok {
		return "", fmt.Errorf("参数path缺失")
	}

	path, ok := pathValue.(string)
	if !ok || path == "" {
		return "", fmt.Errorf("参数path必须是非空字符串")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("解析路径失败: %w", err)
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("读取文件信息失败: %w", err)
	}
	if fileInfo.IsDir() {
		return "", fmt.Errorf("目标路径是目录，不是文件: %s", absPath)
	}

	// 基础工具只返回有限大小的内容，避免一次性把超大文件推给模型。
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	if len(data) > maxReadFileBytes {
		data = data[:maxReadFileBytes]
	}

	return fmt.Sprintf("path: %s\ncontent:\n%s", absPath, string(data)), nil
}

func readLocalDirectory(_ context.Context, params map[string]interface{}) (string, error) {
	pathValue, ok := params["path"]
	if !ok {
		return "", fmt.Errorf("参数path缺失")
	}

	path, ok := pathValue.(string)
	if !ok || path == "" {
		return "", fmt.Errorf("参数path必须是非空字符串")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("解析路径失败: %w", err)
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("读取目录信息失败: %w", err)
	}
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("目标路径不是目录: %s", absPath)
	}

	recursive, err := getOptionalBool(params, "recursive")
	if err != nil {
		return "", err
	}

	includeHidden, err := getOptionalBool(params, "include_hidden")
	if err != nil {
		return "", err
	}

	entries, truncated, err := collectDirectoryEntries(absPath, recursive, includeHidden, maxReadDirEntries)
	if err != nil {
		return "", err
	}

	lines := make([]string, 0, len(entries)+4)
	lines = append(lines, fmt.Sprintf("path: %s", absPath))
	lines = append(lines, fmt.Sprintf("recursive: %t", recursive))
	lines = append(lines, fmt.Sprintf("include_hidden: %t", includeHidden))
	lines = append(lines, "entries:")
	for _, entry := range entries {
		lines = append(lines, fmt.Sprintf("- [%s] %s", entry.entryType, entry.relPath))
	}

	if truncated {
		lines = append(lines, fmt.Sprintf("... truncated, showing first %d entries", len(entries)))
	}

	return strings.Join(lines, "\n"), nil
}

type directoryEntry struct {
	entryType string
	relPath   string
}

func collectDirectoryEntries(absPath string, recursive, includeHidden bool, limit int) ([]directoryEntry, bool, error) {
	if recursive {
		var entries []directoryEntry
		truncated := false

		err := filepath.WalkDir(absPath, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if path == absPath {
				return nil
			}

			name := d.Name()
			if !includeHidden && strings.HasPrefix(name, ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			relPath, err := filepath.Rel(absPath, path)
			if err != nil {
				return fmt.Errorf("计算相对路径失败: %w", err)
			}

			entryType := "file"
			if d.IsDir() {
				entryType = "dir"
			}

			if len(entries) >= limit {
				truncated = true
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			entries = append(entries, directoryEntry{
				entryType: entryType,
				relPath:   relPath,
			})
			return nil
		})
		if err != nil {
			return nil, false, fmt.Errorf("递归读取目录失败: %w", err)
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].relPath < entries[j].relPath
		})
		return entries, truncated, nil
	}

	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, false, fmt.Errorf("读取目录失败: %w", err)
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})

	entries := make([]directoryEntry, 0, len(dirEntries))
	for _, entry := range dirEntries {
		if !includeHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		entryType := "file"
		if entry.IsDir() {
			entryType = "dir"
		}

		entries = append(entries, directoryEntry{
			entryType: entryType,
			relPath:   entry.Name(),
		})
	}

	truncated := false
	if len(entries) > limit {
		entries = entries[:limit]
		truncated = true
	}

	return entries, truncated, nil
}

func getOptionalBool(params map[string]interface{}, key string) (bool, error) {
	value, ok := params[key]
	if !ok {
		return false, nil
	}

	booleanValue, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("参数%s必须是布尔值", key)
	}

	return booleanValue, nil
}
