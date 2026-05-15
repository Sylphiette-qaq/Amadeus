## 1. Dockerfile

- [x] 1.1 新建 `Dockerfile`，build stage 使用 `golang:1.25-alpine`，同时编译 `cmd/agent-server` 和 `cmd/qq` 两个二进制
- [x] 1.2 配置 runtime stage 使用 `node:22-alpine`，从 builder 复制两个二进制，设置 `WORKDIR /app`
- [x] 1.3 验证 `docker build -t amadeus .` 构建成功，镜像包含 `/app/agent-server`、`/app/qq` 和 `npx`

## 2. docker-compose.yml

- [x] 2.1 新建 `docker-compose.yml`，定义 `amadeus-network` 自定义网络
- [x] 2.2 添加 `napcat` 服务：镜像 `mlikiowa/napcat-docker:latest`，固定 `mac_address`，挂载 `./napcat-data/ntqq` 和 `./napcat-data/config`，暴露端口 3000、3001、6099
- [x] 2.3 添加 `amadeus-agent` 服务：`build: .`，`command: /app/agent-server`，`env_file: .env`，挂载 `./skills`、`./tools`、`./checkpoints`，暴露端口 9501，`depends_on` napcat 和 milvus
- [x] 2.4 添加 `qq-adapter` 服务：同一镜像，`command: /app/qq`，`env_file: .env`，暴露端口 8081，`depends_on` napcat 和 amadeus-agent
- [x] 2.5 将 `docker-compose.milvus.yml` 中的 etcd、minio、milvus、attu 四个服务迁移至新 compose 文件，调整 network 为 `amadeus-network`

## 3. 清理与配置更新

- [x] 3.1 删除 `docker-compose.milvus.yml`
- [x] 3.2 更新 `.env.example`：`NAPCAT_API_URL=http://napcat:3000`，`AGENT_SERVER_URL=http://amadeus-agent:9501`，`MILVUS_ADDRESS=milvus:19530`
- [x] 3.3 创建 `napcat-data/ntqq/` 和 `napcat-data/config/` 目录，添加 `.gitkeep`，在 `.gitignore` 中忽略目录内容但保留 `.gitkeep`

## 4. 验证

- [x] 4.1 执行 `NAPCAT_UID=$(id -u) NAPCAT_GID=$(id -g) docker compose up -d`，确认七个服务全部启动
- [x] 4.2 确认 `http://localhost:6099/webui` 可访问（NapCat WebUI）
- [x] 4.3 确认 `http://localhost:9501/chat` 端点可访问（amadeus-agent）
- [x] 4.4 `docker compose down && docker compose up -d` 后，`checkpoints/` 中已有会话历史仍然存在
