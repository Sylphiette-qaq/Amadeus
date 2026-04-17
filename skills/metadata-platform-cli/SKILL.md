---
name: metadata-platform-cli
description: 当用户需要使用 metadata-platform-service-cli 查询支持的 route_v1 API、查看指定 API 说明、拼装 invoke 参数或执行调用时，使用这个 skill。
---

# Metadata Platform CLI

当用户要使用 `metadata-platform-service-cli` 时，使用这个 skill。

## 强制规则

- 当用户要查询 API、查看 API 说明、拼装调用命令或实际调用时，必须先确认用户是否已经提供 `base_url` 和 `token`
- 如果用户尚未提供 `base_url` 或 `token`，必须先主动向用户索取，不能自行假设、不能跳过
- 必须通过 CLI 可执行程序 `metadata-platform-service-cli` 调用
- 禁止使用 `go run .`、`go run main.go` 或任何其他 `go run` 形式
- 在回复中的命令示例、执行步骤和说明文字里，都不要出现 `go run`

适用场景：

- 查询有哪些逻辑 API
- 查看某个 API 的用途、参数和示例
- 拼装 `invoke` 命令
- 实际调用 route_v1 API
- 理解 CLI 返回的 JSON 和退出码

## CLI 能力

核心命令：

- `apis`：列出支持的逻辑 API
- `api-help --api <name>`：查看指定 API 说明
- `invoke --base-url ... --token ... --api ...`：调用指定 API

不负责：

- 登录
- 获取或刷新 token
- 本地缓存 token
- `route_admin.php` 或其他非 `route_v1` 路由

## 推荐顺序

优先按这个顺序工作：

1. 先确认用户是否已经提供 `base_url` 和 `token`；如果没有，先向用户索取

2. 不知道 API 名称时，先查：
   ```bash
   metadata-platform-service-cli apis
   ```

3. 知道 API 名称后，先看说明：
   ```bash
   metadata-platform-service-cli api-help --api business-metadata.query --format json
   ```

4. 确认参数后再调用：
   ```bash
   metadata-platform-service-cli invoke \
     --base-url http://127.0.0.1:9501 \
     --token "$TOKEN" \
     --api business-metadata.query \
     --body '{"page":1,"page_size":10}'
   ```

## 规则

- `invoke` 必填：`--base-url`、`--token`、`--api`
- `--body` 和 `--body-file` 互斥
- `--params` 必须是 JSON object
- 路径参数必须通过 `--params` 提供
- 多 method API 必须传 `--method`
- `invoke` 输出始终为 JSON
- 对 agent，优先使用 `api-help --format json`

## 错误判断

退出码：

- `0`：成功
- `1`：网络、HTTP、解析或业务错误
- `2`：本地参数错误

失败时优先看：

- `error.type`
- `error.message`
- `response.http_status`
- `response.body`

不要把 HTTP 200 直接当成业务成功。

## 常用模式

查询 API 说明：

```bash
metadata-platform-service-cli api-help --api business-metadata.query --format json
```

带路径参数调用：

```bash
metadata-platform-service-cli invoke \
  --base-url http://127.0.0.1:9501 \
  --token "$TOKEN" \
  --api business-metadata.get \
  --params '{"code":"demo_meta"}'
```

多 method 调用：

```bash
metadata-platform-service-cli invoke \
  --base-url http://127.0.0.1:9501 \
  --token "$TOKEN" \
  --api business-data.tcc-confirm \
  --method POST \
  --body '{"resource":"demo","branch_id":"b1","gid":"g1"}'
```

## 工作原则

不要猜 API 语义。

- 不要跳过对 `base_url` 和 `token` 的确认；缺少时先向用户索取
- 不要输出 `go run` 形式的命令；只使用 `metadata-platform-service-cli`
- 先 `apis`
- 再 `api-help`
- 最后 `invoke`
