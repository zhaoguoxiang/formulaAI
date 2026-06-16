# FormulAI — 配方 AI 辅助系统

基于 Material Design 3 的配方管理系统，支持配方矩阵展示、数据分析看板、配方编辑器以及 AI 配方助手。

## 技术栈

| 层 | 技术 |
|---|------|
| 前端 | Angular 21 + Angular Material 3 (M3) |
| 后端 API | Go |
| AI 引擎 | CopilotKit (OpenAI 兼容) |
| 数据库 | PostgreSQL 16 |
| 部署 | Docker Compose |

## 快速部署

### 前提条件

- [Docker](https://docs.docker.com/get-docker/) >= 20.10
- [Docker Compose](https://docs.docker.com/compose/install/) >= 2.0
- OpenAI API Key（用于 AI 配方助手，可选）

### 1. 克隆仓库

```bash
git clone <repo-url> formulaAIsystem
cd formulaAIsystem
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env`，填入你的配置：

```env
# 数据库（默认即可，无需修改）
DB_USER=formula
DB_PASSWORD=changeme
DB_NAME=formula_ai
DB_PORT=5432

# OpenAI API Key（启用 AI 助手必须）
OPENAI_API_KEY=sk-your-key-here
```

> **注意**：不填 `OPENAI_API_KEY` 不影响页面访问，但 AI 配方助手聊天功能将不可用。后端和前端可正常运行。

### 3. 启动所有服务

```bash
docker compose up -d
```

首次启动会自动拉取镜像并构建前端/后端，大约需要 2-3 分钟。

### 4. 验证部署

```bash
docker compose ps
```

所有 4 个容器状态为 `healthy` 即部署成功：

```
NAME                 STATUS
formula-postgres     Up (healthy)
formula-backend      Up (healthy)
formula-frontend     Up (healthy)
formula-copilotkit   Up (healthy)
```

### 5. 访问

| 服务 | 地址 |
|------|------|
| 前端页面 | http://localhost:8081 |
| 后端 API | http://localhost:8080 |
| CopilotKit | http://localhost:3001 |

## 常用命令

```bash
# 停止所有服务
docker compose down

# 重启所有服务
docker compose restart

# 查看日志
docker compose logs -f frontend
docker compose logs -f backend

# 只重建前端（修改前端代码后）
docker compose up -d --build frontend

# 完全重建（包括数据库清空）
docker compose down -v
docker compose up -d --build
```

## 端口配置

如需修改端口，在 `.env` 中添加：

```env
FRONTEND_PORT=8081   # 默认
DB_PORT=5432          # 默认
```

然后 `docker compose up -d` 重新启动生效。

## 常见问题

**Q: 前端页面空白或样式异常？**

浏览器强刷新 (Ctrl+Shift+R)，或清除 Service Worker 缓存。

**Q: 配方数据为空？**

数据库首次启动会自动建表，但配方数据需要手动录入。使用"新增"标签页创建配方。

**Q: CopilotKit 连接失败？**

确认 `.env` 中 `OPENAI_API_KEY` 已正确配置，且 API Key 有可用额度。

**Q: 端口冲突？**

修改 `.env` 中的端口号，或停止占用端口的其他服务。

## 项目结构

```
├── frontend/            # Angular 21 + M3 前端
│   ├── src/
│   │   ├── styles.scss           # M3 Design Token 体系
│   │   └── app/components/       # 页面组件
│   └── Dockerfile
├── backend/             # Go API 服务
│   ├── cmd/server/
│   ├── migrations/               # 数据库迁移
│   └── Dockerfile
├── copilotkit-runtime/  # AI CopilotKit 运行时
│   └── Dockerfile
├── docker-compose.yml   # 容器编排
└── .env.example         # 环境变量模板
```
