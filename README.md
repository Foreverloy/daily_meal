# Daily Meal

个人饮食管理工具。Go 后端 V1 已实现每日目标、食物库、固定菜品、三餐记录和营养汇总。产品需求与验收标准见 [产品文档 V1](docs/product_v1.md)（首版待审查，后续修改须经用户确认），技术实现见 [后端实现文档](docs/implementation_v1.md)，完整接口契约见 [OpenAPI](backend/openapi/openapi.json)。

## 启动

安装 Docker 和 Compose 后，在仓库根目录执行：

```sh
docker compose up --build -d
curl http://localhost:8080/healthz
```

Compose 依次启动 PostgreSQL、执行 Goose 迁移、启动 API。API 地址为 `http://localhost:8080`，接口前缀为 `/api/v1`，文档位于 `/openapi.json`。数据库保存在命名 volume 中。开发环境端口仅绑定本机。

可复制 `.env.example` 为 `.env` 调整数据库密码和业务时区；Compose 会自动读取。默认时区为 `Asia/Shanghai`，本地 PostgreSQL 端口为 `5433`。

### 使用本机 Go 开发

需要 Go 1.25+ 和 PostgreSQL 14+。可只启动 Compose 中的数据库：

```sh
docker compose up -d db
export DATABASE_URL='postgres://dailymeal:dailymeal@localhost:5433/dailymeal?sslmode=disable'
make migrate
make run
```

已有 PostgreSQL 时创建专用数据库，把 `DATABASE_URL` 指向它即可。Go 程序读取环境变量，不会自动加载 `.env`。可选变量为 `HTTP_ADDR`（默认 `:8080`）、`BUSINESS_TIMEZONE`（默认 `Asia/Shanghai`）。

迁移与 HTTP 服务分开执行，服务启动不会修改表结构。迁移命令为 `go run ./cmd/migrate up|down|status`，在 `backend/` 中运行。`down` 会回滚一版迁移并删除对应业务数据，仅用于可丢弃环境。

## 代码结构

```text
backend/
  cmd/api/          HTTP 服务入口、依赖组装、优雅退出
  cmd/migrate/      Goose 迁移入口
  internal/
    config/         环境变量、业务时区
    database/       GORM 和 PostgreSQL 连接池
    httpapi/        路由、JSON 解析、HTTP 状态和统一错误
    service/        目标、食物、菜品、记录和汇总业务；事务边界
    store/          按业务拆分的 GORM 查询
    model/          六张业务表的模型和历史快照结构
    nutrition/      营养字段定义、单位换算、缺失数据汇总
  migrations/       嵌入二进制的 Goose SQL 迁移
  openapi/          接口请求、响应、单位、空值及示例
```

调用方向为 `HTTP Handler → Service → Store → PostgreSQL`。`nutrition` 是独立计算包，不依赖 HTTP 或数据库。没有额外的通用 Repository 接口或依赖注入框架。

建议先看 `internal/httpapi/router.go`，再从 `handlers.go` 跟进对应的 `service/*.go` 和 `store/*.go`；最后读 `nutrition.go` 和接口集成测试。创建饮食记录的入口是 `Service.SaveEntry`，每日汇总的入口是 `Service.DailySummary`。

业务服务负责事务。涉及多次查询的配方读取、快照创建和每日汇总使用 PostgreSQL `REPEATABLE READ`，保证同一次操作看到一致数据；PATCH 和配方替换先锁定被修改行。数据库迁移使用显式 SQL，不使用 GORM AutoMigrate。

数量和目标在 PostgreSQL 中使用 `NUMERIC`，营养及快照使用 `JSONB`。Go 使用 `float64` 保留中间计算精度，不逐项取整，显示格式由前端决定。

## 接口要点

| 操作 | 路径 |
| --- | --- |
| 当前默认目标 | `GET/PUT /api/v1/goal-settings/default` |
| 日期目标 | `GET/PUT/DELETE /api/v1/daily-goals/{date}` |
| 食物列表、创建 | `GET/POST /api/v1/foods` |
| 食物详情、更新 | `GET/PATCH /api/v1/foods/{id}` |
| 菜品列表、创建 | `GET/POST /api/v1/recipes` |
| 菜品详情、替换 | `GET/PUT /api/v1/recipes/{id}` |
| 饮食列表、创建 | `GET/POST /api/v1/meal-entries` |
| 饮食详情、修改、删除 | `GET/PATCH/DELETE /api/v1/meal-entries/{id}` |
| 每日汇总 | `GET /api/v1/daily-summary?date=YYYY-MM-DD` |

食物、菜品列表支持 `q`、`page`、`page_size`，返回 `items/page/page_size/total`；默认 20 条，最多 100 条，按 ID 倒序。饮食列表要求 `date`，可选 `meal_type`，按 ID 正序。每日汇总要求 `date`。只有创建饮食记录时可省略日期，使用业务时区今天。

默认目标首次查询返回 `null`；日期目标返回 `{date, source, goal}`，`source` 为 `default/override/none`。营养指标覆盖热量、三大营养素、钠、钙、维生素 C 和 D；单位由字段名确定。

PATCH 省略字段保留原值。食物营养逐字段更新，显式 `null` 将可选营养改为未知；热量不能清空。食物基准单位创建后固定。饮食 PATCH 仅改数量时默认原快照的 g/ml 基准；给 `unit` 时必须同时给 `quantity`。更换食物需要数量和单位，跨食物/菜品切换需要 `item_type` 和新引用 ID，旧引用自动清空。再次提交同一 ID 不刷新历史快照。

菜品记录整道计入，不能填写数量、单位或份数。汇总的 `known_total=null` 表示全部未知，`complete=false` 表示存在缺失；空餐/空日为 `0/true`。`comparison` 仅包含热量与三大营养素，目标缺失或摄入不完整时 `remaining=null`。

创建返回 `201`，查询/修改返回 `200`，删除返回 `204`。错误形如 `{"error":{"code":"validation_error","message":"业务参数不合法","fields":{"quantity":"数量必须大于 0"}}}`，格式/未知请求字段错误为 `400`，资源不存在为 `404`，业务校验为 `422`，内部错误为 `500`。

## 验证

```sh
make test   # 单元测试和 HTTP 格式测试；无 TEST_DATABASE_URL 时明确跳过集成测试
make check  # gofmt + go vet
make build
```

集成测试使用真实 PostgreSQL 和 `httptest`，必须提供名字以 `_test` 结尾的独立数据库。每个测试使用随机 schema，结束时只删除自己的 schema。不会使用 `DATABASE_URL` 或清空个人饮食数据。

```sh
docker compose exec db createdb -U dailymeal dailymeal_test
export TEST_DATABASE_URL='postgres://dailymeal:dailymeal@localhost:5433/dailymeal_test?sslmode=disable'
make test-integration
```

测试覆盖迁移 up/down/up、历史目标和覆盖优先级、单位换算、未知/零/部分已知、配方事务回滚、历史快照、数量修改、跨类型切换、补记/移动/删除、目标剩余值，以及 OpenAPI 请求和响应匹配。CI 使用 PostgreSQL 17，执行 race 检测、静态检查和构建。

## 后续开发

前端放在 `frontend/`，Python Agent 后续作为独立服务接入 Go API。当前实现专注单人后端 V1。前端、Agent、登录、异步任务和 SSE 按后续阶段推进。
# daily_meal
