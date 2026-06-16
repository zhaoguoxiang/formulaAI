# 配方AI辅助系统 - 单机版 (Standalone Formula AI Assistant)

## TL;DR

> **Quick Summary**: 构建单用户单项目的配方AI辅助系统（单机版）。单页面：左侧工作区（配方展示/分析/编辑3个Tab）+ 右侧CopilotKit LLM聊天窗口。LLM通过MCP工具操作完整的配方CRUD。
>
> **Deliverables**:
> - Go Gin REST API（配方CRUD + 矩阵查询 + 分析数据）
> - Angular SPA + Angular Material UI（3-Tab工作区 + CopilotKit聊天面板 + 现代化风格）
> - CopilotKit Runtime（Node.js，OpenAI集成，MCP工具）
> - PostgreSQL数据库 + Migration
> - Docker Compose 一键部署
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 5 waves
> **Critical Path**: Foundation → Backend API → AI Runtime → Frontend Chat → Docker Integration → Final Review

---

## Context

### Original Request
参考 `../formulaAIsystem` 做一个单机版的配方AI辅助系统。只有一个页面（配方工作台），左侧工作区（多Tabs），右侧LLM聊天窗口。LLM可操作一切，预留MCP。技术栈 Go + Angular + PostgreSQL，使用 CopilotKit 做LLM集成。

### Interview Summary
**Key Discussions**:
- 与参考项目差异大：单用户（无需RBAC），单项目（无需项目管理），Go+Angular重写
- 工作区3个Tab：配方展示（矩阵表）、配方分析（图表仪表盘）、新增（配方编辑）
- LLM Provider: OpenAI，通过CopilotKit集成
- MCP工具范围：完整CRUD（创建/查询/修改/删除配方、配料、步骤、投料动作）
- 部署：Docker Compose一键启动
- 测试：Full TDD（Go backend + Angular frontend）

**Research Findings**:
- 参考项目为 Python/Django + React（非Go/Angular），完全重写
- 参考配方工作台：3个Tab（展示/分析/推荐），推荐由Gemini直接调用
- CopilotKit **官方支持Angular**：`@copilotkit/{core,angular}` npm包，`provideCopilotKit()` DI注入
- 配方领域模型（来自参考）：Formula → FormulaPart(A/B/MAIN) → FormulaIngredient → FormulaStep → FormulaDosingAction
- 验证规则：配料百分比总和=100%，投料比例总和=100%，启用前至少1个步骤

### Metis Review
> Metis服务不稳定（两次超时），跳过。需求已通过多轮用户问答充分明确，无遗留歧义。

---

## Work Objectives

### Core Objective
构建单用户、单项目的配方AI辅助系统，左侧工作区提供配方展示/分析/编辑功能，右侧CopilotKit聊天面板通过MCP工具实现LLM驱动的配方全操作。

### Concrete Deliverables
- Go Gin REST API (`backend/`): 7个API端点
- Angular SPA (`frontend/`): 单页面 3-Tab工作区 + CopilotKit聊天面板
- CopilotKit Runtime (`copilotkit-runtime/`): Node.js服务，OpenAI集成
- PostgreSQL Schema: 6张表 + migration
- Docker Compose: 4个service（Go API + Angular/Nginx + PostgreSQL + CopilotKit Runtime）
- Seed Data: 示例配方数据

### Definition of Done
- [ ] `docker compose up` 一键启动所有服务
- [ ] 配方CRUD完整可用（创建/查看/修改/删除）
- [ ] CopilotKit聊天面板可对话，LLM可通过工具操作配方数据
- [ ] 3个Tab功能完整（展示矩阵、分析图表、编辑器）
- [ ] `go test ./...` 全部通过
- [ ] `ng test` 全部通过

### Must Have
- 配方CRUD完整（Formula → Part → Ingredient → Step → DosingAction）
- CopilotKit LLM聊天窗口 + MCP工具（完整CRUD操作）
- 3-Tab工作区（矩阵展示 / 分析图表 / 配方编辑器）
- Angular Material UI + 现代化设计风格
- Docker Compose一键部署
- PostgreSQL持久化存储
- 配方验证规则（配料%总和=100%，投料比例总和=100%）

### Must NOT Have (Guardrails)
- 多用户认证/RBAC系统
- 项目管理（项目列表、项目详情、项目切换）
- 同步服务（外部数据源同步）
- 检测/指标/样本模块的完整CRUD
- 过度抽象：只做配方工作台需要的数据模型
- 硬编码分析数据（参考项目的问题，新版本从后端查询真实数据）

---

## Verification Strategy (MANDATORY)

### Test Decision
- **Infrastructure exists**: NO（当前为空项目，需从零搭建）
- **Automated tests**: TDD
- **Framework**: Go: `go test` + `testify`; Angular: `karma` + `jasmine`
- **TDD Flow**: 每个任务遵循 RED（先写失败测试）→ GREEN（最小实现）→ REFACTOR

### QA Policy
Every task MUST include agent-executed QA scenarios (see TODO template below).
Evidence saved to `.omo/evidence/task-{N}-{scenario-slug}.{ext}`.

- **API/Backend**: Use Bash (curl) - Send requests, assert status + response fields
- **Frontend/UI**: Use Playwright (playwright skill) - Navigate, interact, assert DOM, screenshot
- **CLI**: Use Bash - Run commands, validate output
- **Database**: Use Bash (psql) - Query tables, verify data integrity

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - Start Immediately, 6 tasks):
├── Task 1: Go project scaffold + module init
├── Task 2: Angular project scaffold + CopilotKit packages
├── Task 3: Docker Compose skeleton
├── Task 4: PostgreSQL schema + migration system
├── Task 5: Go domain types + validation rules
└── Task 6: Go database connection + config management

Wave 2 (Backend API - After Wave 1, 5 tasks):
├── Task 7: Formula CRUD repository (PostgreSQL)
├── Task 8: Formula validation service (business rules)
├── Task 9: Formula REST API handlers (Gin routes)
├── Task 10: Formula matrix query API
└── Task 11: Analysis data API endpoints

Wave 3 (Frontend UI - After Wave 1, parallel with Wave 2, 5 tasks):
├── Task 12: Angular layout (workspace left + chat right split)
├── Task 13: Angular API service + type definitions
├── Task 14: Formula Matrix tab component
├── Task 15: Formula Analysis tab component
└── Task 16: Formula Editor tab component

Wave 4 (AI Integration - After Wave 2+3, 4 tasks):
├── Task 17: CopilotKit runtime setup (Node.js)
├── Task 18: MCP tool definitions (formula CRUD actions)
├── Task 19: OpenAI integration (API client + streaming)
└── Task 20: Angular CopilotKit chat panel

Wave 5 (Integration + Polish - After Wave 4, 4 tasks):
├── Task 21: Docker Compose production config
├── Task 22: Seed data script
├── Task 23: Angular production build + Nginx config
└── Task 24: End-to-end integration verification

Wave FINAL (After ALL tasks — 4 parallel reviews):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
```

**Critical Path**: Task 1 → Task 5 → Task 7 → Task 9 → Task 17 → Task 18 → Task 20 → Task 24 → F1-F4
**Max Concurrent**: 6 (Waves 1, 2, 3)

---

## TODOs

- [x] 1. Go Project Scaffold + Module Init

  **What to do**:
  - Init Go module: `go mod init formula-ai-system/backend`
  - Create directory structure: `cmd/`, `internal/{config,database,models,repository,services,handlers,middleware}`, `migrations/`
  - Create `cmd/server/main.go` with Gin engine init, health check endpoint `GET /api/health`
  - Install dependencies: `gin-gonic/gin`, `lib/pq`, `golang-migrate/migrate`, `stretchr/testify`
  - Create `Makefile` with targets: `run`, `test`, `migrate-up`, `migrate-down`
  - TDD: Write test for health endpoint → implement → verify

  **Must NOT do**:
  - No auth middleware（单用户，无需认证）
  - No CORS middleware yet（加到Task 9）
  - No business logic（纯脚手架）

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准Go项目初始化，无需特殊技能）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4, 5, 6)
  - **Blocks**: 7, 8, 9, 10, 11
  - **Blocked By**: None

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./...` → FAIL（健康检查端点尚未实现）
  - [ ] 健康检查端点返回 `{"status": "ok"}` → `curl http://localhost:8080/api/health`
  - [ ] `go test ./...` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Health check endpoint returns OK
    Tool: Bash (curl)
    Preconditions: Go server running on port 8080
    Steps:
      1. curl -s http://localhost:8080/api/health
      2. Verify HTTP status code = 200
      3. Verify response body contains "status":"ok" (jq '.status')
    Expected Result: {"status":"ok"} with 200
    Failure Indicators: Connection refused, non-200 status, missing status field
    Evidence: .omo/evidence/task-1-health-check.json

  Scenario: Health check fails when server not running
    Tool: Bash (curl)
    Preconditions: Go server stopped
    Steps:
      1. curl -s http://localhost:8080/api/health
      2. Verify exit code != 0
    Expected Result: Connection refused
    Evidence: .omo/evidence/task-1-health-check-error.txt
  ```

  **Commit**: YES
  - Message: `feat(backend): project scaffold with gin health check`
  - Files: `backend/`

- [x] 2. Angular Project Scaffold + CopilotKit Packages

  **What to do**:
  - Init Angular project: `ng new frontend --standalone --style=css --routing=true`
  - Install Angular Material: `ng add @angular/material`（选择 Indigo/Pink 或自定义主题）
  - Install CopilotKit Angular packages: `npm install @copilotkit/core @copilotkit/angular`
  - Configure `provideCopilotKit()` in `app.config.ts` with placeholder `runtimeUrl`
  - Create base `app.component` with Material split layout：`mat-sidenav-container` + workspace + chat panel
  - Install Angular dev dependencies: `@playwright/test`（for QA）
  - TDD: Write test for app component rendering → implement → verify

  **Must NOT do**:
  - No actual chat panel UI yet（Task 20）
  - No tabs implementation yet（Tasks 14-16）
  - No real CopilotKit runtime URL（占位符）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-design`]
    - `frontend-design`: Angular项目布局和样式设计

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4, 5, 6)
  - **Blocks**: 12, 13, 14, 15, 16, 20
  - **Blocked By**: None

  **Acceptance Criteria**:
  - [ ] TDD: `ng test --watch=false` → FAIL（app component渲染测试尚未通过）
  - [ ] `ng serve` → 浏览器访问 `http://localhost:4200` 看到 split layout
  - [ ] `ng test --watch=false` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Angular app boots and renders split layout
    Tool: Playwright
    Preconditions: `ng serve` running on port 4200
    Steps:
      1. Navigate to http://localhost:4200
      2. Wait for app root element to render (selector: 'app-root')
      3. Verify two child divs exist (workspace panel + chat panel placeholder)
      4. Screenshot the page
    Expected Result: Page shows split layout with two panels
    Failure Indicators: Blank page, error in console, missing layout divs
    Evidence: .omo/evidence/task-2-app-boot.png

  Scenario: ng test passes all unit tests
    Tool: Bash
    Preconditions: Angular project initialized with test file
    Steps:
      1. cd frontend && ng test --watch=false --browsers=ChromeHeadless
      2. Verify exit code = 0
      3. Verify output shows "SUCCESS"
    Expected Result: All tests pass, exit 0
    Evidence: .omo/evidence/task-2-ng-test.txt
  ```

  **Commit**: YES
  - Message: `feat(frontend): angular project init with material ui and copilotkit`
  - Files: `frontend/`

- [x] 3. Docker Compose Skeleton

  **What to do**:
  - Create `docker-compose.yml` with 4 services:
    - `postgres`: PostgreSQL 16, port 5432, volume for data persistence, healthcheck
    - `backend`: Go API, port 8080, build from `backend/Dockerfile`, depends_on postgres, env vars
    - `frontend`: Angular dev server, port 4200, build from `frontend/Dockerfile.dev`
    - `copilotkit-runtime`: Node.js, port 3001, placeholder（Task 17完善）
  - Create `backend/Dockerfile` (multi-stage Go build)
  - Create `frontend/Dockerfile.dev` (Node.js with Angular CLI)
  - Create `.env.example` with: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `OPENAI_API_KEY`
  - Verify: `docker compose config` passes validation

  **Must NOT do**:
  - No production Nginx config yet（Task 23）
  - No actual CopilotKit runtime code（Task 17）

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准Docker配置，无需特殊技能）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4, 5, 6)
  - **Blocks**: 21, 23, 24
  - **Blocked By**: None

  **Acceptance Criteria**:
  - [ ] `docker compose config` exits 0（YAML语法正确）
  - [ ] `docker compose up -d postgres` → PostgreSQL容器运行，healthcheck通过
  - [ ] `docker compose down` 清理所有容器

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Docker Compose validates and PostgreSQL starts
    Tool: Bash
    Preconditions: Docker daemon running, docker-compose.yml exists
    Steps:
      1. docker compose config → verify exit 0
      2. docker compose up -d postgres
      3. sleep 5 && docker compose ps postgres → verify status "healthy"
      4. docker compose down
    Expected Result: PostgreSQL container starts, healthcheck passes
    Failure Indicators: docker compose config fails, container crash, healthcheck timeout
    Evidence: .omo/evidence/task-3-docker-up.txt

  Scenario: Docker Compose fails on bad YAML
    Tool: Bash
    Preconditions: docker-compose.yml intentionally malformed (temporary)
    Steps:
      1. cp docker-compose.yml docker-compose.yml.bak
      2. echo "bad: : yaml" >> docker-compose.yml
      3. docker compose config → verify exit != 0
      4. mv docker-compose.yml.bak docker-compose.yml
    Expected Result: docker compose config fails with parse error
    Evidence: .omo/evidence/task-3-docker-error.txt
  ```

  **Commit**: NO（与Task 1, 2合并提交）

- [x] 4. PostgreSQL Schema + Migration System

  **What to do**:
  - Create `backend/migrations/` with golang-migrate format: `{timestamp}_{name}.up.sql` + `.down.sql`
  - Migration 1: `001_create_formulas.up.sql`
    - `formulas`: id (UUID PK), name, code (UNIQUE), component_mode (ENUM: single/double), status (ENUM: draft/active/archived), created_at, updated_at
  - Migration 2: `002_create_formula_parts.up.sql`
    - `formula_parts`: id (UUID PK), formula_id (FK), name (ENUM: A/B/MAIN), mix_ratio DECIMAL(5,2), sort_order INT
  - Migration 3: `003_create_formula_ingredients.up.sql`
    - `formula_ingredients`: id (UUID PK), part_id (FK), sort_order INT, material VARCHAR(200), percentage DECIMAL(5,4), weight DECIMAL(10,4)
  - Migration 4: `004_create_formula_steps.up.sql`
    - `formula_steps`: id (UUID PK), formula_id (FK), part_id (FK nullable), step_no INT, name VARCHAR(200), temperature VARCHAR(50), duration VARCHAR(50)
  - Migration 5: `005_create_formula_dosing_actions.up.sql`
    - `formula_dosing_actions`: id (UUID PK), step_id (FK), ingredient_id (FK), dosing_order INT, use_ratio DECIMAL(5,4), dosing_method VARCHAR(100)
  - Integration: Go code runs migrations on startup via `golang-migrate`
  - TDD: Write integration test that runs migrations up → verifies schema → runs down

  **Must NOT do**:
  - No Project/Sample/Detection/Indicator tables（不在此范围）
  - No RBAC tables（单用户）
  - No sync service related tables

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准SQL schema设计，无需特殊技能）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 5, 6)
  - **Blocks**: 7, 8, 9, 10, 11
  - **Blocked By**: 3（需要PostgreSQL运行）

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/database/...` → FAIL（migration测试先写）
  - [ ] `go run cmd/migrate/main.go up` → 5张表创建成功
  - [ ] `psql -h localhost -U formula -d formula_ai -c "\dt"` → 列出5张表
  - [ ] `go test ./internal/database/...` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Migrations create all tables with correct schema
    Tool: Bash (psql)
    Preconditions: PostgreSQL running, migrations applied
    Steps:
      1. psql -h localhost -U formula -d formula_ai -c "\dt"
      2. Verify 5 tables listed: formulas, formula_parts, formula_ingredients, formula_steps, formula_dosing_actions
      3. psql -d formula_ai -c "\d formulas" → verify columns: id(UUID), name, code, component_mode, status
    Expected Result: All 5 tables exist with correct columns
    Failure Indicators: Missing tables, wrong column types
    Evidence: .omo/evidence/task-4-schema.txt

  Scenario: Migration down removes all tables
    Tool: Bash (psql)
    Preconditions: All migrations applied (up)
    Steps:
      1. Run migration down
      2. psql -h localhost -U formula -d formula_ai -c "\dt"
      3. Verify no formula_* tables exist
    Expected Result: All formula tables dropped
    Evidence: .omo/evidence/task-4-migrate-down.txt
  ```

  **Commit**: YES
  - Message: `feat(db): postgresql schema with formula domain tables`
  - Files: `backend/migrations/`

- [x] 5. Go Domain Types + Validation Rules

  **What to do**:
  - Create `internal/models/` with Go structs mirroring DB schema + JSON tags
    - `Formula`: ID, Name, Code, ComponentMode(Single/Double), Status(Draft/Active/Archived), Parts[], Steps[], CreatedAt, UpdatedAt
    - `FormulaPart`: ID, FormulaID, Name(PartA/PartB/PartMain), MixRatio, SortOrder, Ingredients[]
    - `FormulaIngredient`: ID, PartID, SortOrder, Material, Percentage(float64), Weight(float64), DosingActions[]
    - `FormulaStep`: ID, FormulaID, PartID(*UUID nullable), StepNo, Name, Temperature, Duration
    - `FormulaDosingAction`: ID, StepID, IngredientID, DosingOrder, UseRatio(float64), DosingMethod
  - Create `internal/models/validation.go` with validation rules:
    - `ValidateFormula(f *Formula) error`: 检查Parts匹配ComponentMode，Steps非空（激活时），每Part至少1个Ingredient
    - `ValidatePartIngredients(part *FormulaPart) error`: 所有Ingredient.Percentage总和 ≈ 100%（容差0.01）
    - `ValidateDosingRatios(ingredient *FormulaIngredient) error`: 所有DosingAction.UseRatio总和 ≈ 100%
    - `ValidateIngredientHasDosingActions(ingredient *FormulaIngredient) error`: 每个Ingredient至少1个DosingAction
  - TDD: 每个验证函数先写测试（合法/非法用例），再实现

  **Must NOT do**:
  - No Project/Sample association（单项目，隐含）
  - No user/organization association（单用户）
  - No i18n（中文字段值硬编码在enum中即可）

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准Go struct定义+业务验证逻辑）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4, 6)
  - **Blocks**: 7, 8, 9, 10, 11, 13
  - **Blocked By**: None

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/models/...` → FAIL（验证函数尚未实现）
  - [ ] 所有验证测试用例覆盖：合法输入、非法%总和、非法投料比例、缺少DosingAction、无Step激活
  - [ ] `go test ./internal/models/...` → PASS（≥15个测试用例）

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Valid formula passes all validation rules
    Tool: Bash (go test)
    Preconditions: Test file with valid formula fixture
    Steps:
      1. cd backend && go test ./internal/models/ -run TestValidateFormula_Valid -v
      2. Verify test passes
    Expected Result: No validation errors for correctly structured formula
    Evidence: .omo/evidence/task-5-valid-formula.txt

  Scenario: Formula with ingredients not summing to 100% fails validation
    Tool: Bash (go test)
    Preconditions: Test fixture with part ingredients summing to 85%
    Steps:
      1. go test ./internal/models/ -run TestValidatePartIngredients_SumNot100 -v
      2. Verify test returns error containing "sum" or "100"
    Expected Result: Validation error: ingredient percentages do not sum to 100%
    Evidence: .omo/evidence/task-5-invalid-sum.txt
  ```

  **Commit**: YES
  - Message: `feat(backend): domain types and validation rules`
  - Files: `backend/internal/models/`

- [x] 6. Go Database Connection + Config Management

  **What to do**:
  - Create `internal/config/config.go`: 从环境变量读取配置（DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, SERVER_PORT）
  - Create `internal/database/postgres.go`: 
    - `NewPostgresDB(cfg Config) (*sql.DB, error)` 创建连接池
    - `RunMigrations(db *sql.DB, migrationsPath string) error` 运行golang-migrate
    - Connection pool settings: max open 25, max idle 5, max lifetime 5min
  - Integrate into `cmd/server/main.go`: load config → connect DB → run migrations → start server
  - TDD: Write integration test for DB connection + migration runner

  **Must NOT do**:
  - No ORM（使用原生 `database/sql` + `lib/pq`，或轻量 `sqlx`）
  - No hardcoded credentials（必须从环境变量读取）

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准Go DB连接配置）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 4, 5)
  - **Blocks**: 7, 8, 9, 10, 11
  - **Blocked By**: 3, 4（需要PostgreSQL运行和migration文件）

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/database/... -run TestNewPostgresDB` → FAIL
  - [ ] 配置缺失时启动失败（明确错误信息）
  - [ ] `go test ./internal/database/...` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Database connection succeeds with valid config
    Tool: Bash
    Preconditions: PostgreSQL running, .env configured correctly
    Steps:
      1. cd backend && DB_HOST=localhost DB_PORT=5432 DB_USER=formula DB_PASSWORD=test DB_NAME=formula_ai go run cmd/server/main.go
      2. Verify log output shows "Connected to PostgreSQL"
      3. Verify log output shows "Migrations applied successfully"
    Expected Result: Server starts, connects to DB, applies migrations
    Failure Indicators: Connection refused, auth failed, migration error
    Evidence: .omo/evidence/task-6-db-connect.txt

  Scenario: Missing config causes clear error
    Tool: Bash
    Preconditions: No DB env vars set
    Steps:
      1. cd backend && unset DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME
      2. go run cmd/server/main.go 2>&1
      3. Verify stderr contains "DB_HOST" or "database configuration"
    Expected Result: Clear error message about missing configuration
    Evidence: .omo/evidence/task-6-missing-config.txt
  ```

  **Commit**: YES
  - Message: `feat(backend): database connection and config management`
  - Files: `backend/internal/config/`, `backend/internal/database/`

- [x] 7. Formula CRUD Repository (PostgreSQL)

  **What to do**:
  - Create `internal/repository/formula_repo.go` with interface:
    - `Create(ctx, *Formula) error` — INSERT formula + nested parts/ingredients/steps/dosing
    - `GetByID(ctx, id UUID) (*Formula, error)` — JOIN all nested tables
    - `List(ctx) ([]*Formula, error)` — list all formulas
    - `Update(ctx, *Formula) error` — transactional update (delete+recreate nested)
    - `Delete(ctx, id UUID) error` — CASCADE delete
  - Create `internal/repository/part_repo.go`: CRUD for FormulaPart
  - Create `internal/repository/ingredient_repo.go`: CRUD for FormulaIngredient
  - Create `internal/repository/step_repo.go`: CRUD for FormulaStep
  - Create `internal/repository/dosing_repo.go`: CRUD for FormulaDosingAction
  - Use PostgreSQL transactions for atomic nested operations
  - TDD: 每个CRUD操作先写集成测试（需要PostgreSQL测试容器或test DB）

  **Must NOT do**:
  - No caching layer（简单单用户应用）
  - No soft delete（直接物理删除）
  - No audit logging

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []（核心数据层，需要仔细处理事务和嵌套关系）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 8, 9, 10, 11)
  - **Blocks**: 9, 10, 11, 18
  - **Blocked By**: 4, 5, 6

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/repository/...` → FAIL
  - [ ] Create formula with 2 parts, 3 ingredients each, 5 steps, dosing actions → 数据库验证所有嵌套数据存在
  - [ ] GetByID returns fully populated formula with all nested relations
  - [ ] Delete cascade: 删除formula → 验证所有关联parts/ingredients/steps/dosing被删除
  - [ ] `go test ./internal/repository/...` → PASS (≥10 tests)

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Create and retrieve a full formula with all nested data
    Tool: Bash (curl)
    Preconditions: PostgreSQL running, API server on port 8080
    Steps:
      1. curl -X POST http://localhost:8080/api/formulas -H 'Content-Type: application/json' -d '{"name":"Test","code":"T-001","component_mode":"single","parts":[{"name":"MAIN","mix_ratio":100,"ingredients":[{"material":"A","percentage":50},{"material":"B","percentage":50}]}],"steps":[{"step_no":1,"name":"Mix","temperature":"25C","duration":"10min"}]}'
      2. Verify response status = 201
      3. curl http://localhost:8080/api/formulas/{id-from-response}
      4. Verify response contains 1 part, 2 ingredients, 1 step
    Expected Result: Formula created and retrieved with all nested data intact
    Failure Indicators: Missing nested data, wrong counts, 404 on retrieval
    Evidence: .omo/evidence/task-7-crud-create.json

  Scenario: Delete formula cascades to all nested data
    Tool: Bash (curl + psql)
    Preconditions: Formula created via API
    Steps:
      1. curl -X DELETE http://localhost:8080/api/formulas/{id}
      2. Verify status = 204
      3. psql -d formula_ai -c "SELECT COUNT(*) FROM formula_parts WHERE formula_id='{id}'"
      4. Verify count = 0
    Expected Result: Formula and all nested data deleted
    Evidence: .omo/evidence/task-7-crud-delete.txt
  ```

  **Commit**: YES
  - Message: `feat(backend): formula crud repository with nested operations`
  - Files: `backend/internal/repository/`

- [x] 8. Formula Validation Service (Business Rules)

  **What to do**:
  - Create `internal/services/validator.go`:
    - `ValidateAndPrepare(f *Formula) error` — 所有检查的入口
    - `validatePartsMatchMode(f *Formula) error` — 单组分模式必须有且仅有一个MAIN part；双组分必须有A和B
    - `validatePartIngredients(part *FormulaPart) error` — 每Part配料%总和=100%（容差0.001）
    - `validateDosingCompleteness(f *Formula) error` — 每个Ingredient至少被1个DosingAction引用
    - `validateDosingRatios(ingredient *FormulaIngredient) error` — 每个Ingredient的投料比例总和=100%
    - `validateStepsForActivation(f *Formula) error` — 激活前至少1个Step
  - 区分"警告"（可保存但需修复）和"错误"（阻止保存）
  - TDD: 每个验证规则至少3个测试用例（合法、边界、非法）

  **Must NOT do**:
  - No cross-formula validation（单配方独立验证）
  - No material existence check（配方工作台范围，无物料管理模块）

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []（复杂业务规则，需要仔细处理边界条件）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 9, 10, 11)
  - **Blocks**: 9
  - **Blocked By**: 5

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/services/...` → FAIL
  - [ ] 单组分非法parts配置被拒绝
  - [ ] 配料%总和=95%被拒绝，错误信息包含实际值和期望值
  - [ ] 投料比例总和≠100%被拒绝
  - [ ] 激活无Step的配方被拒绝
  - [ ] `go test ./internal/services/...` → PASS (≥15 tests)

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Single-component mode rejects A+B parts configuration
    Tool: Bash (go test)
    Preconditions: Test function written
    Steps:
      1. go test ./internal/services/ -run TestValidate_SingleMode_RejectsDoubleParts -v
      2. Verify test passes
    Expected Result: Error: single-component mode requires MAIN part only
    Evidence: .omo/evidence/task-8-mode-mismatch.txt

  Scenario: Activating formula without steps is rejected
    Tool: Bash (go test)
    Preconditions: Formula with status=active, no steps
    Steps:
      1. go test ./internal/services/ -run TestValidate_Activation_RequiresSteps -v
      2. Verify error contains "step" or "步骤"
    Expected Result: Error: at least one step required for activation
    Evidence: .omo/evidence/task-8-no-steps.txt
  ```

  **Commit**: YES
  - Message: `feat(backend): formula validation service with business rules`
  - Files: `backend/internal/services/validator.go`

- [x] 9. Formula REST API Handlers (Gin Routes)

  **What to do**:
  - Create `internal/handlers/formula_handler.go`:
    - `POST /api/formulas` — `CreateFormula(c *gin.Context)`: 解析JSON → 验证 → 调用repo.Create → 返回201
    - `GET /api/formulas` — `ListFormulas(c *gin.Context)`: 调用repo.List → 返回200+JSON数组
    - `GET /api/formulas/:id` — `GetFormula(c *gin.Context)`: 解析UUID → 调用repo.GetByID → 返回200或404
    - `PUT /api/formulas/:id` — `UpdateFormula(c *gin.Context)`: 解析JSON → 验证 → 调用repo.Update → 返回200
    - `DELETE /api/formulas/:id` — `DeleteFormula(c *gin.Context)`: 解析UUID → 调用repo.Delete → 返回204
  - Create `internal/handlers/part_handler.go`, `ingredient_handler.go`, `step_handler.go`, `dosing_handler.go` with similar CRUD
  - Create `internal/handlers/routes.go`: 注册所有路由到Gin Engine
  - 中间件：CORS（允许localhost:4200）、JSON Content-Type、错误处理
  - TDD: 每个端点先写handler测试（使用httptest模拟请求）

  **Must NOT do**:
  - No auth middleware（单用户）
  - No pagination for List（单用户，数据量可控）
  - No rate limiting

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []（核心API层，需要处理请求校验、错误码、边界情况）

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 10, 11)
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 10, 11)
  - **Blocks**: 18, 24
  - **Blocked By**: 5, 6, 7, 8

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/handlers/...` → FAIL
  - [ ] `curl -X POST localhost:8080/api/formulas -d '{valid json}'` → 201
  - [ ] `curl localhost:8080/api/formulas/:id` → 200 + 完整formula JSON
  - [ ] `curl -X DELETE localhost:8080/api/formulas/:id` → 204
  - [ ] 非法JSON返回400 + 清晰错误信息
  - [ ] 不存在的id返回404
  - [ ] `go test ./internal/handlers/...` → PASS (≥12 tests)

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Create formula with valid data returns 201
    Tool: Bash (curl)
    Preconditions: Server running on port 8080
    Steps:
      1. curl -s -X POST http://localhost:8080/api/formulas -H 'Content-Type: application/json' -d '{"name":"Test Formula","code":"TF-001","component_mode":"single","parts":[{"name":"MAIN","mix_ratio":100,"ingredients":[{"material":"Material A","percentage":60},{"material":"Material B","percentage":40}]}],"steps":[{"step_no":1,"name":"Mixing","temperature":"25°C","duration":"30min"}]}'
      2. Verify HTTP status = 201
      3. Verify response has "id" field (jq '.id')
      4. Verify response "name" = "Test Formula" (jq '.name')
    Expected Result: 201 with created formula object including UUID
    Failure Indicators: Non-201 status, missing id, wrong name
    Evidence: .omo/evidence/task-9-create-201.json

  Scenario: Create formula with invalid data returns 400
    Tool: Bash (curl)
    Preconditions: Server running
    Steps:
      1. curl -s -X POST http://localhost:8080/api/formulas -H 'Content-Type: application/json' -d '{"name":"Bad","component_mode":"single","parts":[{"name":"MAIN","ingredients":[{"material":"X","percentage":50}]}]}' 
         (note: ingredients sum to 50%, not 100%)
      2. Verify HTTP status = 400
      3. Verify response body contains error message about sum
    Expected Result: 400 with validation error detail
    Evidence: .omo/evidence/task-9-create-400.json
  ```

  **Commit**: YES
  - Message: `feat(backend): formula rest api handlers with full crud`
  - Files: `backend/internal/handlers/`

- [x] 10. Formula Matrix Query API

  **What to do**:
  - Create `internal/handlers/matrix_handler.go`:
    - `GET /api/formulas/matrix` — `GetFormulaMatrix(c *gin.Context)`
  - 查询所有Formula + Parts + Ingredients + DosingActions，返回扁平化的矩阵数据结构：
    ```json
    {
      "formulas": [{
        "id", "name", "code", "component_mode",
        "parts": [{ "name", "mix_ratio", "ingredients": [{"material", "percentage", "weight"}] }],
        "steps": [{ "step_no", "name", "temperature", "duration", 
                    "dosing_actions": [{"ingredient_material", "dosing_order", "use_ratio", "dosing_method"}] }]
      }]
    }
    ```
  - Query参数：`?component_mode=single|double` 过滤组分类型
  - TDD: handler test + integration test（数据库中先seed测试数据）

  **Must NOT do**:
  - No project filter（单项目）
  - No detection data in matrix（无检测模块）
  - No pagination

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（查询格式化，逻辑简单）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 9, 11)
  - **Blocks**: 14
  - **Blocked By**: 5, 6, 7

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/handlers/... -run TestGetFormulaMatrix` → FAIL
  - [ ] `curl localhost:8080/api/formulas/matrix` → 200 + JSON数组（可为空数组）
  - [ ] Seed 2个配方后，matrix返回2个配方及其嵌套数据
  - [ ] `curl localhost:8080/api/formulas/matrix?component_mode=single` → 只返回单组分配方
  - [ ] `go test ./internal/handlers/... -run TestGetFormulaMatrix` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Matrix returns all formulas with nested structure
    Tool: Bash (curl)
    Preconditions: 2 formulas created via API
    Steps:
      1. curl -s http://localhost:8080/api/formulas/matrix | jq '.formulas | length'
      2. Verify result = 2
      3. curl -s http://localhost:8080/api/formulas/matrix | jq '.formulas[0].parts[0].ingredients'
      4. Verify ingredients array is not empty
    Expected Result: 2 formulas with nested parts and ingredients
    Failure Indicators: Empty array, missing nested data, wrong count
    Evidence: .omo/evidence/task-10-matrix.json
  ```

  **Commit**: YES
  - Message: `feat(backend): formula matrix query api`
  - Files: `backend/internal/handlers/matrix_handler.go`

- [x] 11. Analysis Data API Endpoints

  **What to do**:
  - Create `internal/handlers/analysis_handler.go`:
    - `GET /api/analysis/ingredient-distribution` — 返回所有配方的配料分布统计（material → count, avg_percentage）
    - `GET /api/analysis/component-mode-ratio` — 返回单组分vs双组分配方数量和占比
    - `GET /api/analysis/step-count-distribution` — 返回配方步骤数分布（1-2步、3-5步、5+步各有多少配方）
    - `GET /api/analysis/dosing-method-stats` — 返回投料方法使用频次统计
  - 使用SQL聚合查询（GROUP BY, COUNT, AVG）
  - TDD: 每个端点先写测试（seed数据 → 调用API → 验证统计结果）

  **Must NOT do**:
  - No hardcoded mock data（参考项目FormulaAnalysis.tsx的问题是数据硬编码）
  - No AI-generated analysis（分析=统计聚合，AI推荐由CopilotKit处理）

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（SQL聚合查询+JSON返回）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7, 8, 9, 10)
  - **Blocks**: 15
  - **Blocked By**: 5, 6, 7

  **Acceptance Criteria**:
  - [ ] TDD: `go test ./internal/handlers/... -run TestAnalysis` → FAIL
  - [ ] Seed 3个配方（2单组分+1双组分），验证 `component-mode-ratio` 返回正确统计
  - [ ] Seed配方使用不同配料，验证 `ingredient-distribution` 返回正确计数
  - [ ] `go test ./internal/handlers/... -run TestAnalysis` → PASS (≥4 tests)

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Ingredient distribution returns correct stats
    Tool: Bash (curl)
    Preconditions: Formulas created with known ingredient distribution
    Steps:
      1. Create formula with Material-A (60%) and Material-B (40%)
      2. Create formula with Material-A (30%) and Material-C (70%)
      3. curl -s http://localhost:8080/api/analysis/ingredient-distribution | jq .
      4. Verify Material-A count = 2, avg_percentage ≈ 45
    Expected Result: Correct aggregated stats for each material
    Failure Indicators: Wrong counts, wrong averages, empty response
    Evidence: .omo/evidence/task-11-analysis.json
  ```

  **Commit**: YES
  - Message: `feat(backend): analysis data api with aggregation queries`
  - Files: `backend/internal/handlers/analysis_handler.go`

- [x] 12. Angular Layout (Workspace Left + Chat Right Split)

  **What to do**:
  - 实现 `app.component` 的 split-pane 布局，使用 Angular Material：
    - 左侧 70%: `<mat-tab-group>` — 3个Tab（展示/分析/编辑），每个Tab渲染对应组件
    - 右侧 30%: CopilotKit聊天面板placeholder（Task 20完善），使用 `mat-card` 包裹
  - 响应式：桌面端70/30，移动端100%堆叠
  - 现代化风格：圆角卡片、柔和阴影、流畅过渡动画
  - 配色：深色侧边栏 + 浅色内容区，主色调 teal（`#0d9488`）
  - TDD: 写组件测试验证布局渲染和Tab切换

  **Must NOT do**:
  - No CopilotKit actual chat yet（Task 20）
  - No tab content implementation（Tasks 14-16）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-design`, `playwright`]
    - `frontend-design`: 专业配色的split布局设计
    - `playwright`: 浏览器端到端验证布局

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 13, 14, 15, 16)
  - **Blocks**: 14, 15, 16, 20
  - **Blocked By**: 2

  **Acceptance Criteria**:
  - [ ] TDD: `ng test --watch=false --include='**/app.component.spec.ts'` → FAIL
  - [ ] 页面渲染出左侧workspace区域和右侧chat区域
  - [ ] Tab切换：点击"配方分析"Tab → 分析区域可见，展示区域隐藏
  - [ ] `ng test --watch=false` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Split layout renders with correct proportions
    Tool: Playwright
    Preconditions: `ng serve` running on port 4200
    Steps:
      1. Navigate to http://localhost:4200
      2. Wait for .workspace-panel selector to appear
      3. Verify .chat-panel selector exists
      4. Check .workspace-panel width > .chat-panel width
      5. Screenshot full page
    Expected Result: Left panel occupies ~70%, right panel ~30%
    Failure Indicators: Missing panels, wrong proportions, layout broken
    Evidence: .omo/evidence/task-12-layout.png

  Scenario: Tab switching works
    Tool: Playwright
    Preconditions: App running with 3 tabs
    Steps:
      1. Navigate to http://localhost:4200
      2. Click tab "配方分析" (selector: .tab-analysis or text="配方分析")
      3. Verify analysis content area is visible (not display:none)
      4. Verify matrix content area is hidden
      5. Screenshot
    Expected Result: Active tab content visible, inactive hidden
    Evidence: .omo/evidence/task-12-tab-switch.png
  ```

  **Commit**: YES
  - Message: `feat(frontend): split layout with workspace and chat panels`
  - Files: `frontend/src/app/app.component.*`

- [x] 13. Angular API Service + Type Definitions

  **What to do**:
  - Create `frontend/src/app/services/formula-api.service.ts`:
    - `getFormulas()` → `GET /api/formulas`
    - `getFormula(id)` → `GET /api/formulas/:id`
    - `createFormula(data)` → `POST /api/formulas`
    - `updateFormula(id, data)` → `PUT /api/formulas/:id`
    - `deleteFormula(id)` → `DELETE /api/formulas/:id`
    - `getFormulaMatrix()` → `GET /api/formulas/matrix`
    - `getAnalysis(endpoint)` → `GET /api/analysis/:endpoint`
  - Create `frontend/src/app/types/formula.types.ts`:
    - TypeScript interfaces matching Go models: `Formula`, `FormulaPart`, `FormulaIngredient`, `FormulaStep`, `FormulaDosingAction`, `FormulaMatrix`
  - 使用Angular的 `HttpClient` + `Observable`
  - 配置proxy: `proxy.conf.json` 将 `/api` 转发到 `http://localhost:8080`
  - TDD: HTTP mocking测试验证每个方法返回正确的Observable

  **Must NOT do**:
  - No state management library (NgRx/Akita) — 使用Service + BehaviorSubject足够
  - No caching logic yet（简单单用户场景）

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准Angular HttpClient封装）

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 14, 15, 16)
  - **Blocks**: 14, 15, 16, 20
  - **Blocked By**: 5

  **Acceptance Criteria**:
  - [ ] TDD: `ng test --watch=false --include='**/formula-api.service.spec.ts'` → FAIL
  - [ ] 所有API方法正确定义（编译通过）
  - [ ] HTTP mock测试：`getFormulas()` 返回mock数据
  - [ ] HTTP mock测试：`createFormula()` 发送POST请求并处理响应
  - [ ] `ng test --watch=false` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: API service correctly calls backend endpoint
    Tool: Bash (curl via proxy)
    Preconditions: Go API running on 8080, Angular dev server on 4200 with proxy
    Steps:
      1. Create a test formula via curl to Go API
      2. curl -s http://localhost:4200/api/formulas | jq '. | length'
      3. Verify response contains the created formula
    Expected Result: Angular proxy forwards to Go API, returns data
    Failure Indicators: 404, proxy not working, CORS error
    Evidence: .omo/evidence/task-13-api-proxy.json
  ```

  **Commit**: YES
  - Message: `feat(frontend): api service and type definitions`
  - Files: `frontend/src/app/services/`, `frontend/src/app/types/`

- [x] 14. Formula Matrix Tab Component

  **What to do**:
  - Create `FormulaMatrixComponent`:
    - 使用 `mat-table` 渲染交叉表，支持展开行（`mat-row` + expand animation）
    - 调用 `formulaApiService.getFormulaMatrix()` 获取数据
    - 双组分模式：A/B parts用 `mat-divider` 和颜色标签区分
    - 顶部工具栏：`mat-toolbar` + `mat-button-toggle-group` 过滤组分模式
    - Loading/Empty/Error 三种状态使用 `mat-spinner` / `mat-card` 空状态
  - TDD: 写组件测试验证数据渲染、loading状态、空状态

  **Must NOT do**:
  - No sample/detection columns（参考项目有，单机版无此数据）
  - No project selector（单项目）
  - No inline editing（编辑功能在Editor Tab，Task 16）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-design`, `playwright`]
    - `frontend-design`: 专业表格样式设计
    - `playwright`: 浏览器端渲染验证

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13, 15, 16)
  - **Blocks**: None
  - **Blocked By**: 10, 12, 13

  **Acceptance Criteria**:
  - [ ] TDD: `ng test --watch=false --include='**/formula-matrix.component.spec.ts'` → FAIL
  - [ ] 无数据时显示空状态提示
  - [ ] 有数据时渲染交叉表，展开后显示配料和投料信息
  - [ ] 双组分配方A/B parts正确分区显示
  - [ ] `ng test --watch=false` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Matrix displays formula data in cross-tabulation
    Tool: Playwright
    Preconditions: 2 formulas created via API
    Steps:
      1. Navigate to http://localhost:4200
      2. Click "配方展示" tab
      3. Wait for table rows to render (selector: '.formula-row')
      4. Verify at least 2 formula rows visible
      5. Click expand on first formula
      6. Verify ingredients row appears (selector: '.ingredient-row')
      7. Screenshot
    Expected Result: Cross-tabulation table with expandable formula rows
    Failure Indicators: Empty table, no rows, expand not working
    Evidence: .omo/evidence/task-14-matrix.png

  Scenario: Empty state when no formulas exist
    Tool: Playwright
    Preconditions: Database empty (fresh migration)
    Steps:
      1. Navigate to http://localhost:4200
      2. Click "配方展示" tab
      3. Verify empty state message visible (e.g., "暂无配方数据")
      4. Screenshot
    Expected Result: Clear empty state message, no table
    Evidence: .omo/evidence/task-14-empty-state.png
  ```

  **Commit**: YES
  - Message: `feat(frontend): formula matrix tab with cross-tabulation`
  - Files: `frontend/src/app/components/formula-matrix/`

- [x] 15. Formula Analysis Tab Component

  **What to do**:
  - Create `FormulaAnalysisComponent`:
    - 调用 `formulaApiService.getAnalysis(endpoint)` 获取4个统计端点数据
    - 使用 `ng2-charts` 或 `chart.js` (Angular wrapper) 渲染图表：
      - 饼图：单组分 vs 双组分配方占比
      - 柱状图：配料使用频次 Top 10
      - 柱状图：步骤数分布
      - 表格：投料方法统计
    - Loading/Empty/Error 三种状态
  - Install: `npm install ng2-charts chart.js`
  - TDD: 写组件测试验证图表渲染（不验证canvas像素，验证data binding）

  **Must NOT do**:
  - No hardcoded chart data（参考项目FormulaAnalysis.tsx的问题，必须从API获取真实数据）
  - No AI score card or improvement roadmap（这些由LLM聊天窗口生成）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-design`, `playwright`]
    - `frontend-design`: 专业仪表盘设计
    - `playwright`: 浏览器端渲染验证+截图

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13, 14, 16)
  - **Blocks**: None
  - **Blocked By**: 11, 12, 13

  **Acceptance Criteria**:
  - [ ] TDD: `ng test --watch=false --include='**/formula-analysis.component.spec.ts'` → FAIL
  - [ ] 4张图表全部渲染（饼图、柱状图×2、表格）
  - [ ] 无数据时每张图表显示空状态
  - [ ] 图表数据来自API（非硬编码）
  - [ ] `ng test --watch=false` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Analysis dashboard renders all charts with real data
    Tool: Playwright
    Preconditions: 3+ formulas with varied data created
    Steps:
      1. Navigate to http://localhost:4200
      2. Click "配方分析" tab
      3. Wait for charts to render (selector: 'canvas')
      4. Verify at least 3 canvas elements visible
      5. Screenshot full analysis tab
    Expected Result: All 4 chart types rendered with data
    Failure Indicators: Missing canvases, blank charts, errors in console
    Evidence: .omo/evidence/task-15-analysis.png

  Scenario: Analysis shows empty state when no data
    Tool: Playwright
    Preconditions: Empty database
    Steps:
      1. Navigate to http://localhost:4200
      2. Click "配方分析" tab
      3. Verify empty state message visible
      4. Screenshot
    Expected Result: Informative empty state, no broken charts
    Evidence: .omo/evidence/task-15-empty.png
  ```

  **Commit**: YES
  - Message: `feat(frontend): formula analysis tab with charts`
  - Files: `frontend/src/app/components/formula-analysis/`

- [x] 16. Formula Editor Tab Component

  **What to do**:
  - Create `FormulaEditorComponent`，使用 Angular Material 表单：
    - 所有输入使用 `mat-form-field` + `matInput`，选择使用 `mat-select`
    - 基础信息区：`mat-card` 包裹 name, code, component_mode (mat-select), status
    - Parts区域（动态FormArray）：`mat-expansion-panel` 每个Part可折叠，`mat-icon-button` 增删
    - Ingredients（动态FormArray）：`mat-chip-list` 或 `mat-list` 展示配料列表
    - Steps区域：`mat-stepper` 或 `mat-accordion` 展示工艺步骤
    - DosingActions：`mat-dialog` 或 inline `mat-select` 关联Step和Ingredient
    - 实时验证反馈：`mat-hint` + `mat-error`（绿色/红色），%总和指示器用 `mat-progress-bar`
    - 提交按钮：`mat-raised-button color="primary"`
  - TDD: 表单验证逻辑、动态FormArray增删、提交行为测试

  **Must NOT do**:
  - No material autocomplete（无物料管理模块）
  - No batch import/export
  - No version history

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`frontend-design`, `playwright`]
    - `frontend-design`: 复杂表单UI设计
    - `playwright`: 表单填写、提交、验证流程测试

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 12, 13, 14, 15)
  - **Blocks**: None
  - **Blocked By**: 9, 12, 13

  **Acceptance Criteria**:
  - [ ] TDD: `ng test --watch=false --include='**/formula-editor.component.spec.ts'` → FAIL
  - [ ] 新建模式：填写所有字段 → 提交 → API调用成功 → 页面跳转到展示Tab
  - [ ] 编辑模式：加载已有配方数据 → 修改 → 保存 → 数据更新
  - [ ] 实时验证：配料%总和≠100%时显示红色提示
  - [ ] 动态增删Parts/Ingredients/Steps/DosingActions正常工作
  - [ ] `ng test --watch=false` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Create new formula with full nested data
    Tool: Playwright
    Preconditions: App running, on editor tab
    Steps:
      1. Navigate to http://localhost:4200
      2. Click "新增" tab
      3. Fill name field: "New Formula" (selector: input[name="name"])
      4. Fill code field: "NF-001" (selector: input[name="code"])
      5. Select component_mode: "single" (selector: select[name="component_mode"])
      6. In MAIN part, click "添加配料"
      7. Fill ingredient 1: material="Resin", percentage=60
      8. Click "添加配料" again, fill ingredient 2: material="Hardener", percentage=40
      9. Verify percentage sum indicator shows 100% (green)
      10. Click "添加步骤", fill step: step_no=1, name="Mix", temperature="25°C", duration="10min"
      11. Click "提交" button
      12. Verify redirected to matrix tab, new formula visible
      13. Screenshot before and after submit
    Expected Result: Formula created successfully with all nested data
    Failure Indicators: Validation errors on valid data, submit fails, redirect broken
    Evidence: .omo/evidence/task-16-create-formula.png

  Scenario: Editor shows validation error for invalid percentages
    Tool: Playwright
    Preconditions: On editor tab
    Steps:
      1. Fill ingredients: Resin=50%, Hardener=30% (sum=80%)
      2. Verify percentage indicator turns red and shows "80% (需要100%)"
      3. Click "提交"
      4. Verify error message appears (form not submitted)
      5. Screenshot
    Expected Result: Client-side validation blocks submission with clear error
    Evidence: .omo/evidence/task-16-validation-error.png
  ```

  **Commit**: YES
  - Message: `feat(frontend): formula editor tab with dynamic form`
  - Files: `frontend/src/app/components/formula-editor/`

- [x] 17. CopilotKit Runtime Setup (Node.js)

  **What to do**:
  - Create `copilotkit-runtime/` directory with Node.js project:
    - `package.json`: dependencies `@copilotkit/runtime`, `openai`, `express`, `cors`
    - `tsconfig.json`: TypeScript配置
  - Create `src/index.ts`:
    - Express server on port 3001
    - CopilotKit runtime endpoint `POST /api/copilotkit`
    - 从环境变量读取 `OPENAI_API_KEY`, `GO_API_URL`
  - Create `Dockerfile` for production
  - Integrate into `docker-compose.yml` service
  - TDD: Write integration test (start runtime → send message → verify response)

  **Must NOT do**:
  - No tool definitions yet（Task 18）
  - No custom UI components（CopilotKit Angular handles chat UI in Task 20）

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`claude-api`]
    - `claude-api`: OpenAI SDK integration patterns相似

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4 (sequential with Task 18; parallel with Task 20 after Task 18)
  - **Blocks**: 18, 19, 20, 24
  - **Blocked By**: 1（需要项目结构建立）

  **Acceptance Criteria**:
  - [ ] TDD: `cd copilotkit-runtime && npm test` → FAIL
  - [ ] `npm run dev` → 服务启动在3001端口
  - [ ] `curl -X POST http://localhost:3001/api/copilotkit -H 'Content-Type: application/json' -d '{"messages":[{"role":"user","content":"hello"}]}'` → 返回流式响应或JSON
  - [ ] `npm test` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: CopilotKit runtime responds to simple message
    Tool: Bash (curl)
    Preconditions: Runtime running on port 3001, OPENAI_API_KEY set
    Steps:
      1. curl -s -X POST http://localhost:3001/api/copilotkit -H 'Content-Type: application/json' -d '{"messages":[{"role":"user","content":"Say hello"}]}'
      2. Verify response contains streamed text or JSON response
      3. Verify response includes assistant message
    Expected Result: Runtime calls OpenAI and returns response
    Failure Indicators: Connection refused, empty response, auth error
    Evidence: .omo/evidence/task-17-runtime-response.json

  Scenario: Runtime handles missing API key gracefully
    Tool: Bash
    Preconditions: OPENAI_API_KEY not set
    Steps:
      1. unset OPENAI_API_KEY && npm run dev 2>&1
      2. Verify error message about missing API key
    Expected Result: Clear error, server fails to start
    Evidence: .omo/evidence/task-17-missing-key.txt
  ```

  **Commit**: YES
  - Message: `feat(ai): copilotkit runtime setup with express`
  - Files: `copilotkit-runtime/`

- [x] 18. MCP Tool Definitions (Formula CRUD Actions)

  **What to do**:
  - Create `copilotkit-runtime/src/tools/` with tool definitions:
    - `list_formulas`: 查询所有配方 → `GET {GO_API_URL}/api/formulas`
    - `get_formula`: 查询单个配方 → `GET {GO_API_URL}/api/formulas/:id`
    - `create_formula`: 创建配方 → `POST {GO_API_URL}/api/formulas`，参数：name, code, component_mode, parts, steps
    - `update_formula`: 修改配方 → `PUT {GO_API_URL}/api/formulas/:id`
    - `delete_formula`: 删除配方 → `DELETE {GO_API_URL}/api/formulas/:id`
    - `get_analysis`: 获取统计数据 → `GET {GO_API_URL}/api/analysis/:endpoint`
  - 每个工具定义包含：
    - `name`, `description`（中文描述，帮助LLM理解何时调用）
    - `parameters`（JSON Schema定义参数类型和必填项）
    - `handler`（调用Go API + 格式化返回结果）
  - 注册到CopilotKit runtime的actions
  - TDD: Mock Go API，测试每个工具handler的请求/响应

  **Must NOT do**:
  - No business logic in tools（只做API代理，验证在Go后端）
  - No tool for operations outside formula CRUD scope

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []（工具定义+API代理，需要精确的JSON Schema和错误处理）

  **Parallelization**:
  - **Can Run In Parallel**: NO（在Task 17 runtime搭建后才能测试）
  - **Parallel Group**: Wave 4 (with Task 19)
  - **Blocks**: 20
  - **Blocked By**: 9, 10, 11, 17

  **Acceptance Criteria**:
  - [ ] TDD: `cd copilotkit-runtime && npm test -- --testPathPattern=tools` → FAIL
  - [ ] 每个工具定义有完整的JSON Schema
  - [ ] LLM调用 `create_formula` 工具 → Go API收到正确的POST请求 → 返回创建的formula
  - [ ] LLM调用 `list_formulas` → 返回配方列表
  - [ ] 工具错误时返回友好的错误描述给LLM
  - [ ] `npm test` → PASS (≥6 tests)

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: create_formula tool correctly proxies to Go API
    Tool: Bash (curl to CopilotKit runtime)
    Preconditions: Go API and CopilotKit runtime running
    Steps:
      1. Send message to runtime: "请创建一个名为Test的双组分硅胶配方，code为T-001"
      2. Verify CopilotKit runtime calls create_formula tool
      3. Verify Go API received POST /api/formulas with correct data
      4. curl http://localhost:8080/api/formulas → verify new formula exists
    Expected Result: LLM uses create_formula tool, formula created in database
    Failure Indicators: Tool not called, wrong API endpoint, formula not created
    Evidence: .omo/evidence/task-18-create-tool.json

  Scenario: Tool gracefully handles Go API error
    Tool: Bash
    Preconditions: Go API running, tool definition in place
    Steps:
      1. Send message requesting formula with intentionally invalid data
      2. Verify tool returns error message (not crash)
      3. Verify runtime continues responding
    Expected Result: Error message returned to LLM, runtime doesn't crash
    Evidence: .omo/evidence/task-18-tool-error.txt
  ```

  **Commit**: YES
  - Message: `feat(ai): mcp tool definitions for formula crud`
  - Files: `copilotkit-runtime/src/tools/`

- [x] 19. OpenAI Integration (API Client + Streaming)

  **What to do**:
  - Create `copilotkit-runtime/src/openai/client.ts`:
    - 封装OpenAI SDK调用（使用 `openai` npm包）
    - 支持 function calling（tool use）
    - 支持流式响应（SSE streaming）
    - 错误处理：rate limit重试、token超限截断
  - 系统提示词（System Prompt）：
    - 中文："你是一个配方研发助手，可以帮助用户查询、创建、修改和删除化学配方。配方包含组分(A/B/MAIN)、配料、工艺步骤和投料动作。"
    - 列出可用工具及其用途
    - 约束：操作前确认，敏感删除需二次确认
  - TDD: Mock OpenAI API，测试client调用和流式响应

  **Must NOT do**:
  - No hardcoded API key（从环境变量读取）
  - No prompt injection防护（单用户可信任环境）
  - No conversation history persistance（单会话）

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`claude-api`]
    - `claude-api`: OpenAI API patterns + streaming experience相似

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Task 18)
  - **Blocks**: 20
  - **Blocked By**: 17

  **Acceptance Criteria**:
  - [ ] TDD: `npm test -- --testPathPattern=openai` → FAIL
  - [ ] 发送消息 → OpenAI返回响应（含文本）
  - [ ] 发送消息触发工具调用 → OpenAI返回function_call → 工具执行 → 返回结果
  - [ ] 流式响应正确传递SSE事件
  - [ ] API key缺失时给出明确错误
  - [ ] `npm test` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: LLM responds to simple query without tools
    Tool: Bash (curl to CopilotKit runtime)
    Preconditions: Runtime running, OpenAI key valid
    Steps:
      1. curl -s -X POST http://localhost:3001/api/copilotkit -d '{"messages":[{"role":"user","content":"你好，请介绍一下你自己"}]}'
      2. Verify response contains Chinese assistant message
    Expected Result: LLM responds in Chinese introducing itself
    Evidence: .omo/evidence/task-19-chat-response.json

  Scenario: LLM uses function calling for formula query
    Tool: Bash (curl)
    Preconditions: Runtime running, tools defined, formulas exist in DB
    Steps:
      1. Send: "列出所有配方"
      2. Verify response includes function_call to list_formulas
      3. Verify final response lists formulas from DB
    Expected Result: LLM calls list_formulas tool, returns actual data
    Evidence: .omo/evidence/task-19-function-call.json
  ```

  **Commit**: YES
  - Message: `feat(ai): openai integration with function calling and streaming`
  - Files: `copilotkit-runtime/src/openai/`

- [x] 20. Angular CopilotKit Chat Panel

  **What to do**:
  - Create `ChatPanelComponent`，使用 Angular Material 卡片风格：
    - `mat-card` 外层容器，`mat-card-header` 标题"AI 配方助手"
    - 消息列表：`mat-list` + 用户消息右对齐（`mat-list-item` + 自定义CSS），AI消息左对齐
    - 工具调用展示：`mat-chip` 显示调用的工具名，`mat-expansion-panel` 展开查看详情
    - 输入区：`mat-form-field` + `matInput` + `mat-icon-button` 发送按钮
    - Loading状态：`mat-progress-bar` 顶部或 `mat-spinner` 内嵌
  - 使用 `@copilotkit/angular` 的 `CopilotKit` 服务
  - 配置 `provideCopilotKit({ runtimeUrl: '/api/copilotkit' })`（proxy转发到3001）
  - TDD: 组件渲染测试 + 消息发送测试

  **Must NOT do**:
  - No custom chat bubble design（使用CopilotKit默认样式或简单Tailwind）
  - No conversation history persistence across page refresh

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`frontend-design`, `playwright`]
    - `frontend-design`: 专业聊天UI设计
    - `playwright`: 聊天交互端到端测试

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖Task 19后才能测试完整流程）
  - **Parallel Group**: Wave 4 (after Tasks 17, 18, 19)
  - **Blocks**: 24
  - **Blocked By**: 2, 12, 17, 18, 19

  **Acceptance Criteria**:
  - [ ] TDD: `ng test --watch=false --include='**/chat-panel.component.spec.ts'` → FAIL
  - [ ] 发送"列出所有配方" → LLM响应并列出表格中的配方
  - [ ] 发送"创建一个单组分配方，名为XXX" → LLM调用create_formula工具 → 配方出现在展示Tab
  - [ ] 发送非法请求 → LLM友好拒绝（不崩溃）
  - [ ] `ng test --watch=false` → PASS

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Chat panel sends message and receives LLM response
    Tool: Playwright
    Preconditions: Go API, CopilotKit runtime running, Angular on port 4200
    Steps:
      1. Navigate to http://localhost:4200
      2. Locate chat input (selector: '.chat-input' or 'textarea')
      3. Type: "你好，请列出当前所有的配方"
      4. Press Enter
      5. Wait for response (selector: '.assistant-message', timeout: 30s)
      6. Verify response contains formula list or "暂无配方"
      7. Screenshot chat panel
    Expected Result: LLM responds with formula data from API
    Failure Indicators: No response, error message, timeout
    Evidence: .omo/evidence/task-20-chat-response.png

  Scenario: Chat creates formula via tool call
    Tool: Playwright
    Preconditions: All services running
    Steps:
      1. Navigate to http://localhost:4200
      2. In chat input: "创建一个单组分配方，name为AI测试配方，code为AI-001，包含MAIN part，两种配料各50%，一个混合步骤"
      3. Press Enter
      4. Wait for tool call indicator (selector: '.tool-call')
      5. Wait for success response
      6. Switch to "配方展示" tab
      7. Verify "AI测试配方" appears in matrix table
      8. Screenshot both chat and matrix
    Expected Result: LLM creates formula via tool, visible in matrix tab
    Evidence: .omo/evidence/task-20-create-via-chat.png
  ```

  **Commit**: YES
  - Message: `feat(frontend): copilotkit chat panel with tool call display`
  - Files: `frontend/src/app/components/chat-panel/`

- [x] 21. Docker Compose Production Config

  **What to do**:
  - 完善 `docker-compose.yml` 生产配置：
    - `postgres`: PostgreSQL 16 Alpine, volume `pgdata:/var/lib/postgresql/data`, healthcheck `pg_isready`
    - `backend`: 多阶段Go构建（build stage → alpine run stage），环境变量注入，healthcheck `curl /api/health`
    - `frontend`: Angular production build → Nginx Alpine，Nginx proxy `/api` → backend:8080，`/api/copilotkit` → copilotkit-runtime:3001
    - `copilotkit-runtime`: Node.js Alpine，环境变量 `OPENAI_API_KEY`, `GO_API_URL=http://backend:8080`
  - Create `frontend/nginx.conf`: SPA路由重写 + API proxy
  - Create `.env` file with defaults
  - 网络配置：`formula-network` bridge network
  - 验证：`docker compose build` + `docker compose up -d` → 所有服务healthy

  **Must NOT do**:
  - No HTTPS/TLS（单机本地部署）
  - No Docker Swarm/K8s配置

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准Docker Compose/nginx配置）

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 (sequential integration)
  - **Blocks**: 24
  - **Blocked By**: 3, 9, 17, 20

  **Acceptance Criteria**:
  - [ ] `docker compose build` 无错误
  - [ ] `docker compose up -d` → 4个服务全部running+healthy
  - [ ] `curl http://localhost:8080/api/health` → `{"status":"ok"}`
  - [ ] `curl http://localhost:3001/api/copilotkit` → 200
  - [ ] 浏览器 `http://localhost` → Angular SPA正常加载

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: One-command deployment starts all services
    Tool: Bash
    Preconditions: Docker daemon running, .env configured
    Steps:
      1. docker compose up -d
      2. sleep 10 && docker compose ps
      3. Verify all 4 services status = "running" or "healthy"
      4. docker compose logs backend | grep "Connected to PostgreSQL"
      5. docker compose logs copilotkit-runtime | grep "listening"
    Expected Result: All 4 services start and pass healthchecks
    Failure Indicators: Container crashes, healthcheck failures, connection errors
    Evidence: .omo/evidence/task-21-docker-ps.txt

  Scenario: Clean shutdown removes all containers
    Tool: Bash
    Preconditions: All services running
    Steps:
      1. docker compose down
      2. docker compose ps → verify no running containers
      3. docker volume ls → verify data volume still exists
    Expected Result: Containers removed, data volume preserved
    Evidence: .omo/evidence/task-21-docker-down.txt
  ```

  **Commit**: YES
  - Message: `feat(infra): docker compose production config with nginx`
  - Files: `docker-compose.yml`, `frontend/nginx.conf`, `.env`

- [x] 22. Seed Data Script

  **What to do**:
  - Create `backend/cmd/seed/main.go`:
    - 读取 `seed/data.json` 文件
    - 创建3个示例配方：
      1. 单组分硅酮密封胶：MAIN part，5种配料（基胶/交联剂/催化剂/填料/助剂），3个步骤
      2. 双组分环氧树脂：A part（环氧树脂+稀释剂），B part（固化剂+促进剂），各2个步骤
      3. 单组分水性涂料：MAIN part，4种配料，2个步骤
    - 验证规则通过后再写入数据库
  - Seed data包含合理的百分比和投料比例（确保通过验证）
  - 幂等性：检查 `code` 是否已存在，避免重复seed
  - 集成到 Docker Compose：backend启动后自动执行（环境变量 `SEED_ON_START=true`）

  **Must NOT do**:
  - No fixture data from reference project（参考项目seed的是组织架构/RBAC数据，不需要）
  - No real/confidential data

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（数据生成+写入，逻辑简单）

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 (sequential)
  - **Blocks**: None
  - **Blocked By**: 9

  **Acceptance Criteria**:
  - [ ] `go run cmd/seed/main.go` → 3个配方创建成功
  - [ ] `curl http://localhost:8080/api/formulas` → 返回3个配方
  - [ ] Matrix API返回包含所有seed数据的完整嵌套结构
  - [ ] 幂等：二次运行不重复创建（code UNIQUE约束处理）

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Seed creates 3 formulas with valid nested data
    Tool: Bash (curl)
    Preconditions: PostgreSQL running, migrations applied
    Steps:
      1. go run cmd/seed/main.go
      2. curl -s http://localhost:8080/api/formulas | jq '. | length'
      3. Verify count = 3
      4. curl -s http://localhost:8080/api/formulas/matrix | jq '.formulas | length'
      5. Verify count = 3 with nested parts/ingredients
    Expected Result: 3 valid formulas with complete nested data
    Failure Indicators: Wrong count, missing nested data, validation errors
    Evidence: .omo/evidence/task-22-seed.json

  Scenario: Seed is idempotent (no duplicates)
    Tool: Bash
    Preconditions: Seed already run once
    Steps:
      1. go run cmd/seed/main.go (second run)
      2. curl -s http://localhost:8080/api/formulas | jq '. | length'
      3. Verify count still = 3
    Expected Result: No duplicate formulas created
    Evidence: .omo/evidence/task-22-idempotent.txt
  ```

  **Commit**: YES
  - Message: `feat(db): seed data script with 3 sample formulas`
  - Files: `backend/cmd/seed/`, `backend/seed/`

- [x] 23. Angular Production Build + Nginx Config

  **What to do**:
  - 创建生产环境配置 `frontend/src/environments/environment.prod.ts`
  - 配置 `frontend/nginx.conf`:
    - SPA fallback: `try_files $uri $uri/ /index.html`
    - API proxy: `location /api/` → `proxy_pass http://backend:8080`
    - CopilotKit proxy: `location /api/copilotkit` → `proxy_pass http://copilotkit-runtime:3001`
    - Gzip压缩，静态资源缓存（1年）
  - 创建 `frontend/Dockerfile`（多阶段：build stage + Nginx stage）
  - 验证：`docker build -f frontend/Dockerfile -t formula-frontend .` → image构建成功
  - 验证：Nginx配置语法 `nginx -t`

  **Must NOT do**:
  - No SSR（Angular Universal）
  - No CDN配置

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []（标准Angular生产构建+nginx配置）

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 (sequential)
  - **Blocks**: None
  - **Blocked By**: 2, 12, 20

  **Acceptance Criteria**:
  - [ ] `ng build --configuration production` 成功（无错误无警告）
  - [ ] `dist/` 目录生成，包含index.html和bundled JS/CSS
  - [ ] Nginx配置语法验证通过
  - [ ] Docker image构建成功且大小合理（<200MB）

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Production build serves Angular app via Nginx
    Tool: Playwright
    Preconditions: Docker Compose up with production config
    Steps:
      1. Navigate to http://localhost
      2. Verify page loads (no blank page, no 404)
      3. Verify split layout renders correctly
      4. Verify API calls work (matrix tab loads data)
      5. Screenshot
    Expected Result: Full Angular app served via Nginx with working API proxy
    Failure Indicators: Blank page, 404, API calls fail, CORS errors
    Evidence: .omo/evidence/task-23-prod-serve.png
  ```

  **Commit**: YES
  - Message: `feat(frontend): production build and nginx configuration`
  - Files: `frontend/nginx.conf`, `frontend/Dockerfile`

- [x] 24. End-to-End Integration Verification

  **What to do**:
  - 端到端验证完整用户流程：
    1. `docker compose up -d` 启动所有服务
    2. 验证种子数据加载（3个配方）
    3. 通过前端创建第4个配方（编辑器Tab）
    4. 验证矩阵Tab显示4个配方
    5. 验证分析Tab图表更新
    6. 通过CopilotKit聊天查询配方
    7. 通过CopilotKit聊天修改配方
    8. `docker compose down` 清理
  - 编写集成测试脚本 `scripts/e2e-test.sh`
  - 记录结果和截图到 `.omo/evidence/e2e/`

  **Must NOT do**:
  - No performance/load testing
  - No security scanning

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`playwright`]
    - `playwright`: 端到端浏览器流程自动化

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 (final integration)
  - **Blocks**: None
  - **Blocked By**: 21, 22, 23

  **Acceptance Criteria**:
  - [ ] 完整用户流程通过（启动→浏览→创建→聊天操作→关闭）
  - [ ] 所有API端点正常响应
  - [ ] CopilotKit工具调用成功
  - [ ] 数据持久化（重启后数据仍在）

  **QA Scenarios** (MANDATORY):
  ```
  Scenario: Full user journey from deploy to LLM interaction
    Tool: Bash + Playwright
    Preconditions: Clean Docker environment
    Steps:
      1. docker compose up -d && sleep 15
      2. curl http://localhost:8080/api/health → 200
      3. curl http://localhost:8080/api/formulas | jq '. | length' → 3 (seed)
      4. Playwright: Navigate to http://localhost, verify matrix shows 3 formulas
      5. Playwright: Switch to editor tab, create 4th formula
      6. Playwright: Switch to matrix tab, verify 4 formulas
      7. Playwright: In chat, type "显示所有配方" → verify response lists 4 formulas
      8. Playwright: In chat, type "删除名为XXX的配方" → verify formula removed
      9. docker compose down
    Expected Result: Complete user workflow succeeds end-to-end
    Failure Indicators: Any step fails
    Evidence: .omo/evidence/e2e/
  ```

  **Commit**: YES
  - Message: `feat(integration): end-to-end verification script`
  - Files: `scripts/e2e-test.sh`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists. For each "Must NOT Have": search codebase for forbidden patterns. Check evidence files exist in `.omo/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `golangci-lint` + `go test ./...` + `ng test` + `ng lint`. Review all changed files for: `any` types, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` skill)
  Start from clean state (`docker compose up`). Execute EVERY QA scenario from EVERY task. Test cross-task integration. Test edge cases: empty state, invalid input, rapid actions. Save to `.omo/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read spec vs actual diff. Verify 1:1. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **1**: `feat(backend): project scaffold` - go.mod, main.go, config/
- **2**: `feat(frontend): angular project init` - angular.json, package.json, app/
- **3**: `feat(infra): docker compose skeleton` - docker-compose.yml
- **4**: `feat(db): postgresql schema and migrations` - migrations/
- **5**: `feat(backend): domain types and validation` - models/, validator/
- **6**: `feat(backend): database connection and config` - database/, config.go
- **7**: `feat(backend): formula crud repository` - repository/
- **8**: `feat(backend): formula validation service` - services/validator.go
- **9**: `feat(backend): formula rest api handlers` - handlers/, routes.go
- **10**: `feat(backend): formula matrix query api` - handlers/matrix.go
- **11**: `feat(backend): analysis data api` - handlers/analysis.go
- **12**: `feat(frontend): layout split workspace and chat` - app.component
- **13**: `feat(frontend): api service and type definitions` - services/, types/
- **14**: `feat(frontend): formula matrix tab` - matrix.component
- **15**: `feat(frontend): formula analysis tab` - analysis.component
- **16**: `feat(frontend): formula editor tab` - editor.component
- **17**: `feat(ai): copilotkit runtime setup` - copilotkit-runtime/
- **18**: `feat(ai): mcp tool definitions` - copilotkit-runtime/tools/
- **19**: `feat(ai): openai integration` - copilotkit-runtime/openai/
- **20**: `feat(frontend): copilotkit chat panel` - chat.component
- **21**: `feat(infra): docker compose production config` - docker-compose.yml
- **22**: `feat(db): seed data script` - seed/
- **23**: `feat(frontend): production build and nginx config` - nginx.conf
- **24**: `feat(integration): end-to-end verification` - e2e/

---

## Success Criteria

### Verification Commands
```bash
# Backend tests
cd backend && go test ./... -v

# Frontend tests
cd frontend && ng test --watch=false

# Docker Compose up
docker compose up -d

# API health check
curl http://localhost:8080/api/health | jq .

# Formula CRUD smoke test
curl -X POST http://localhost:8080/api/formulas -H 'Content-Type: application/json' -d '{"name":"Test","code":"T-001"}' | jq .
curl http://localhost:8080/api/formulas | jq '. | length'
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] Docker Compose deploys all services
- [ ] CopilotKit chat can create/query/modify/delete formulas
