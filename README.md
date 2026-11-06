# Daily Meal 后端

个人饮食管理工具。本仓库仅维护 Go 后端，已实现每日目标、食物库、固定菜品、三餐记录和营养汇总。后端交付范围见[后端产品文档 V1](docs/backend_product_v1.md)，共同业务规则与整体规划见[总产品文档 V1](docs/product_v1.md)（首版待审查，后续修改须经用户确认），技术实现见 [后端实现文档](docs/implementation_v1.md)，完整接口契约见 [OpenAPI](backend/openapi/openapi.json)。

当前采用学习优先的主干实现，默认调用方提供符合业务约定的参数，保留请求解析、计算、保存和直接错误返回。严格参数校验、跨操作事务、行锁和创建幂等已移除；未覆盖产品文档中校验、配方原子保存和 Agent 重试去重的验收要求，不能据此认定产品验收完成。相关缺口写在代码 `TODO` 中，产品文档的约定与状态没有改变。

当前 OpenAPI 契约版本为 `2.2.0`，路由仍保留 `/api/v1`，包含单轮聊天和截图导入食物入口。此前 `2.0.0` 已移除通用业务参数校验、`409` 幂等冲突和 `Idempotency-Key` 去重；截图接口的 `422` 专门表示识别信息不足或营养基准不匹配。

## 截图添加与更新食物

网页的“新建食物”和“编辑食物”支持选择一张 PNG、JPEG 或 WebP 营养截图，点击“识别并保存”直接写入并显示食物详情。编辑入口始终更新当前食物，不按名称自动匹配其他食物。Go 复用 `backend/config/agent.yaml` 中的模型配置，所选模型需要支持图片输入和 JSON 输出。

- `POST /api/v1/foods/from-image` 新建食物；`PATCH /api/v1/foods/{id}/from-image` 更新指定食物。
- 请求为 `{ "image_data_url": "data:image/png;base64,...", "name": "可选名称" }`。非空名称优先于识别名称；新增缺少名称或热量时返回 `422`。
- Go 调用 Python 的 `POST /api/v1/food-recognitions` 提取原始基准、营养数字与单位，然后换算到每 100g/ml。kJ 除以 4.184 转为 kcal，营养素按 g/mg/ug 换算，明确的份量（例如 250ml）按比例折算。
- 更新只写入识别到的营养，保留缺项、收藏状态和历史记录；不允许 g 与 ml 互转。只有“每份”而没有具体重量/体积时无法保存。
- Python 不访问业务数据库，Go 复用已有食物 Service 保存；原图与识别请求不持久化。识别失败返回 `502`，保存前信息不足返回 `422`，均不写入食物。
- TODO：暂未做重复导入去重；新建请求结果不明确时先查看食物库，避免重复创建。

## 启动

安装 Docker 和 Compose 后，在仓库根目录执行：

```sh
# 聊天前填写 backend/config/agent.yaml 中的模型名和 API Key
docker compose up --build -d
curl http://localhost:8080/healthz
```

Compose 启动 PostgreSQL、执行 Goose 迁移、启动 API。API 地址为 `http://localhost:8080`，接口前缀为 `/api/v1`，文档位于 `/openapi.json`。数据库保存在命名 volume 中。开发环境端口仅绑定本机。

可复制 `.env.example` 为 `.env` 调整数据库密码和业务时区；Compose 会自动读取。默认时区为 `Asia/Shanghai`，本地 PostgreSQL 端口为 `5433`。

### 使用本机 Go 开发

需要 Go 1.25+ 和 PostgreSQL 14+。可只启动 Compose 中的数据库：

```sh
docker compose up -d db
export DATABASE_URL='postgres://dailymeal:dailymeal@localhost:5433/dailymeal?sslmode=disable'
make migrate
make run
```

已有本机 PostgreSQL 时，可以直接使用它，不需要 Docker。以下以数据库用户 `postgres`、端口 `5432` 为例，请替换为自己的数据库用户名和端口。在仓库根目录执行：

```sh
# 仅首次创建数据库
createdb -h 127.0.0.1 -p 5432 -U postgres dailymeal

export DATABASE_URL='postgres://postgres@127.0.0.1:5432/dailymeal?sslmode=disable'
make migrate
make run
```

如果数据库需要密码，在连接地址中使用 `postgres://用户名:密码@127.0.0.1:5432/dailymeal?sslmode=disable`。

Go 程序读取环境变量，不会自动加载 `.env`，所以每次在新终端启动后端前都需要设置 `DATABASE_URL`。可选变量为 `HTTP_ADDR`（默认 `:8080`）、`BUSINESS_TIMEZONE`（默认 `Asia/Shanghai`）。

迁移与 HTTP 服务分开执行，服务启动不会修改表结构。迁移命令为 `go run ./cmd/migrate up|down|status`，在 `backend/` 中运行。`down` 会回滚一版迁移并删除对应业务数据，仅用于可丢弃环境。

## 代码结构

```text
backend/
  config/           Agent 默认 YAML 配置
  cmd/api/          HTTP 服务入口、依赖组装和启动监听
  cmd/migrate/      Goose 迁移入口
  internal/
    config/         环境变量、业务时区、默认模型配置读取
    agent/          Python Agent 的 HTTP 客户端
    database/       GORM 和 PostgreSQL 连接
    httpapi/        路由、JSON 解析、HTTP 状态和统一错误
    service/        目标、食物、菜品、记录和汇总业务
    store/          按业务拆分的 GORM 查询
    model/          饮食模型、历史快照与单轮对话记录
    nutrition/      营养字段定义、单位换算、缺失数据汇总
  migrations/       嵌入二进制的 Goose SQL 迁移
  openapi/          接口请求、响应、单位、空值及示例
```

调用方向为 `HTTP Handler → Service → Store → PostgreSQL`。`nutrition` 是独立计算包，不依赖 HTTP 或数据库。没有额外的通用 Repository 接口或依赖注入框架。

建议先看 `internal/httpapi/router.go`，再从 `handlers.go` 跟进对应的 `service/*.go` 和 `store/*.go`；最后读 `nutrition.go` 和接口集成测试。创建饮食记录的入口是 `Service.SaveEntry`，每日汇总的入口是 `Service.DailySummary`。

业务流程直接顺序查询和写入，不额外开启跨操作事务或行锁；并发读取可能看到不同时间的数据，多步写入失败可能留下部分结果。启动流程不设置专门的连接池参数或优雅停机流程，进程退出时不等待在途请求完成。框架自身的默认行为和数据库原有约束保留。数据库迁移使用显式 SQL，不使用 GORM AutoMigrate。

数量和目标在 PostgreSQL 中使用 `NUMERIC`，营养及快照使用 `JSONB`。Go 使用 `float64` 保留中间计算精度，不逐项取整，显示格式由调用方决定。

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
| 单轮饮食助手 | `POST /api/v1/chat` |

食物、菜品列表支持 `q`、`page`、`page_size`，返回 `items/page/page_size/total`；默认第 1 页、每页 20 条，按 ID 倒序。分页范围由调用方保证有效，服务不再校验最多 100 条。饮食列表要求 `date`，可选 `meal_type`，按 ID 正序。每日汇总要求 `date`。只有创建饮食记录时可省略日期，使用业务时区今天。

默认目标首次查询返回 `null`；日期目标返回 `{date, source, goal}`，`source` 为 `default/override/none`。营养指标覆盖热量、三大营养素、钠、钙、维生素 C 和 D；单位由字段名确定。

PATCH 省略字段保留原值。食物营养逐字段更新，显式 `null` 将可选营养改为未知；调用方应保证热量不为空。食物基准单位创建后固定。饮食 PATCH 仅改数量时默认原快照的 g/ml 基准；调用方提供 `unit` 时应同时给 `quantity`。更换食物需要数量和单位，跨食物/菜品切换需要 `item_type` 和新引用 ID，旧引用自动清空。再次提交同一 ID 不刷新历史快照。这些输入约定不代表服务会完整检查非法组合。

创建饮食记录 `POST /api/v1/meal-entries` 每次成功请求都会新建一条记录，`Idempotency-Key` 请求头被忽略；重复提交可能产生重复记录。原有迁移 3 和 `meal_entry_requests` 表继续保留，避免修改已有迁移历史，但当前业务代码不再读写旧幂等结果。

菜品记录整道计入，调用方不应填写数量、单位或份数。汇总的 `known_total=null` 表示全部未知，`complete=false` 表示存在缺失；空餐/空日为 `0/true`。`comparison` 仅包含热量与三大营养素，目标缺失或摄入不完整时 `remaining=null`。

创建返回 `201`，查询/修改返回 `200`，删除返回 `204`。JSON 语法或类型错误，以及 ID、分页、布尔值、日期的解析错误返回 `400 bad_request`；资源不存在返回 `404 not_found`；数据库等其他错误返回 `500 internal_error`。错误形如 `{"error":{"code":"bad_request","message":"参数解析失败","fields":{}}}`，内部错误不暴露数据库细节。未知的请求结构字段（如创建食物时的 `favorite`）会被忽略，不再返回 `400`；`nutrition_per_100` 的营养键仍应按文档定义提交。参数的必填、范围、枚举和组合条件是调用方应遵守的约定，不承诺逐项检查；数据库原有约束拒绝写入时按数据库错误返回。

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

测试围绕迁移 up/down/up、历史目标和覆盖优先级、单位换算、未知/零/部分已知、配方保存、历史快照、数量修改、跨类型切换、补记/移动/删除、目标剩余值，以及 OpenAPI 请求和响应匹配。验证通过仅表示当前实现范围通过，不代表已满足严格校验、配方原子性或重试去重的产品验收要求。CI 使用 PostgreSQL 17，执行 race 检测、静态检查和构建。

## 后续开发

已接入 Python Agent 的单轮对话入口；截图识别、会话历史加载、登录、异步任务和 SSE 按后续阶段推进。

## 单轮饮食助手

网页的「饮食助手」向 Go 的 `POST /api/v1/chat` 发送当前消息。Go 读取默认模型配置，调用 Python Agent，保存本轮结果后返回。每条消息独立执行，暂不加载历史、建立会话或重放旧结果。

仓库提供 `backend/config/agent.yaml` 默认配置。模型名和 API Key 初始为占位值，聊天前填写自己的值。

在 `backend/config/agent.yaml` 的 `default` 中直接填写 `agent_url`、`api_key`、`base_url`、`model` 和 `use_responses_api`。当前固定读取这一项，没有配置选择或环境变量覆盖；每次聊天读取文件，修改后下一次请求生效。本机 `make run` 在 `backend` 目录运行，读取 `config/agent.yaml`。默认 YAML 随代码提交，没有新增 Git 忽略规则；以后填入的真实密钥不要加入代码提交。

设置 `DATABASE_URL` 后运行 `make migrate` 和 `make run`。在相邻 `daily_meal_agent` 仓库安装依赖并启动 Agent，再从 `daily_meal_frontend` 启动网页；完整启动命令见各端 README。API Key 只配置在 Go 这一侧，Agent 的 `.env` 只需业务 API 地址与时区。

```sh
curl http://127.0.0.1:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"查一下今天吃了多少热量"}'
```

成功响应包含 `id`、`message`、`reply`、`status`、`tool_results` 和记录时间。Go 的 `chat_requests` 表保存消息、回复、工具结果和 `running/completed/failed` 状态，不保存模型配置或 Key。记录目前只写入，不提供历史查询、不读取为上下文。

Agent 报错时，Go 保存已返回的工具结果并返回 `502 agent_error`。每次 POST 都独立执行，重复发送可能重复记餐；目前没有请求去重、自动重试或跨工具事务。进程或连接中断可能留下 `running` 或缺少部分工具结果，先检查实际饮食数据。

Docker Compose 将本地 YAML 挂载到 Go 容器的 `/app/config/agent.yaml`，与镜像的工作目录 `/app` 对应。使用 Agent 仓库的 Compose 同时启动服务时，将 YAML 的 `agent_url` 改为 `http://agent:8000`；模型 `base_url` 也须能从 Agent 容器访问。配置只挂载给 Go，不打包进镜像。
