---
name: metadata-platform-cli
description: 当用户需要使用 metadata-platform-service-cli 查询支持的 route_v1 API、查看指定 API 说明、查看 APIJSON 语法、生成调用参数、生成 APIJSON 请求体或执行调用时，使用这个 skill。
---

# Metadata Platform CLI

当用户要使用 `metadata-platform-service-cli` 时，使用这个 skill。

## 强制规则

- 必须通过 CLI 可执行程序 `metadata-platform-service-cli` 调用；优先使用仓库中的 `metadata-platform-service-cli`
- 禁止使用 `go run .`、`go run main.go` 或任何其他 `go run` 形式
- 在回复中的命令示例、执行步骤和说明文字里，都不要出现 `go run`
- 只有在真正执行 `invoke` 时，才需要 `base_url` 和 `token`
- 当用户要执行 `invoke` 但尚未提供 `base_url` 或 `token` 时，必须先主动向用户索取，不能自行假设、不能跳过
- 当请求体是 agent 新生成或 agent 修改后的 APIJSON 时，必须先把待执行的 body 明确展示给用户，并获得用户确认后，才能执行 `invoke`
- 如果用户只是让 agent 帮忙“生成 APIJSON”或“拼装命令”，默认只生成，不执行
- 如果用户明确要求“直接执行”，但 body 属于 agent 生成的 APIJSON，仍然必须先展示 body 并等待确认
- 对固定结构接口，如果 body 只是把用户明确给出的字段原样组装，可以直接执行；但只要存在字段猜测、条件补全、APIJSON 结构推断，就必须先确认

适用场景：

- 查询有哪些逻辑 API
- 查看某个 API 的用途、参数和示例
- 查看 APIJSON 语法说明和模板
- 拼装 `invoke` 命令
- 生成 APIJSON 请求体
- 实际调用 route_v1 API
- 理解 CLI 返回的 JSON 和退出码

## CLI 能力

核心命令：

- `apis`：列出支持的逻辑 API
- `api-help --api <name>`：查看指定 API 说明
- `apijson-help`：查看 APIJSON 语法、操作符、功能键和模板
- `invoke --base-url ... --token ... --api ...`：调用指定 API

不负责：

- 登录
- 获取或刷新 token
- 本地缓存 token
- `route_admin.php` 或其他非 `route_v1` 路由

## 推荐顺序

优先按这个顺序工作：

1. 不知道 API 名称时，先查：
   ```bash
   metadata-platform-service-cli apis
   ```

2. 知道 API 名称后，先看说明：
   ```bash
   metadata-platform-service-cli api-help --api business-metadata.query --format json
   ```

3. 如果 `api-help --format json` 返回中包含 `"body_style": "apijson"`，再看 APIJSON 语法：
   ```bash
   metadata-platform-service-cli apijson-help --format json
   ```

4. 对 `business-data.query`，优先按 `{"TableName[]": {...}}` 生成查询体

5. 如果用户要“生成请求体”或“帮忙拼装调用”，先产出 `params` / `body` 草案，不要直接执行

6. 如果 `body_style` 是 `apijson`，并且请求体由 agent 生成、补全或修改，先向用户确认待执行的 APIJSON

7. 只有在需要执行 `invoke` 时，才确认用户是否已经提供 `base_url` 和 `token`；缺少时先向用户索取

8. 用户确认后再调用：
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
- 对动态 `business-data.*` 接口，优先通过 `body_style` / `syntax_ref` 判断是否需要再查 `apijson-help`
- 对动态 `business-data.*` 接口，优先把工作拆成三步：识别接口 -> 生成 APIJSON -> 用户确认 -> 执行
- 对 `business-data.query`，优先使用 `{"TableName[]": {...}}` 作为顶层查询体
- `api-help`、`apijson-help`、`apis` 可以直接执行，不需要用户确认
- `invoke` 是否需要确认，取决于请求体是否由 agent 推断或生成；其中 APIJSON 一律按高风险处理

## 标准流程

### 流程 A：发现 API

适用：
- 用户不知道 API 名称
- 用户只知道业务目标，不知道该调用哪个逻辑 API

动作：
1. 运行 `apis`
2. 根据候选 API 运行一个或多个 `api-help --format json`
3. 把候选 API 和理由告诉用户

### 流程 B：查看说明

适用：
- 用户已经给出 API 名称
- 用户想知道路径、方法、关键字段、幂等性或示例

动作：
1. 运行 `api-help --api <name> --format json`
2. 如果返回 `body_style=apijson`，继续运行 `apijson-help --format json`
3. 对 `business-data.query`，明确提示查询体使用 `TableName[]` 顶层 key
4. 先解释接口契约，再决定是否进入“生成”或“执行”

### 流程 C：生成调用参数

适用：
- 用户要 agent 帮忙拼 `params`
- 用户要 agent 帮忙拼 `body`
- 用户要 agent 帮忙写 APIJSON

动作：
1. 先查 `api-help`
2. 如有 `apijson`，再查 `apijson-help`
3. 对 `business-data.query`，优先生成 `{"TableName[]": {...}}` 形式的 body
4. 输出建议的 `params` / `body` / `method`
5. 不执行 `invoke`，除非用户明确要求

### 流程 D：执行调用

适用：
- 用户明确要求实际调用接口

动作：
1. 确认 `base_url`、`token`
2. 确认 `api-help` 已经查过，接口语义清楚
3. 如果请求体是 agent 生成或改写的 APIJSON，先展示完整 body 并向用户确认
4. 用户确认后，再执行 `invoke`
5. 返回退出码、关键 JSON 字段、业务结果和失败原因

## APIJSON 确认规则

以下任一情况都必须先确认，再执行：

- agent 根据自然语言新写了 APIJSON body
- agent 根据 APIJSON 模板推导了实体名、字段名、`@combine`、`@conditions`
- agent 对用户已有 APIJSON 做了修改
- agent 根据接口说明补全了用户未明确给出的过滤条件、分页、排序、聚合

确认时必须包含：

- 目标 API 名称
- HTTP 方法
- `params`（如果有）
- 完整 `body` 或 `body-file` 内容
- 一句简短说明：“这是将要执行的请求体，请确认是否执行”

如果用户只说“看起来可以”或“执行吧”，即可继续；如果用户提出修改，先修改后再次确认。

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

查询 APIJSON 语法：

```bash
metadata-platform-service-cli apijson-help --format json
```

生成但不执行 APIJSON：

```bash
metadata-platform-service-cli api-help --api business-data.query --format json
metadata-platform-service-cli apijson-help --format json
```

然后把建议的 body 发给用户确认，不要直接运行 `invoke`。

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

动态 APIJSON 查询：

```bash
metadata-platform-service-cli api-help --api business-data.query --format json
metadata-platform-service-cli apijson-help --format json
metadata-platform-service-cli invoke \
  --base-url http://127.0.0.1:9501 \
  --token "$TOKEN" \
  --api business-data.query \
  --body '{"demo_resource[]":{"status{}":["pending","approved"],"@column":"id,order_no,status","@order":"created_at-","@offset":0,"@limit":20,"@combine":"status{}","@method":"GET"}}'
```

但对上面这类命令，agent 必须先把 `--body` 展示给用户确认，再执行。

## 工作原则

不要猜 API 语义。

- 不要在仅查询 `apis`、`api-help`、`apijson-help` 时要求 `base_url` 或 `token`
- 不要跳过对 `base_url` 和 `token` 的确认；只有在需要执行 `invoke` 且缺少时才向用户索取
- 不要输出 `go run` 形式的命令；只使用 `metadata-platform-service-cli`
- 不要把“生成请求体”和“执行请求”混成一步
- 不要在未经确认的情况下执行 agent 生成的 APIJSON
- 不要猜字段名、实体名或 `@combine`；拿不准时先查 `api-help` / `apijson-help`，再向用户确认
- 先 `apis`
- 再 `api-help`
- 如有需要再 `apijson-help`
- 生成 `params` / `body`
- 用户确认
- 最后 `invoke`
