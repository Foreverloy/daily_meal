# Daily Meal 后端联调验收报告

验收日期：2026-10-12，业务时区 Asia/Shanghai。

原始测试证据已存档到仓库外的 `/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence`；报告中的证据链接指向该本机目录，仓库仅保留两份验收报告。

**结论：本地后端功能验收通过。** 静态检查、构建、21 个顶层测试（含 10 个真实 PostgreSQL 集成测试）、race 检测和补充的真实 HTTP 三餐流程均通过，未发现已确认的后端业务缺陷。前端仍有两项营养展示问题，详见[前端报告](frontend_report.md)，整体 V1 尚不能判定全部验收通过。

本报告按现存文档核查实现符合性。总产品文档、后端产品文档及前端产品文档均仍标注“待用户审查”；本次验收不改变产品审批状态，也不代表生产部署验收。

## 1. 验收对象与环境

| 项目 | 本次对象 |
| --- | --- |
| 后端仓库 | `/Users/loyforever/Developer/daily_meal` |
| 后端版本 | `a342f791c6f009c59f40209f307dc3572d15a093`，验收开始时工作树干净 |
| 需求基准 | [总产品文档](../../product_v1.md)第 4、5、8 节；[后端产品文档](../../backend_product_v1.md)第 2～4 节 |
| 技术契约 | [实现文档](../../implementation_v1.md)、[OpenAPI](../../../backend/openapi/openapi.json) |
| 运行环境 | macOS arm64，Go 1.25.5，PostgreSQL 14.19 |
| 数据隔离 | 全新临时实例，仅绑定 `127.0.0.1:55432`；`dailymeal_acceptance_test` 和 `dailymeal_e2e_test` |
| 实际联调 | Go API `127.0.0.1:18081`；前端临时副本 Vite `127.0.0.1:5175`，通过 `/api` 代理访问 Go |
| 自动 E2E | 另用 API 18080、Vite 5174 和 `dailymeal_e2e_test`，与人工验收分开 |

没有连接或修改个人饮食数据库。Go 集成测试只在专用 `_test` 数据库中创建随机 schema，结束清理自己的 schema。本次未改业务代码、原有测试断言或产品文档；构建只生成被 Git 忽略的二进制，报告与证据为新增验收产物。

## 2. 执行结果

| 检查 | 结果 | 证据 |
| --- | --- | --- |
| `GOCACHE=/private/tmp/daily-meal-acceptance-go-cache make check` | 退出 0，gofmt / go vet 通过 | [check 日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/backend-check.log) |
| 同一 GOCACHE 下 `make build` | 退出 0，API / migrate 构建通过 | [build 日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/backend-build.log) |
| 未设置测试库的 `go test -count=1 -json ./...` | 11 个顶层测试通过，10 项集成测试明确跳过；仅作为首轮结果 | [首轮日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/backend-unit.log) |
| 设置专用 `TEST_DATABASE_URL` 后 `make test-integration` | 退出 0，21 个顶层测试通过、0 失败、0 跳过；race 未报告数据竞争 | [集成日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/backend-integration.log)、[测试汇总](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/backend-integration-summary.json) |
| HTTP / OpenAPI 验证 | 176 次 httptest 请求，其中 158 次集成 fixture 请求校验成功请求及全部响应；覆盖全部 19 个业务方法与路径组合 | [请求汇总](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/backend-http-request-summary.json) |
| 实际启动 / Goose | API 成功监听；空测试库迁移至版本 2；真实 `/healthz` 返回 `{"status":"ok"}` | [真实请求](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/live-api-requests.json) |
| 补充网络联调 | 41 次真实 HTTP 请求、323 次独立断言通过；原始输入独立核算，不使用接口结果反算预期 | [结果](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/live-api-result.json)、[复跑脚本](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/live-api-acceptance.py) |
| 浏览器联调 | 12 个前端单元用例、6 个现有 E2E 全过；另手工完成目标、重量/体积食物、混合配方、整道记录、汇总 | [前端报告](frontend_report.md) |

数据库测试执行时使用：

```sh
TEST_DATABASE_URL='postgres://dailymeal_acceptance@127.0.0.1:55432/dailymeal_acceptance_test?sslmode=disable' \
GOCACHE=/private/tmp/daily-meal-acceptance-go-cache \
GOFLAGS='-json -coverprofile=/private/tmp/daily-meal-acceptance/backend-coverage.out' \
make test-integration
```

请求状态分布为 200×94、201×23、204×5、400×12、404×5、422×35、500×2；两次 500 属于受控数据库失败用例，均符合预期。21 为顶层测试计数；包含子案例共 68 次 pass 事件，不能相加作为独立产品场景数。

默认逐包覆盖率为 HTTP 包 91.7%、营养包 92.5%。本次没有设置跨包 `-coverpkg`，HTTP 集成调用的 Service / Store 不计入它们自己的直接包覆盖率；不能据 Service 6.0% 或整体 31.9% 推断业务验收覆盖不足。[原始覆盖率](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/backend-coverage.out)仅保留为测试证据。

## 3. 产品场景验收

| 场景 | 实际验证与结果 | 主要自动化证据 |
| --- | --- | --- |
| 完整核心流程 | 目标→食物→菜品→饮食→汇总可运行；补充同日早/午/晚餐，共 4 条记录，真实 API / 页面一致 | `TestIntegrationSnapshotsAndSummary`、现有 E2E、补充 HTTP / UI |
| 目标计算 | 2000 kcal、30/40/30% → 150 / 200 / 66.666…g；页面显示舍入后的 66.7g | `TestGoalTargets`、真实 HTTP / UI |
| 目标历史与覆盖 | 更早默认目标保留；覆盖优先；删除恢复适用默认或 none；首次默认前不会反向套用 | `TestIntegrationGoalHistory`、真实 HTTP |
| 无目标 | 可保存饮食和计算摄入；目标与剩余为 null，页面有设置引导 | `TestIntegrationSnapshotsAndSummary`、真实 HTTP / UI |
| 用量与单位 | 0.25kg→250g、0.25L→250ml；跨重量/体积、份、非法/非正数量拒绝 | `TestQuantity`、`TestIntegrationFoodsAndRecipes`、真实 HTTP |
| 菜品与混合记录 | 重量与体积食材求和；整道记录不填写份数；同餐可同时含食物和菜品 | `TestIntegrationSnapshotsAndSummary`、真实 HTTP / UI |
| 配方原子性 | 非法食材及写入后计算失败整事务回滚；失败替换不留下部分名称或食材 | `TestIntegrationRecipeRollbackAfterWrite`、真实 HTTP |
| 历史快照 | 修改食物/配方不改变旧记录；旧记录数量调整继续用原营养，新记录使用当前模板 | `TestIntegrationSnapshotsAndSummary`、真实 HTTP、E2E |
| 未知 / 零 / 部分已知 / 空餐 | 8 指标齐全；null/false、0/true、已知合计/false 分别正确；空餐不污染全未知日汇总 | `TestUnknownZeroAndPartialTotals`、集成测试、真实 HTTP |
| 超出目标 | complete=true 时保留负剩余；前端原有展示明确“超出” | `TestIntegrationSnapshotsAndSummary`；前端超出场景未另作手工视觉验证 |
| 纠正记录 | 补记、改量、改餐、跨日移动与删除更新汇总；移动原日/新日都核验 | 集成测试、真实 HTTP、E2E |
| 错误与契约 | 类型/未知字段/多 JSON 等 400；不存在 404；业务 422；内部错误 500 不泄漏 SQL；204 空体 | `TestTransportErrors`、fixture OpenAPI 校验 |
| 数据库迁移 | 空库 up、幂等 up、down/up、收藏存量默认值和回滚通过 | `TestIntegrationMigrations`、`TestIntegrationFoodFavoritesMigration` |

测试实现位置见[隔离与契约 fixture](../../../backend/internal/httpapi/integration_test.go)、[营养测试](../../../backend/internal/nutrition/nutrition_test.go)、[收藏集成测试](../../../backend/internal/httpapi/food_favorites_integration_test.go)。

## 4. 三餐独立核算样例

补充验证日期为 `2026-10-10`。米饭原始基准每 100g 为 100 kcal、2g 蛋白、20g 碳水、0g 脂肪；牛奶每 100ml 为 60 kcal、3g 蛋白、5g 碳水、3g 脂肪；鸡肉每 100g 为 150 kcal、30g 蛋白、0g 碳水、3g 脂肪。混合菜品为米饭 100g + 牛奶 200ml。历史米饭模板后改为 200 kcal，旧记录仍按原始 100 kcal 基准计算。

| 指标 | 早餐：米饭 250g | 午餐：牛奶 250ml + 整道菜 | 晚餐：鸡肉 200g | 全天 |
| --- | ---: | ---: | ---: | ---: |
| 热量 kcal | 250 | 370 | 300 | 920 |
| 蛋白质 g | 5 | 15.5 | 60 | 80.5 |
| 碳水 g | 50 | 42.5 | 0 | 92.5 |
| 脂肪 g | 0 | 13.5 | 6 | 19.5 |
| 钠 mg | 0 | 180 | 140 | 320 |
| 钙 mg | 25 | 550 | 16 | 591 |
| 维生素 C mg | 未知，不完整 | 已知 0，不完整 | 0，完整 | 已知 0，不完整 |
| 维生素 D μg | 未知，不完整 | 已知 2.25，不完整 | 0，完整 | 已知 2.25，不完整 |

日期覆盖目标为 2000 kcal、30/40/30%；剩余为 1080 kcal、蛋白 69.5g、碳水 107.5g、脂肪 47.166…g。浏览器所展示全天数值及三餐热量一致；每餐其他指标已由后端正确返回，前端缺少展示，见 F-01。

## 5. 收藏扩展

当前代码 / OpenAPI 使用 `favorite` 字段。默认 false、重复收藏/取消、与名称搜索和分页组合、`favorite=false` 表示全部、持久化、非法参数回滚、迁移及不影响营养/配方/历史均通过。

当前总 / 各端产品正文尚未列收藏需求，因此作为已实现扩展接口回归记录。网页没有收藏 / 取消 / 收藏筛选入口，收藏功能尚未完成网页联调；不将其自行追认为原 V1 必做需求，也未修改产品文档。

## 6. 未执行项与交付建议

- PostgreSQL 17、Docker Compose 全栈构建与启动未执行；当前主机没有可用 `docker` 命令。PostgreSQL 14.19 满足 README 所列本机最低版本，但不等价于验证部署目标版本。
- 远程 CI、生产数据库、生产部署、前端 Docker / Nginx 的代理与深层刷新未验证。
- 未执行压力、容量、长时间运行或所有业务写入的并发测试；race 通过不等价于这些验证。
- 登录、多用户、手机、Agent、截图识别、推荐、iOS 为文档排除范围，不计本次缺陷。

后端可以进入后续联调修正阶段；整体交付前应修正前端 F-01 / F-02 并回归。产品需求审批仍由用户完成。验收结束后已停止本次启动的前端、API 和临时数据库；复跑需重新启动隔离环境，勿将脚本改指个人数据库。
