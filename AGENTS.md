### 项目背景

我是 Ruby 后端程序员，准备转向 Go 和 Agent 应用开发，已经学习了 Go 基础与 LangChain。这个项目结合个人健身餐需求，用于练习后续工作可能涉及的技术。

### 项目介绍

Daily Meal 是个人饮食管理工具，参考“薄荷健康”相关功能和交互，支持每日营养目标、食物库、菜品配方、三餐记录和营养汇总。

V1 先实现 Go 后端。后续接入 Python Agent，通过截图维护食物，通过对话记录饮食、调整目标、录入菜品和推荐菜单，复用 Go 业务接口。

前端由 AI 实现，先支持电脑和手机网页，后续通过 Capacitor 支持 iOS App。用户主要专注 Go 与 Agent 开发。

### 业务约定

- 目标：设置每日大卡及蛋白质、碳水、脂肪的供能占比；默认每天沿用，允许按日期调整，保留历史目标。
- 食物：按每 100g 或每 100ml 保存营养；支持 g/kg、ml/L 同类换算，不支持“个”“份”等单位或重量与体积互转。
- 营养：钠、维生素等只记录和汇总，不设置目标；缺失值保持未知，汇总存在缺项时标明不完整。
- 菜品：固定食物组成和用量，营养按食材求和；默认整道吃完，不引入成品称重、分份或剩余量。
- 记录：按日期和早、中、晚餐混合记录食物、菜品；支持补记、修改和删除，历史营养保留当时快照。
- 后续 Agent 调整菜品时，只改变本次配方；用户明确要求修改模板时才更新原菜品。推荐与实际饮食记录分开。
- 保持实现简单，围绕已确认需求开发。

### 技术栈

- 最终架构：同一仓库，Vue 前端调用 Go 业务后端，Python Agent 作为独立服务。
- 前端：Vue 3 + TypeScript + Vite + Vue Router，使用 HTML/CSS 构建响应式页面；后续接入 Capacitor。
- Go 后端：Gin + GORM + Goose，负责业务逻辑、营养计算和业务数据。
- Python Agent：FastAPI + Pydantic + LangChain + LangGraph，负责模型调用、工具调用和 Agent 状态。
- 数据库：PostgreSQL；业务数据与 Agent 检查点使用独立数据库或 schema，Python 通过 Go 内部接口访问业务数据。
- 服务通信：HTTP/JSON + OpenAPI；Agent 阶段使用 SSE 沿 Python → Go → 页面返回流式输出。
- 开发部署：Docker Compose；测试使用 Go testing/httptest、Python pytest，并接入 CI。
- 后续扩展：Redis + Asynq，由 Go Worker 调用 Python Agent；OpenTelemetry 用于跨服务追踪。

### 实现文档

- [后端 V1 实现文档](docs/implementation_v1.md)：业务规则、数据模型、接口和开发顺序。
- [前端 V1 实现文档](docs/frontend_implementation_v1.md)：Vue 技术栈、页面、接口对接及后续 iOS 接入。

### Git 提交规范

- 用户要求提交时，必须遵循[约定式提交规范 1.0.0](https://www.conventionalcommits.org/zh-hans/v1.0.0/)，由 Agent 根据实际改动编写提交信息。
- 提交标题格式为 `<type>[可选 scope][可选 !]: <description>`；新增功能使用 `feat`，修复问题使用 `fix`，其他改动按实际性质使用 `refactor`、`docs`、`test`、`chore` 等类型。
- 描述应简明说明实际改动；需要正文时，与标题之间空一行，补充原因、行为变化或验证结果。
- 破坏性变更必须通过标题中的 `!` 或脚注 `BREAKING CHANGE:` 明确标记。
