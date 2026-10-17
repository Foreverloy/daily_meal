# Daily Meal 前端联调验收报告

验收日期：2026-10-12，业务时区 Asia/Shanghai。

原始测试证据已存档到仓库外的 `/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence`；报告中的证据链接指向该本机目录，仓库仅保留两份验收报告。

**结论：前端可启动，核心业务闭环可用；按当前产品文档的完整符合性验收未通过。** 类型检查、格式、生产构建、12 个单元用例及 6 个现有 E2E 用例全部通过，但真实浏览器复现两项营养展示缺陷：每餐只展示热量；全部未知指标缺少“不完整”标识。后端正确返回相应数据，详见[后端报告](backend_report.md)。

总 / 前端产品文档仍标注“待用户审查”。本报告核查现存规则，不批准产品需求、不改写产品文档，也不将测试全绿等同于全部业务验收通过。

## 1. 验收对象与启动结果

| 项目 | 本次对象 |
| --- | --- |
| 前端仓库 | `/Users/loyforever/Developer/daily_meal_frontend` |
| 版本边界 | 原仓库无可解析 HEAD，全部源文件为 untracked；验收对象为当前工作区快照 |
| 快照证据 | [46 文件 SHA256 清单](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/frontend-snapshot-manifest.json)，覆盖源码、E2E、配置、产品文档及 Go OpenAPI |
| 需求基准 | [总产品文档](../../product_v1.md)第 4、5、8 节；前端 `docs/product_v1.md` 第 2～4 节 |
| 运行方式 | 在 `/private/tmp/daily-meal-acceptance/frontend` 完整复制源码和已有依赖后执行，不改原前端仓库 |
| 环境 | Node 24.12.0、npm 11.6.2、Vue 3.5.43、Vite 8.3.0、TypeScript 6.0.3、Vitest 4.1.11、Playwright 1.63.0 |
| 浏览器 | Chromium Headless Shell 153.0.8010.12 执行现有 E2E；Ego Lite 执行手工页面流程 |
| 数据 | 全新 PostgreSQL 14.19 实例，两套名称以 `_test` 结尾的专用数据库 |

前端实际监听 `127.0.0.1:5175` 并通过 Vite `/api` 代理访问 Go API `127.0.0.1:18081`，浏览器成功加载四模块及真实业务数据。现有 E2E 单独使用 5174 / 18080，不与手工验收数据混用。

初轮 E2E 在浏览器启动前因缺 Chromium 全部失败，尚未执行业务断言。补齐仅位于临时目录的浏览器后，原样重跑 6/6 通过；该环境问题已经解除，不记为当前启动阻碍。当前主机没有可用 Docker，采用原生 Go + PostgreSQL + Vite 完成联调。

## 2. 自动化执行结果

| 命令 | 实际结果 | 证据 |
| --- | --- | --- |
| `npm run format:check` | 退出 0，匹配文件格式检查通过 | [日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/frontend-format.log) |
| `npm run build` | 退出 0，包含 vue-tsc 类型检查；Vite 构建 61 modules | [日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/frontend-build.log) |
| `npm run test:unit` | 退出 0，1 文件 / 12 用例全部通过 | [日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/frontend-unit.log) |
| `npm run test:e2e` | 补齐浏览器后退出 0，6/6 desktop 用例通过，9.3 秒，单 worker、无 retry | [日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/frontend-e2e.log) |

E2E 环境变量：

```sh
E2E_TEST_DATABASE_URL='postgres://dailymeal_acceptance@127.0.0.1:55432/dailymeal_e2e_test?sslmode=disable'
E2E_BACKEND_DIR='/Users/loyforever/Developer/daily_meal/backend'
GOCACHE='/private/tmp/daily-meal-acceptance-go-cache'
PLAYWRIGHT_BROWSERS_PATH='/private/tmp/daily-meal-acceptance/playwright-browsers'
```

E2E 自动迁移数据库至版本 2，构建真实 API，再启动 Vite。业务正向流程使用真实 PostgreSQL / Go；422、读请求中断、写请求中断则由 Playwright 注入，属于前端错误处理验证。[运行版本](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/frontend-runtime-versions.log)、[初轮浏览器缺失日志](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/frontend-e2e-browser-missing.log)一并保留。

单元验证涵盖未知 / 零 / 部分已知、PATCH 字段差异及 null、日期与闰日、204、字段错误、写入中断不自动重发、旧请求丢弃、保存锁、422 输入保留、历史快照、菜品无用量。

现有 E2E 涵盖目标、体积换算、同食材去重、同餐混合记录、旧快照、移日 / 改餐 / 删除、日期覆盖、字段错误、刷新 / 重试、未知与零、修改配方、记录类型切换、13 条搜索分页及写入不确定提示。

## 3. 补充真实浏览器与三餐核验

手工从页面保存 2000 kcal、30/40/30% 默认目标；创建燕麦（g）和牛奶（ml）；重复选燕麦仍只有一行；保存燕麦 50g + 牛奶 0.2L 的配方；从详情记入整道菜，表单没有菜品份数或用量输入；保存后真实汇总为 295 kcal。

独立核算为 `350×50/100 + 60×200/100 = 295 kcal`。蛋白质已知 6g、不完整，碳水 40g、脂肪 9g、钠已知 0mg，钙 / 维生素 C / D 全部未知。API 保留上述完整性，页面部分已知与零正确，全部未知缺少完整性提示（F-02）。证据：[配方 API](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-recipe-api.json)、[配方页面](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-recipe.txt)、[整道记录表单](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-recipe-entry-form.txt)、[保存后汇总](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-summary-api.json)。

另以真实 HTTP 建立 `2026-10-10` 同日早 / 午 / 晚餐四条记录，并在浏览器查看：

| 项目 | 独立核算 / API | 页面实际 |
| --- | --- | --- |
| 早餐 / 午餐 / 晚餐热量 | 250 / 370 / 300 kcal | 一致 |
| 全天热量 | 920 kcal，目标 2000，剩余 1080 | 一致 |
| 蛋白质 | 80.5g，目标 150，剩余 69.5 | 一致 |
| 碳水 | 92.5g，目标 200，剩余 107.5 | 一致 |
| 脂肪 | 19.5g，目标 66.666…，剩余 47.166… | 显示 66.7 / 47.2，属于显示舍入 |
| 全天钠 / 钙 | 320mg / 591mg | 一致 |
| 全天维生素 C / D | 已知 0mg / 2.25μg，均不完整 | 显示 0 / 2.3，并标明数据不完整 |
| 每餐其他 7 项营养 | 后端全部返回 | 页面无展示入口，F-01 |

三餐证据：[页面快照](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-three-meals.txt)、[DOM 与窗口结果](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-three-meals-observed.json)、[独立 HTTP 核算](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/live-api-result.json)。三餐页面核对了全部已展示数字；每餐未展示指标的缺口没有用接口通过替代。

## 4. 确认缺陷

本报告 P1 表示应在完整 V1 验收通过前修正的功能缺口，P2 表示影响信息表达 / 操作可靠性的质量问题；均依据本次现存文档与实际行为评级。

### F-01 每餐缺少其余七项营养汇总

- **级别：P1；状态：真实浏览器确认。**
- 需求：总产品第 5.5 节第 111～120 行要求分别展示每餐和全天全部支持指标，包括未知指标。
- 复现：在三餐都有记录的 `2026-10-10` 日记页查看早餐、午餐、晚餐，再展开“更多营养”。
- 预期：每餐均可查看热量、蛋白质、碳水、脂肪、钠、钙、维生素 C / D 的合计和完整性。
- 实际：三餐卡片只显示 kcal；“更多营养”只显示全天数据。早餐蛋白 5g、午餐 15.5g、晚餐 60g 已由 API 正确返回，用户无法在网页查看对应每餐合计。
- 代码：前端 `src/views/TodayView.vue:124`～166 只取每餐 `energy_kcal`；168～170 仅把全天 `nutrition_totals` 传给 NutritionGrid。
- 影响：每餐营养验收条款不满足，数据保存和后端计算没有因该问题出错。
- 建议：增加每餐展开区域或详情入口，复用后端每餐 Totals；回归每餐 8 项、空餐 0/完整和缺项提示。
- 证据：[三餐 DOM](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-three-meals-observed.json)、[页面快照](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-three-meals.txt)、[API 独立结果](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/live-api-result.json)。

### F-02 全部未知指标缺少数据不完整标识

- **级别：P2；状态：真实浏览器确认。**
- 需求：总产品 `docs/product_v1.md:117` 要求有记录但某指标全部未知时显示“未知”，并标明数据不完整。
- 复现：创建钙、维生素 C / D 留空的燕麦与牛奶，保存混合配方，查看菜品营养详情。
- 预期：上述指标显示“未知，数据不完整”。
- 实际：只显示“未知”及单位；同页蛋白质部分已知则正确显示“6（已知合计，数据不完整）”。API 对全部未知项已返回 `{"known_total":null,"complete":false}`。
- 代码：前端 `src/lib/format.ts:36`～39 的 `amountText` 对 null 直接返回“未知”；`src/components/NutritionGrid.vue:11` 复用它。快照与全天更多营养同样复用该函数。今日宏量卡片另有不完整文案，不能误报所有页面都缺提示。
- 影响：营养完整性表达不符合现存规则，不影响后端未知值保存或求和。
- 建议：让全未知且 incomplete 的 Totals 也展示完整性；保留普通食物基准 Values 的未知语义；补全未知 / 部分未知 / 零 / 空餐显示回归。
- 证据：[配方页面](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-recipe.txt)、[配方 API](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/ui-recipe-api.json)。

## 5. 产品场景验收矩阵

| 场景 | 页面 / 交互结果 | 验收状态 |
| --- | --- | --- |
| 四模块及核心流程 | 导航、目标、食物、菜品、记录与汇总实际可运行 | 通过 |
| 目标克数、来源、日期覆盖与恢复操作 | 正向 E2E 与手工目标保存通过；历史数值 / 恢复 none 由后端真实 HTTP 补核 | 已测交互通过；浏览器历史 / 恢复数值未独立核验 |
| 没有目标仍可记录 | 初始无目标页面提示与暂无法计算正确；后端无目标记录通过 | 通过 |
| 单位与整道菜 | 页面只给同类单位；L/ml 换算、混合配方与无份数表单通过 | 通过 |
| 搜索 / 分页 / 空结果 | 13 条搜索显示 12+1，空结果正确 | 通过 |
| 模板修改及历史快照 | E2E 核对旧名称、配方旧营养、数量修改、类型切换 | 通过 |
| 补记 / 跨日移动 / 删除 | E2E 页面导航与删除通过；真实 HTTP 核对两日汇总 | 通过 |
| 未知、已知零和部分合计 | 数字与部分已知提示正确；全未知缺不完整提示 | 部分通过，F-02 |
| 每餐与全天全部营养 | 全天正确，每餐只展示热量 | 未通过，F-01 |
| 422、连接失败、重试与保留输入 | 模拟错误下输入保留、字段错误和重新加载通过 | 通过，限已测错误路径 |
| 重复提交与写入不确定 | unit / E2E 锁定提交、无自动重复 POST、引导查看已保存数据 | 通过 |
| 快速变化丢弃旧响应 | 单元验证旧结果不覆盖当前结果 | 通过，未另做浏览器延迟注入 |
| 深层路由刷新与桌面布局 | Vite 刷新可用；现有 E2E 无横向溢出；1280×720 截图已目视检查 | 通过，限已测开发环境 / 窗口 |
| 超出目标的视觉文案 | 源码有“超出”，后端负剩余通过 | 未额外做真实浏览器超额场景 |
| 全键盘 / 多浏览器 / 生产托管 | 不属于本次已执行证据 | 未执行 |

现有主流程用例名称含“混合三餐”，实际只创建同日早餐食物 + 菜品，再移动一条至另一日晚餐；本次因此补做同日三餐核算。日期覆盖 E2E 主要断言保存 / 恢复提示，未独立核对恢复数值。现有绿灯没有捕获 F-01 / F-02，应补有意义的展示断言。

## 6. 扩展差异与操作观察

**收藏尚未完成网页接入。** 后端 OpenAPI 已有 `favorite`、PATCH 与筛选；前端 Food / FoodPatch 类型、列表查询和 UI 没有收藏功能。当前产品正文未列收藏，因此作为扩展交付差异，未据此把原 V1 判为失败，也未补写需求。

**重新选择同一食物可能改变原用量，属于源码观察。** `src/views/EntryForm.vue:63`～78 即使判断 same=true，仍无条件把用量重置为 100 基准单位。原 250g 记录重新选择同一项可能被改为 100g。浏览器已核对原用量为 250，但重新选择的操作因自动化指针命中失败未完成，因此未将此项记为运行确认缺陷；当前产品也未明确该交互预期。建议后续复核并优先保留同一食物原用量。

## 7. 证据限制与交付建议

- 已目视检查[现有 E2E 截图](/Users/loyforever/Documents/Codex/daily-meal-acceptance/2026-10-12/evidence/e2e-daily-summary.png)：1280×720 视口中的每日能量、宏量卡片、三餐布局清晰；不能据此代表全部页面或窗口尺寸都已验证。
- Ego 的页面截图接口两次发生 `Page.captureScreenshot` 超时；手工缺陷证据采用实际 DOM 快照与 API 对照，没有把缺失截图宣称为视觉通过。
- 未重跑从零 `npm ci`；执行副本采用已有依赖完整副本。
- Docker / Nginx 生产代理与 SPA fallback、远程 CI、Safari / Firefox、全部键盘流程、手机均未验证。手机 / iOS / Agent 在当前文档排除范围内。
- 产品文档待审查，前端源仓库无提交版本；发布前应落实需求审批与版本留存。

建议先修正 F-01 / F-02，补每餐 8 指标和全未知完整性的展示断言，再回归现有 6 项 E2E 与本次三餐 / 配方样例。本次仅验收与出具报告，没有实施修复。验收结束后已停止测试服务与临时数据库。
