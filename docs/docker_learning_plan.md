# 已完结
# Docker 基础学习计划：结合 Daily Meal

制定日期：2026-10-13。
最近更新：2026-10-16。

课程状态：已结课（2026-10-16）。结合此前的概念理解反馈、学习者自行编写和修正配置、完整构建输出以及运行与接口实测，本轮 Docker 基础学习目标已达到。各阶段保留的“未独立复述”“未独立复测”等说明用于如实记录证据，不是追加考试或继续补课的要求。多阶段构建、缓存优化、镜像发布和底层原理均为可选进阶，后续按实际需要学习。

目标：通过 Daily Meal 的 Go 后端，熟悉 Dockerfile、常用命令、网络与环境变量、存储、日志排错和 Docker Compose，同时理解相关概念，能够解释命令与概念之间的对应关系。

本文同时记录学习计划与实际进度。实操记录区分原始命令输出、学习者口头反馈和助手实测，概念掌握情况单独记录。介绍新知识时，先用完整的文字按“是什么 → 为什么 → 怎么做”讲清概念、工作机制、因果关系和使用方式，再给出围绕一个完整任务的成组步骤。实操用于验证已经讲清的知识；学习者可以连续操作，遇到问题时再停下来讨论，不要求每条命令执行后都回来确认。

## 学习与讲解约束

以下约束适用于所有阶段，以及后续结合本计划开展的讲解和练习：

- 必须同时学习概念和命令。能够复制命令并运行成功，或把项目启动起来，不能单独作为学会的依据。
- 新知识必须先文字讲解、后安排实操，顺序为“是什么 → 为什么 → 怎么做”：说明概念及相关对象是什么、解决什么问题、为什么这样工作、与已学知识的关系，以及具体怎样使用。涉及行为变化时，在执行前讲清预期结果及其原因；不能用“运行一下再观察”“先猜结果”或让学习者自行归纳来代替讲解，也不能把关键原理留到实操之后才解释。
- 面向有 Ruby 后端经验、能够阅读技术文字的成年开发者讲解。使用清楚、直接、有足够信息量的文字；不把提问视为理解能力不足，不使用幼儿化的引导、反复确认或逐条考问。可以用具体例子和图表辅助，但不能用类比、简短定义或操作现象代替必要的原理与因果说明。
- 根据前文沟通和已经讲过的内容调整信息量，直接承接已有知识；每次可以给出较长的完整样例和连续步骤，按一个完整任务推进，不机械地拆成“一条命令 → 等反馈”。新概念可在整组步骤之前或对应步骤的命令之前讲清，讲解后学习者即可继续操作，不必每个概念都等待确认或复述。
- 按本计划的阶段推进，结合对应的 Docker 官方概念文档和命令参考学习。讲解时应指出相关链接或条目，并用简单语言解释，再对应到 Daily Meal 的实际代码和练习结果；不能只提供命令或文档链接。
- 使用一个命令前，先解释它操作的对象、主要参数、预期行为，以及它与本阶段概念的关系。例如，学习 `docker run` 时，需要解释镜像与容器的关系，而不只是记住启动语法。
- 首次引入镜像、工具、配置项或路径时，先交代它的来源、作用及何时创建或生效，区分 Docker 自身的规则、镜像提供者的约定和本项目的选择；不要默认学习者已经知道未讲过的背景。例如，应在使用 PostgreSQL 镜像前说明它由维护者预先构建，使用时如何获取，数据库又在什么时机初始化。
- 每个关键概念在文字讲清后安排具体操作验证。执行前由助手说明预期结果和依据，执行后将实际状态、日志或文件变化与说明对照；如有差异，再根据证据分析原因。观察结果用于验证和排错，不承担首次介绍概念的任务。
- 学习者明确表示某处不清楚时，针对该问题补充文字解释和具体例子；必要时再缩小例子或安排验证。不因一个问题就把整个课程退回逐条命令确认，也不通过堆叠术语或连续执行更多命令来跳过解释。
- 排错教学围绕可复用的方法展开。学习者已经理解状态、日志、实际配置和修复验证之间的关系后，不再为每一种错误重复安排完整实验；故障清单作为按需参考，不按案例数量验收。后续优先在完整任务遇到的实际问题中运用方法，只有出现新的概念或明确的理解缺口时才补充针对性练习。
- 文档中的“要理解”和“完成标准”共同构成学习要求。只有操作完成、概念能够说明、结果能够解释时，才将对应阶段标记为完成。
- 完成标准用于记录掌握情况，不作为每一步都必须答题、复述或提交输出的门槛。可结合自然对话、整组操作反馈和阶段回顾确认；未单独验证的能力如实保留，不据此机械中断后续讲解。
- 每完成一个阶段，助手必须在当轮主动回写本学习文档，并在进入下一阶段前完成更新，不等待用户再次提醒。回写至少包括完成日期、执行的命令、实际结果、概念复述情况及尚未验证的内容，并同步更新“当前进度”、学习顺序表、该阶段的记录与完成标准，以及文末的“当前下一步”。
- 阶段内只有部分实操或复述完成时，也应如实记录已有进展，保留未满足的完成标准；不得把预期结果、助手的讲解或尚未执行的下一步记成学习者已经完成。

每个阶段都需要满足以下共同完成标准：

- [ ] 能用自己的话解释本阶段的核心概念，并区分容易混淆的概念。
- [ ] 能独立选择并使用相关命令，解释主要参数的含义。
- [ ] 能说明命令改变了什么、为什么产生观察到的结果，并预测改变一个条件后的结果。
- [ ] 能结合官方文档和当前项目，指出解释与实际行为的对应依据。

## 当前进度

截至 2026-10-15，第 0～3 阶段的主要实操已经完成：Docker 命令可连接 Engine，本机编译得到 API 和迁移程序，教学镜像 `daily-meal:learn` 已实际构建并能在镜像列表中查到；已用该镜像创建 API 容器，验证缺少 `DATABASE_URL` 时退出，并查看了容器状态、日志和退出码。

已经讨论了编译入口、Dockerfile 各行、构建上下文、运行时环境变量和容器可写层。学习者已反馈看懂编译入口和 Dockerfile；完整的独立复述与从头编写尚未验收，因此下方保留相应的未勾选项。

第 3 阶段已由学习者复述缺少应用配置导致退出，以及退出后容器仍然保留。列表之间的区别、删除容器与删除镜像的完整边界仍待独立复述，因此尚不将整个阶段标记为全部验收完成。

2026-10-15，学习者确认“创建网络 → 启动 PostgreSQL → 执行迁移 → 启动 API → 验证接口”整组流程已走完且没有问题。第 4 阶段主要实操及第 5 阶段首次端口映射、接口访问按该口头反馈记录为完成；本次未提供逐条命令输出，助手未代为执行或独立复测。概念已讲解，完整的独立复述仍待结合后续操作复核。

2026-10-15，学习者完成第 5 阶段双实例与生命周期整组练习，并反馈“完成了，我理解相关知识点了”。本阶段按学习者对实操完成和概念理解的确认记为完成；未收到原始命令输出或逐项独立复述，助手未代为执行或独立复测。

2026-10-15，第 6 阶段已完成。学习者先复述可写层的理解，并在补充讲解写时复制、权限边界及挂载路径后反馈“ok我都理解了”，随后确认“这三部我也完成了”，对应普通文件生命周期、命名 volume 数据库重建和 bind mount 三组实验。本阶段按概念理解与实操完成的整体口头反馈记录完成；助手只独立检查过实验前的容器与挂载，未独立复测三组实验，未收到原始实验输出。

2026-10-15，第 7 阶段开始。助手准备了 `compose.learn.yaml`，明确使用教学 Dockerfile、独立项目 `dailymeal-learn`、API 端口 `18084` 和新卷 `dailymeal-learn_pgdata`，并实际通过 `docker compose -p dailymeal-learn -f compose.learn.yaml config` 检查。启动前只读检查时，原手动 API 占用 `18080`、`18082`、`18083`，未发现 `18084` 监听进程，学习项目的网络和卷尚不存在。当时助手未执行 Compose 构建、启动、接口写入或重建实验，配置检查不代表运行验证。

2026-10-16，第 7 阶段已完成。学习者确认“已经完成了，相关概念也理解了”，对应上一轮的 Compose 配置检查与启动、状态和日志检查、接口访问与食物写入，以及保留命名卷后删除重建并查询数据的整组练习。概念理解和实操结果按这次整体口头反馈记录；未收到原始命令输出、实际食物 ID 或逐项独立复述，助手未独立复测。独立从头编写 Dockerfile 与 Compose 文件仍留到第 8 阶段综合复习。

2026-10-16，第 8 阶段开始，首组为缺少 `DATABASE_URL` 的定位与修复。练习前助手只读确认学习数据库健康、迁移容器成功退出、API 运行，默认网络存在，教学镜像默认执行 `/app/bin/api` 且未预设 `DATABASE_URL`；未发现 `18085` 端口监听。本次状态检查不等于重新验证第 7 阶段的数据保留实验。

同日，学习者对首组练习反馈“好了”，按整组实操完成记录，涵盖制造故障、状态与日志检查、核对变量、再次启动及补齐变量后重建和接口验证；未收到原始输出，概念掌握未单独确认。随后助手在当前 `desktop-linux` context 只读检查发现容器列表为空，`dailymeal-learn_default` 网络不存在，但 `daily-meal:learn` 镜像和 `dailymeal-learn_pgdata` 卷仍在，因此未能复查上组故障日志或修复后的运行状态。下一组先恢复 Compose 学习环境，再练习数据库主机名写错；尚未执行。

此后，学习者提供恢复后的网络检查输出，并确认“现在我对你上面提出的流程都了解了”，据此记录前两组排错流程及服务名、容器名、网络别名的区别已理解。助手只读复核 `meal-learn-debug-db` 正在运行并映射 `18085`，日志含 `listening`，请求该端口的 `/healthz` 实际返回 `{"status":"ok"}`；Compose 数据库健康、迁移成功退出、API 运行。主机名写错时的原始日志未提供，不补写故障输出；修复后的当前运行与健康接口已实测。随后曾准备数据库密码认证失败练习，尚未执行。

2026-10-16，学习者指出本阶段变成穷举，没有必要每一种情况都操作一遍。按此反馈调整：结束逐个制造故障的安排，密码错误、端口占用等保留为按需参考，取消至少定位三类故障的数量要求。下一步直接进行独立编写 Dockerfile 与 Compose 的综合练习，实际遇到问题时运用已有排查方法。跳过的案例不记为已经执行，整个第 8 阶段仍待综合练习完成。

同日，学习者提交 `backend/Dockerfile.practice` 和 `compose.practice.yaml` 初稿供审阅。配置已包含依赖下载、两个程序的编译、三个服务、启动条件、数据库连接和命名卷。助手检查发现 `FROM` 镜像标签分隔、`go build` 拼写、`CMD` 引号与路径、Compose 的 `depends_on` 和 `ports` 字段需要修正；当前主机端口 `18084` 已有监听。已执行 Compose 配置校验和 Dockerfile 构建规则检查，均在初稿错误处失败；尚未实际构建或启动本次练习服务。下一步由学习者修正后继续验证，第 8 阶段仍进行中。

随后，学习者修正上述写法并将主机端口改为 `18086`，提供 Compose 校验和实际构建输出：配置校验未报错，构建在获取 `golang:1.25.0` 元数据时因连接 `auth.docker.io` 超时而失败。助手复核配置校验通过并重现该超时；显式使用系统已有的 `127.0.0.1:7890` 代理访问令牌接口返回 HTTP 200，再为单次 `docker build --check` 指定该代理后，基础镜像元数据获取成功。检查仍因指令大小写警告返回非零，完整构建与本次练习服务启动尚未验证；下一步统一指令大写并带临时代理运行构建。

同日，学习者提供带临时代理执行 Compose 构建的完整输出，并反馈“现在好了”：构建耗时 `250.2s`，基础镜像下载、依赖下载、API 与迁移程序编译以及镜像导出均完成，最终显示 `Image daily_meal_practice Built`。大小写警告仍存在，但没有阻止这次实际构建。助手只读确认 `daily_meal_practice:latest` 与 `postgres:17-alpine` 均在本地，练习项目当前没有容器；构建结果已确认，下一步启动服务并验证接口。

同日，学习者对启动、状态与日志检查以及两个接口请求整组操作反馈“成功了”。助手随后只读复核：`db` 为运行且健康，`migrate` 为 `Exited (0)` 且日志确认 Goose 迁移到版本 3，`api` 运行并映射 `18086`；`/healthz` 返回 HTTP 200 和 `{"status":"ok"}`，`/api/v1/foods` 返回 HTTP 200 和空分页列表。第 8 阶段的综合实操已完成。概念理解沿用此前阶段的反馈，本轮没有新增独立复述，未验收的概念项继续保留；本次没有验证练习数据库的数据写入或删除重建后的持久化。后续可以进入多阶段构建学习，结合自然交流补齐概念复核，不重复已完成的实验。

结课确认（2026-10-16）：根据学习者对课程边界的提醒，本轮以基础知识学习和综合任务完成为止，现已结课。多阶段构建等内容改为以后按需学习的可选进阶，当前不追加练习或考核。

## 学习顺序

| 顺序 | 学习内容 | 本阶段的具体结果 | 状态 |
| --- | --- | --- | --- |
| 0 | 准备 Docker 环境 | 确认命令行能连接 Docker Engine | 实操完成，待独立解释 |
| 1 | 看懂项目怎样编译和启动 | 找到 API、迁移程序与环境变量的入口 | 本机编译完成，待综合复述 |
| 2 | 手写简单 Dockerfile，构建镜像 | 得到一份包含后端程序的教学镜像 | 镜像构建完成，待独立说明 |
| 3 | 创建容器，观察退出和日志 | 解释缺少配置时容器为什么退出 | 实操完成，已复述退出原因，部分概念待复核 |
| 4 | 环境变量与容器网络 | 后端能通过网络连接学习数据库并执行迁移 | 整组实操完成（口头反馈），待独立说明 |
| 5 | 端口映射与容器生命周期 | 从电脑访问 API，能独立启动两份 API | 已完成（2026-10-15 学习者确认实操完成并理解知识点；未独立复测） |
| 6 | Volume 与目录挂载 | 区分容器文件和持久数据，验证数据库数据保留 | 已完成（2026-10-15 学习者确认理解概念并完成三组实验；实验未独立复测） |
| 7 | 手写 Docker Compose | 用一份配置管理数据库、迁移和 API | 已完成（2026-10-16 学习者确认整组实操完成并理解相关概念；运行实验未独立复测） |
| 8 | 综合排错与复习 | 独立编写配置并在实际问题中运用排查方法 | 已完成，基础课程结课（2026-10-16；结合此前理解反馈、自行编写配置、构建输出和运行实测综合确认，单项证据边界保留） |

日志排错从第 3 阶段开始使用，后续每个阶段都继续练习，最后再集中复习。

## 练习约定

- 以 Go 源码为依据，从简单示例开始，不把现有 Dockerfile、Compose 文件或 Docker 启动文档作为课程起点。
- 所有终端命令默认从仓库根目录执行。
- 第 2 阶段已创建 `backend/Dockerfile.learn` 和对应的忽略文件；学习者先写了基础镜像行，随后按其请求由助手补全并逐行讲解。第 7 阶段已由助手准备根目录的 `compose.learn.yaml`。第 8 阶段学习者已自行提交 `backend/Dockerfile.practice` 和 `compose.practice.yaml` 初稿，经审阅后自行修正，完成构建、启动和接口验证。
- 教学镜像统一命名为 `daily-meal:learn`；手动练习的容器、网络和 volume 使用 `meal-learn-` 前缀。
- 数据库练习使用单独的 `dailymeal_docker_learn` 数据库和可丢弃的学习数据，不连接个人饮食数据库。
- 文件有变化时先说明影响：修改源码后需要重新构建镜像；修改运行时配置通常需要重新创建容器。

## 0. 准备 Docker 环境

要理解：终端里的 Docker 命令负责发出请求，Docker Engine 负责实际构建镜像和管理容器。

练习：

1. 在 macOS 上安装并启动 Docker Desktop。
2. 执行以下命令，查看输出。

```sh
docker version
docker compose version
```

观察：`docker version` 应能显示 Client 和 Server 信息；只有客户端版本或无法连接 Server，不能说明运行环境已经准备好。

已完成记录（2026-10-13）：

- [x] Docker Desktop 已启动，界面显示 Engine running。
- [x] `docker version` 同时显示 Client 和 Server：Client/Engine 为 `29.8.2`，Docker Desktop 为 `4.95.0`；客户端为 `darwin/arm64`，Engine 为 `linux/arm64`，Context 为 `desktop-linux`。
- [x] `docker compose version` 返回 `v5.6.0`。
- [x] 新开的终端也能使用 Docker 命令。

最初出现 `zsh: command not found: docker`，临时执行 `export PATH="$HOME/.docker/bin:$PATH"` 后恢复。检查发现 `.zprofile` 已有 Docker CLI 路径配置，之后学习者确认新终端正常。这次问题是终端查找命令的路径，不是镜像或业务代码的问题。

完成标准：

- [ ] 能说明 Docker Desktop、命令行和 Docker Engine 分别做什么。
- [x] 两条版本命令正常返回。

参考：[Docker Desktop 的 macOS 安装说明](https://docs.docker.com/desktop/setup/install/mac-install/)。

## 1. 看懂项目怎样编译和启动

先阅读这些实际源码：

| 文件 | 只需要关注什么 |
| --- | --- |
| [backend/go.mod](../backend/go.mod) | Go 版本和项目依赖；当前要求 Go 1.25 |
| [backend/cmd/api/main.go](../backend/cmd/api/main.go) | 读取配置、连接数据库、启动 HTTP 服务 |
| [backend/cmd/migrate/main.go](../backend/cmd/migrate/main.go) | 读取配置、连接数据库、执行迁移 |
| [backend/internal/config/config.go](../backend/internal/config/config.go) | `DATABASE_URL` 必填，`HTTP_ADDR` 默认 `:8080` |
| [backend/migrations/migrations.go](../backend/migrations/migrations.go) | SQL 迁移文件通过 `go:embed` 嵌入程序 |
| [Makefile](../Makefile) | `make build` 怎样编译这两个程序 |

练习：执行 `make build`，观察生成的 `backend/bin/api` 和 `backend/bin/migrate`。

观察：这一步只是在本机编译 Go 程序，没有制作 Docker 镜像，也没有执行数据库迁移。

已完成记录（2026-10-13）：

```text
make build
cd backend && go build -o bin/api ./cmd/api && go build -o bin/migrate ./cmd/migrate
ls backend/bin
api     migrate
```

- [x] 本机编译成功，生成了两个可执行程序。
- 已讲解：`cmd` 是项目组织可执行程序入口的目录惯例；`./cmd/api` 指定入口包，Go 会沿导入关系编译它依赖的其他包，不是只编译 `main.go`，也不是无条件编译整个 `backend`。
- 已讲解：先 `cd backend`，所以 `-o bin/api` 指定的是 `backend/bin/api` 输出文件，`./cmd/api` 指的是 `backend/cmd/api` 输入入口目录。
- 已讲解：API 启动需要 `DATABASE_URL`，`HTTP_ADDR` 默认是 `:8080`。阅读入口是为了对应镜像中要编译哪个程序、启动哪个程序及传入哪些配置。
- 学习者已反馈看懂编译入口；本阶段没有运行 API，也没有执行迁移。

完成标准：

- [ ] 能指出两个程序的入口，并解释 API 和迁移为什么分开执行。
- [ ] 能说明程序启动需要哪些配置，以及配置缺失时会发生什么。
- [ ] 能区分 Go 源码、编译后的程序和 Docker 镜像。

## 2. 手写简单 Dockerfile，构建镜像

要理解：Dockerfile 是制作步骤；构建后得到的镜像已经包含真实文件和启动配置。

练习：

1. 新建 `backend/Dockerfile.learn`，从包含 Go 1.25 的基础镜像开始。
2. 逐行加入工作目录、复制源码、下载依赖、编译 API 和迁移程序的步骤。
3. 记录默认启动命令，让容器启动时执行 API。
4. 第一版使用单阶段构建，保留源码和 Go 工具，方便观察。
5. Docker 的构建上下文需要排除 `.env`、本机构建产物等文件。为教学文件编写对应的 `backend/Dockerfile.learn.dockerignore`，并理解它和 `.gitignore` 的区别。
6. 文件写好以后，再执行构建命令。

```sh
docker build -f backend/Dockerfile.learn -t daily-meal:learn backend
docker image ls daily-meal
```

本阶段逐个学习 `FROM`、`WORKDIR`、`COPY`、`RUN` 和 `CMD`。重点区分：`RUN` 在构建时执行，`CMD` 记录将来启动容器时的默认命令。

观察：构建完成后能查到镜像，但 API 尚未启动。构建过程不需要连接业务数据库。

已完成记录：

- [x] 2026-10-13：完成 [Dockerfile.learn](../backend/Dockerfile.learn) 和 [Dockerfile.learn.dockerignore](../backend/Dockerfile.learn.dockerignore)，逐行讲解；学习者在实际构建前反馈看懂该文件。
- [x] 2026-10-13：助手执行 `docker build --check --progress=plain -f backend/Dockerfile.learn backend`，返回 `Check complete, no warnings found.`。这只是构建检查，当时没有执行真正的镜像构建。
- [x] 2026-10-14：学习者执行上面的 `docker build`，输出 `Building 107.9s (12/12) FINISHED`，真正完成依赖下载、容器内编译和镜像导出。
- [x] 2026-10-14：学习者执行 `docker image ls daily-meal`，确认本地镜像存在。

首次构建输出中的几个观察点：

| 项目 | 本次结果 | 含义 |
| --- | --- | --- |
| 构建上下文传输 | `218.46kB` | 本次传给构建器的上下文内容大小，不是最终镜像大小 |
| 基础镜像步骤 | `55.4s` | 获取并准备 `golang:1.25` 基础镜像 |
| `go mod download` | `27.6s` | 下载项目的 Go 依赖 |
| 编译 API / migrate | `9.7s` / `0.5s` | 在镜像构建环境中生成两个 Linux 程序 |
| 导出镜像 | `9.9s` | 保存镜像层及相关元数据等 |

以上是本次构建快照，不代表每次构建都会耗时相同，也不是所有内部操作都必须顺序相加。

本次镜像列表：

| IMAGE | ID | DISK USAGE | CONTENT SIZE |
| --- | --- | --- | --- |
| `daily-meal:learn` | `50003547225c` | `1.97GB` | `449MB` |

`daily-meal` 是镜像名，`learn` 是标签。当前镜像存储方式同时保留压缩内容与解包后的文件；`CONTENT SIZE` 表示内容大小，`DISK USAGE` 包括内容和解包后的存储。这些数字不是 API 可执行文件的大小，共享的镜像层也不能按各镜像显示值简单累加。本版是单阶段构建，包含 Go 工具链、源码、依赖和编译产物，之后再学习缩小镜像。

输出中的 `docker.io/library/daily-meal:learn` 是镜像的完整命名形式；本次只构建本地镜像，没有执行 `docker push`。镜像列表顶部的 `U In Use` 是标记说明，本行 `EXTRA` 为空，不能由表头推断 API 正在运行。

已讲解、后续还要用操作验证的概念：

- `-f backend/Dockerfile.learn` 选择构建文件，`-t daily-meal:learn` 指定镜像名和标签；末尾 `backend` 是相对仓库根目录的构建上下文路径。
- `COPY` 的源路径以构建上下文为基准；目标相对路径以 `WORKDIR` 为基准。因此 `COPY . .` 把参与构建的后端文件复制到 `/app`，不会把电脑上整个仓库原样变成镜像的根目录。
- `WORKDIR /app` 设置后续指令和默认启动进程的工作目录，不是规定只有 `/app` 可写。镜像还包含基础镜像的目录，容器的可写层也不局限于 `/app`；写入仍受权限和挂载配置影响。
- `RUN` 在构建时执行；`CMD ["/app/bin/api"]` 保存默认启动命令。本次构建没有执行这个 `CMD`，所以构建成功不等于 API 或数据库连接已经成功。
- 当前镜像包含 `/app/bin/api` 和 `/app/bin/migrate`，由 Dockerfile 和成功构建输出判断；尚未在容器里用 `pwd`、`ls` 实际查看目录。
- 教学忽略文件排除本机 `bin`、覆盖率文件、环境变量文件和 `.DS_Store`；它控制构建上下文，作用与 `.gitignore` 不同。

完成标准：

- [ ] 能解释 Dockerfile 的每一行，不依赖照抄。
- [ ] 能说明命令里的 `-f`、`-t` 和末尾 `backend` 各指什么。
- [ ] 能指出镜像里有哪些实际文件，以及默认启动哪个程序。

参考：[Dockerfile](https://docs.docker.com/build/concepts/dockerfile/)、[Dockerfile 指令参考](https://docs.docker.com/reference/dockerfile/)、[构建上下文与忽略文件](https://docs.docker.com/build/concepts/context/)、[镜像](https://docs.docker.com/get-started/docker-concepts/the-basics/what-is-an-image/)、[构建命令](https://docs.docker.com/reference/cli/docker/buildx/build/)、[镜像列表命令](https://docs.docker.com/reference/cli/docker/image/ls/)、[containerd 镜像存储占用](https://docs.docker.com/engine/storage/containerd/#disk-space-usage)。

## 3. 创建容器，观察退出和日志

要理解：创建和启动容器是另一个步骤；镜像构建成功，不代表程序启动一定成功。

当前状态：2026-10-14 已完成下方四条命令的实操反馈，并复述了退出原因与容器保留的现象；列表区分和删除操作的完整边界仍待独立复述。

练习：第一次故意不传 `DATABASE_URL`，使用第 2 阶段的镜像启动 API。

```sh
docker run --name meal-learn-api-missing-env daily-meal:learn
docker ps -a
docker logs meal-learn-api-missing-env
docker inspect meal-learn-api-missing-env
```

预期：根据当前配置代码，API 会报告 `DATABASE_URL is required` 并退出。这是本阶段要观察的现象，不应通过修改业务代码来绕过。

观察：镜像仍然存在，已退出的容器也仍然存在；日志能帮助解释退出原因。`docker exec` 需要容器处于运行状态，不能用它进入已经退出的 API 容器。

已完成记录（2026-10-14）：

| 操作 | 实际结果与依据 |
| --- | --- |
| `docker run --name meal-learn-api-missing-env daily-meal:learn` | 学习者贴出日志：`2026/10/14 07:23:17 ERROR server stopped error="DATABASE_URL is required"` |
| `docker ps -a` | 学习者贴出容器 `063b503fa786`，镜像为 `daily-meal:learn`，启动命令为 `/app/bin/api`，状态为 `Exited (1)`，名称为 `meal-learn-api-missing-env` |
| `docker logs meal-learn-api-missing-env` | 学习者贴出与首次运行内容、时间均相同的错误日志，验证容器停止后仍能读取之前的日志 |
| `docker inspect meal-learn-api-missing-env` | 学习者确认 `State` 与预期一致：`Status: exited`、`Running: false`、`ExitCode: 1`；未粘贴完整 JSON，依据为口头反馈 |

概念讲解与复述：

- 已结合实际代码说明：`DATABASE_URL` 必填由 `backend/internal/config/config.go` 中的 Go 校验逻辑规定；镜像构建时编译这段逻辑，启动 API 时才读取环境变量并执行检查，具体连接地址在运行时传入。
- 学习者已独立解释：没有传递应用运行所需的信息，应用启动时报错，导致容器退出。
- 学习者已独立说明：这次发生的是容器退出，容器仍然存在。已补充本次没有使用 `--rm`，因此不会在退出后自动删除容器。
- 学习者提出“删除镜像必须没有对应的容器”，助手已补充标签和强制删除的边界；这一补充尚未由学习者再次独立复述，不将其记为已经完全掌握。
- 已讲解 `docker image ls`、`docker ps`、`docker ps -a` 的区别；本阶段实际反馈了 `docker ps -a`，三者的独立选择和复述仍待验证。删除容器和删除镜像尚未实操。

完成标准：

- [ ] 能区分镜像列表、正在运行的容器列表和全部容器列表。
- [x] 能在日志中找到错误，并在容器状态中找到退出信息。
- [ ] 能解释“程序退出”“删除容器”“删除镜像”是三件不同的事。

参考：[容器](https://docs.docker.com/get-started/docker-concepts/the-basics/what-is-a-container/)、[docker run](https://docs.docker.com/reference/cli/docker/container/run/)。

## 4. 环境变量与容器网络

要理解：环境变量配置程序的运行行为；两个容器通过网络连接，不需要把数据库程序放进 API 镜像。

提前讨论过的概念：只读镜像不妨碍启动时通过 `-e` 或 `--env-file` 给容器进程设置环境变量。使用 `--env-file` 时，Docker CLI 读取指定文件并传入变量；当前 Go 程序通过 `os.Getenv` 读取变量，不会自动加载 `.env`。这些配置不需要改写镜像。本轮已通过 `-e` 实操；`--env-file` 尚未实操。

当前状态（2026-10-15）：学习者确认整组网络、数据库、迁移及 API 访问流程已完成且无问题，依据为口头反馈。保留概念复述的未勾选项，不把整体实操反馈等同于所有概念已独立掌握。

练习：

1. 创建名为 `meal-learn-network` 的自定义 bridge 网络。
2. 用 PostgreSQL 官方镜像启动独立的学习数据库容器 `meal-learn-db`，加入这个网络。
3. 明确设置学习数据库名、用户名和密码，检查数据库是否已经能接受连接。
4. 用 `daily-meal:learn` 镜像创建一次性迁移容器，同样加入该网络，并通过 `-e DATABASE_URL=...` 传入连接地址。
5. 覆盖默认启动命令，执行迁移程序的 `up` 命令。
6. 迁移成功后，再启动 API 容器并传入相同的数据库连接配置。

连接地址中的数据库主机使用 `meal-learn-db`，数据库端口使用容器内部的 `5432`。在 API 容器里，`localhost` 指向 API 容器自己的网络空间。

自定义 bridge 网络支持容器按名称互相访问；容器之间通过内部端口通信，不要求先把数据库端口发布到电脑上。[官方网络说明](https://docs.docker.com/engine/network/drivers/bridge/)

本阶段学习 `docker network create`、`docker network inspect`、`docker run -e` 和 `docker exec`。这里先用可丢弃的数据库数据，第 6 阶段再明确设置持久存储。

已完成记录（2026-10-15）：

学习者针对前一轮整组步骤反馈“这个流程走完了，没问题”。下面保存该组步骤使用的命令和配置；未收到原始终端输出，不补写网络 ID、容器 ID、具体迁移日志或 HTTP 响应原文。

```sh
docker network create meal-learn-network

docker run -d \
  --name meal-learn-db \
  --network meal-learn-network \
  -e POSTGRES_USER=meal \
  -e POSTGRES_PASSWORD=meal \
  -e POSTGRES_DB=dailymeal_docker_learn \
  postgres:17-alpine

docker exec meal-learn-db \
  pg_isready -h 127.0.0.1 -U meal -d dailymeal_docker_learn

docker run \
  --name meal-learn-migrate \
  --network meal-learn-network \
  -e 'DATABASE_URL=postgres://meal:meal@meal-learn-db:5432/dailymeal_docker_learn?sslmode=disable' \
  daily-meal:learn \
  /app/bin/migrate up

docker run -d \
  --name meal-learn-api \
  --network meal-learn-network \
  -e 'DATABASE_URL=postgres://meal:meal@meal-learn-db:5432/dailymeal_docker_learn?sslmode=disable' \
  -p 127.0.0.1:18080:8080 \
  daily-meal:learn

docker ps -a
docker logs meal-learn-api
curl -i http://127.0.0.1:18080/healthz
curl -i http://127.0.0.1:18080/api/v1/foods
```

| 验证项 | 记录结果与依据 |
| --- | --- |
| 网络和学习数据库 | 学习者整体确认已完成创建网络、启动数据库和就绪检查；未贴具体输出 |
| 跨容器迁移 | 学习者整体确认迁移正常完成；该步骤使用数据库容器名和内部端口连接 |
| API 运行与电脑访问 | 学习者整体确认 API 启动、状态检查及两个接口访问正常；未贴状态码和响应体原文 |

概念讲解与反馈：

- 已按“是什么 → 为什么 → 怎么做”讲解 bridge 网络、容器名称解析、各容器自己的 `localhost`，以及网络连接与 `DATABASE_URL` 配置的不同作用。
- 已对应 `config.go` 说明 `docker run -e` 设置运行时环境变量，Go 通过 `os.Getenv("DATABASE_URL")` 读取；学习者反馈“了解”，完整连接地址的独立解释尚未验收。
- 已说明 `POSTGRES_*` 是 PostgreSQL 官方镜像的初始化约定，`DATABASE_URL` 是本项目的配置约定；同一个 `-e` 参数负责传入这些不同用途的变量。
- 已说明 `daily-meal:learn` 由本项目构建，而 `postgres:17-alpine` 由官方镜像维护者预先构建。`docker run` 默认在本地缺少镜像时下载；数据库在容器首次初始化空数据目录时创建。本轮没有收到下载日志，不推断镜像是否原已缓存。
- 已说明镜像名后的 `/app/bin/migrate up` 覆盖当前教学镜像的默认 `CMD`，以及迁移正常结束与 API 错误退出的区别；独立复述仍待复核。
- `docker network inspect` 此前已介绍，但本次获确认的整组步骤没有包含它，因此不据此标记该命令已实操。数据库挂载检查、持久化和 Compose 仍待后续阶段。

参考：[docker run](https://docs.docker.com/reference/cli/docker/container/run/)、[PostgreSQL 官方镜像](https://hub.docker.com/_/postgres)、[pg_isready](https://www.postgresql.org/docs/17/app-pg-isready.html)、[docker exec](https://docs.docker.com/reference/cli/docker/container/exec/)。

完成标准：

- [ ] 能解释 `DATABASE_URL` 中每个部分对应什么。
- [ ] 能解释 API 容器为什么不能通过自己的 `localhost` 找到数据库容器。
- [x] 迁移容器正常完成；API 容器能够持续运行（2026-10-15 学习者整体口头确认）。
- [ ] 能区分“迁移成功后退出”和“API 因错误退出”。

## 5. 端口映射与容器生命周期

要理解：容器内部监听的端口，与电脑上访问它使用的端口，是两个位置。

当前状态（2026-10-15）：已完成。第一份 API 已随第 4 阶段整组实操使用 `127.0.0.1:18080:8080`，学习者确认电脑访问 `/healthz` 与 `/api/v1/foods` 正常；本轮又确认完成双实例、停止/启动和更换端口重建的整组练习，并理解相关知识点。完成依据为学习者反馈，未收到原始输出，未进行助手独立复测。

练习：

1. 为 API 配置 `127.0.0.1:18080:8080`，通过 `curl http://127.0.0.1:18080/healthz` 验证接口（已获整体完成反馈）。
2. 使用 `docker logs` 观察请求相关的输出，使用 `docker inspect` 查看端口和启动配置。
3. 先用文字讲清 `docker stop`、`docker start` 和 `docker rm` 分别改变什么、保留什么及其原因，再对运行中的 API 执行这些命令，验证已说明的行为。
4. 用同一镜像启动第二个 API 容器 `meal-learn-api-2`，内部仍监听 `8080`，电脑端口使用 `18081`，分别访问两份 API。
5. 对第二个 API 练习停止/启动，再停止并删除它；使用相同镜像和名称重新创建，电脑端口改为 `18082`，比较容器 ID、端口及第一份 API 的可用性。

两个 API 可以使用同一数据库，因此会看到同一份业务数据。这是访问共享数据库的结果，不能用来推断它们共享各自写入的容器文件。

已完成记录（2026-10-15）：

- [x] 首次端口映射及从电脑访问 API：学习者对第 4 阶段整组流程口头确认完成；使用的命令记录在上一节。
- [x] 双实例与生命周期：学习者针对下方三组操作反馈“完成了，我理解相关知识点了”。按整体反馈记录完成，不补写具体容器 ID、镜像 ID、状态码或响应体原文。
- 已讲解 `127.0.0.1:18080:8080` 分别对应电脑监听地址、电脑端口和容器端口；同一 bridge 网络中的两个 API 容器各有自己的网络空间，内部均可监听 `8080`，电脑入口分别使用不同端口。
- 已讲解同一镜像可以创建多个容器，各自拥有独立的进程、容器 ID 和可写层；本次没有要求重新构建镜像。
- 已讲解 `stop` 停止进程但保留容器，`start` 沿用原容器与配置重新执行启动命令；`rm` 删除容器及其可写层，镜像仍保留；再次 `run` 创建新容器，使用新的容器 ID 和本次指定的配置。
- 已讲解更换端口映射需重新创建容器；修改 Go 源码后需重新构建镜像并替换容器，启动旧容器不会自动加载电脑上的源码修改。
- 已说明 `/healthz` 检查数据库连接，`/api/v1/foods` 进一步访问业务表；未把这两个接口的完成反馈扩展为完整业务验收或持久化验证。
- 概念掌握按学习者“理解相关知识点”的整体确认记录；本轮没有逐项独立复述或从零选取命令的验收。文件保留、数据库挂载和持久化实验留到第 6 阶段。

本轮获确认完成的命令如下；`docker logs meal-learn-api-2` 仅作为启动失败时的排错建议，不据此记录本轮已执行该命令。

第一组：创建第二份 API，访问两个入口并比较容器、镜像身份。

```sh
docker run -d \
  --name meal-learn-api-2 \
  --network meal-learn-network \
  -e 'DATABASE_URL=postgres://meal:meal@meal-learn-db:5432/dailymeal_docker_learn?sslmode=disable' \
  -p 127.0.0.1:18081:8080 \
  daily-meal:learn

docker ps -a
curl -i http://127.0.0.1:18080/healthz
curl -i http://127.0.0.1:18081/healthz

docker inspect \
  --format 'name={{.Name}} container={{.Id}} image={{.Image}}' \
  meal-learn-api meal-learn-api-2
```

第二组：停止第二份 API，验证两个入口，再启动原容器。

```sh
docker stop meal-learn-api-2
docker ps -a
curl -i --max-time 3 http://127.0.0.1:18081/healthz
curl -i http://127.0.0.1:18080/healthz

docker start meal-learn-api-2
docker ps -a
curl -i http://127.0.0.1:18081/healthz
```

第三组：停止并删除第二份容器，改用电脑端口 `18082` 重新创建。

```sh
docker stop meal-learn-api-2
docker rm meal-learn-api-2

docker run -d \
  --name meal-learn-api-2 \
  --network meal-learn-network \
  -e 'DATABASE_URL=postgres://meal:meal@meal-learn-db:5432/dailymeal_docker_learn?sslmode=disable' \
  -p 127.0.0.1:18082:8080 \
  daily-meal:learn

docker ps -a
curl -i http://127.0.0.1:18082/healthz
curl -i http://127.0.0.1:18080/healthz
curl -i --max-time 3 http://127.0.0.1:18081/healthz
```

| 验证项 | 执行前说明的预期 | 实际反馈与证据边界 |
| --- | --- | --- |
| 双实例 | 两个入口可用，容器 ID 不同；期间未重建或改动镜像标签时镜像 ID 相同 | 学习者整体确认完成；未提供具体 ID 或 HTTP 输出 |
| 停止与启动 | 停止第二份 API 后 `18081` 不可用，`18080` 仍可用；启动后恢复 `18081`，容器 ID 不变 | 学习者整体确认完成；未提供逐步状态与响应 |
| 删除与重建 | 新容器 ID 改变；`18082` 与 `18080` 可用，旧入口 `18081` 不可用 | 学习者整体确认完成；未提供新旧 ID 或端口检查原文 |

参考：[运行容器](https://docs.docker.com/engine/containers/run/)、[端口映射](https://docs.docker.com/engine/network/port-publishing/)、[docker stop](https://docs.docker.com/reference/cli/docker/container/stop/)、[docker start](https://docs.docker.com/reference/cli/docker/container/start/)、[docker rm](https://docs.docker.com/reference/cli/docker/container/rm/)、[docker inspect](https://docs.docker.com/reference/cli/docker/inspect/)。

完成标准（2026-10-15 按学习者对实操完成及知识点理解的整体确认勾选，未逐项独立复述或复测）：

- [x] 能解释端口映射三个部分的含义。
- [x] 能使用 `18080` 和 `18081` 分别访问两份 API，并说明为何只需要一份后端镜像。
- [x] 能区分“容器处于运行状态”和“HTTP 接口及数据库连接正常”。
- [x] 能解释修改源码后为何需要重新构建镜像并替换容器。

## 6. Volume 与目录挂载

要理解：镜像提供初始文件；容器可以有自己的文件变化；需要长期保留的数据应明确保存到持久存储中。

当前状态（2026-10-15）：已完成。学习者确认理解可写层、写时复制、挂载路径与存储生命周期，并确认完成三组存储实验。按整体反馈记录完成，未收到原始命令输出，助手未独立复测实验。

概念承接：镜像层只读，容器有各自的可写层；`WORKDIR` 设置工作目录，不划定可写层的目录范围。本阶段用普通文件、数据库重建和 bind mount 三组实验验证这些概念。

实验前核对与概念记录（2026-10-15，以下容器状态为助手在学习者实验前的只读检查快照）：

- 执行 `docker ps -a`、`docker inspect`、`docker volume ls` 和 `docker image inspect` 的格式化查询，核对容器状态、`Mounts`、网络和镜像的 `Config.Volumes`。
- `meal-learn-db` 正在运行，`Mounts` 中为 `Type=volume`，随机卷名以 `2713a12c85d0` 开头，挂载目标为 `/var/lib/postgresql/data`；本地 `postgres:17-alpine` 镜像也声明了这一 volume 路径。现有数据库使用匿名 volume，不能作为“文件只在容器可写层”的示例。
- `meal-learn-api` 与 `meal-learn-api-2` 正在运行，电脑入口分别为 `127.0.0.1:18080` 与 `127.0.0.1:18082`，两者 `Mounts=[]`；本地 `daily-meal:learn` 镜像未声明 volume。
- 本次卷列表只有上述匿名卷，尚无 `meal-learn-pgdata`。`lsof -nP -iTCP:18083 -sTCP:LISTEN` 未列出监听进程；这是检查时的快照，不能保证后续端口始终空闲。
- 已准备命名 volume 实验：新建 `meal-learn-volume-db`，使用 `meal-learn-pgdata` 和独立的学习数据库 `dailymeal_docker_learn`；对应 API 使用 `meal-learn-volume-api` 和电脑端口 `18083`。这是一套新的空数据库，现有 `meal-learn-db` 的数据不会自动复制过去。
- 上述为实验前的环境核对和练习安排；助手没有创建、停止或删除容器，没有写入数据库或操作 volume 数据。学习者随后反馈的实验完成情况单独记录在下方。
- 学习者自行复述：修改镜像中的文件不会改动镜像原文件，而会在容器自己的可写层保存变化；可写层不是容器内某个固定目录。助手补充：同一容器的多项普通文件变化进入同一个可写层，是否能写仍受文件权限、只读配置与挂载影响。
- 已结合现有数据库挂载解释 `Source` 与 `Destination`：前者是 Docker 主机上的卷存储路径，后者是容器内的访问路径；在当前 Mac 环境中，Docker 主机位于 Docker Desktop 的 Linux 虚拟机内。挂载路径中的写入进入 volume，普通未挂载路径中的变化进入容器可写层。
- 学习者针对上述补充讲解反馈“ok我都理解了”。这一条反馈仅用于记录概念理解；三组实验完成另以之后的实操反馈为依据。

已完成记录（2026-10-15，学习者整体口头确认）：

学习者针对已给出的三组练习确认“这三部我也完成了”。下面记录获确认的操作与结果，不补写实际容器 ID、食物 ID、HTTP 响应原文或 `diff` 输出。

| 实验 | 获确认的命令与操作范围 | 结果与证据边界 |
| --- | --- | --- |
| 普通文件生命周期 | `docker exec meal-learn-api-2 sh -c 'echo container-only > /app/meal-learn-note.txt'` 写文件，随后用 `docker exec ... cat` 读取；对该容器执行 `docker stop`、`docker start`；再 `docker stop`、`docker rm`，用相同镜像、网络、数据库地址和 `127.0.0.1:18082:8080` 重新 `docker run -d`；读取临时文件并用 `ls -l /app/bin/api` 检查程序文件 | 学习者整体确认完成：停止/启动后普通文件保留，删除重建后消失，镜像中的 API 程序仍在；未提供逐条输出 |
| 命名 volume 数据库重建 | `docker volume create meal-learn-pgdata`；创建 `meal-learn-volume-db` 时使用 `--mount type=volume,source=meal-learn-pgdata,target=/var/lib/postgresql/data`；通过 `pg_isready` 检查就绪，临时迁移容器执行 `/app/bin/migrate up`；启动端口 `18083` 的 `meal-learn-volume-api`，使用 `curl` 创建并查询 `volume-demo` 食物；停止实验 API、停止并删除实验数据库容器，用 `docker volume inspect meal-learn-pgdata` 检查卷，再以同一卷重建数据库并启动 API；重新查询并用 `diff` 比较保存的响应 | 学习者整体确认完成：数据库容器重建后，学习记录仍然保留；未提供记录 ID、前后 JSON 或比较输出 |
| Bind mount 双向读写 | `mkdir -p /private/tmp/meal-learn-bind`；电脑用 `echo` 写入 `from-mac.txt`，临时容器通过 `--mount type=bind,source=/private/tmp/meal-learn-bind,target=/shared` 挂载并用 `cat /shared/from-mac.txt` 读取；另一临时容器用 `sh -c 'echo from-container > /shared/from-container.txt'` 写入，电脑用 `cat` 读取；两次临时容器均使用 `docker run --rm` | 学习者整体确认完成：电脑与容器可以访问挂载目录中的文件，临时容器删除后电脑上的文件仍在；未提供原始读取输出 |

数据库实验使用独立的 `meal-learn-volume-db`、`meal-learn-volume-api`、`meal-learn-pgdata` 和 `dailymeal_docker_learn`；API 入口为 `http://127.0.0.1:18083`。前后查询按练习分别保存到 `/private/tmp/meal-learn-volume-before.json` 与 `/private/tmp/meal-learn-volume-after.json`。以上名称和路径来自获确认完成的练习，助手没有读取这些结果文件或在实验结束后复查运行状态。

参考：[存储与可写层](https://docs.docker.com/engine/storage/)、[写时复制](https://docs.docker.com/engine/storage/drivers/#the-copy-on-write-cow-strategy)、[Volume](https://docs.docker.com/engine/storage/volumes/)、[Bind mount](https://docs.docker.com/engine/storage/bind-mounts/)、[PostgreSQL 官方镜像](https://hub.docker.com/_/postgres)、[docker exec](https://docs.docker.com/reference/cli/docker/container/exec/)。

练习：

1. 在运行中的教学容器里创建一个临时文件，观察停止、重新启动以及删除重建容器后它的变化。
2. 检查数据库容器的 `Mounts`，确认 PostgreSQL 数据目录实际挂载到了哪里，不猜测数据库文件一定在容器可写层里。
3. 为学习数据库明确配置命名 volume `meal-learn-pgdata`，再执行迁移。
4. 通过现有 API 新建一条临时食物记录，保存查询结果。
5. 停止并移除这个学习数据库容器，保留 volume，再用同一 volume 创建数据库容器。
6. 重新查询临时记录，验证数据仍然存在。
7. 用一个专门的临时目录练习 bind mount，观察电脑和容器对该目录内文件的访问。

命名 volume 由 Docker 管理；bind mount 把电脑上的指定路径挂载到容器。删除容器和删除 volume 的影响不同，练习时逐一检查操作对象。[官方 volume 说明](https://docs.docker.com/engine/storage/volumes/)

完成标准：

- [x] 能指出程序文件、容器临时文件和数据库文件分别保存在哪里（2026-10-15 按复述与补充讲解后的整体理解反馈记录）。
- [x] 能说明停止容器、移除容器和移除 volume 对数据的不同影响（2026-10-15 按概念讲解后的整体理解反馈记录，未逐项独立复述）。
- [x] 实际验证数据库容器重建后学习记录仍然存在（2026-10-15 学习者对三组实验整体确认，未提供原始输出或由助手独立复测）。

## 7. 手写 Docker Compose

要理解：Dockerfile 描述怎样制作镜像，Compose 描述怎样运行和组合服务。

当前状态（2026-10-16）：已完成，依据为学习者对整组实操与相关概念理解的确认。教学文件为 [compose.learn.yaml](../compose.learn.yaml)。以下分别记录助手在 2026-10-15 完成的准备检查，以及学习者在 2026-10-16 确认的实操结果。

| 已完成准备 | 实际结果与依据 |
| --- | --- |
| 助手只读检查容器、卷、网络与 `18084` 端口 | 手动练习 API 正在使用 `18080`、`18082`、`18083`；未发现 `18084` 监听进程；`dailymeal-learn_default` 与 `dailymeal-learn_pgdata` 尚不存在 |
| 创建教学 Compose 配置 | `db` 使用 `postgres:17-alpine`；`api` 与 `migrate` 使用 `backend/Dockerfile.learn` 构建 `daily-meal:learn`；迁移覆盖默认命令为 `/app/bin/migrate up`；API 绑定 `127.0.0.1:18084:8080` |
| `docker compose -p dailymeal-learn -f compose.learn.yaml config` | 助手实测退出码 0；展开后的默认网络名为 `dailymeal-learn_default`，卷名为 `dailymeal-learn_pgdata`；数据库地址为 `db:5432/dailymeal_docker_learn` |

已讲解内容（2026-10-16 学习者整体确认理解）：

- Compose 文件声明服务及其运行配置；服务是配置定义，容器是按定义创建的实际实例。项目名用于将这套容器、网络和卷分组，命令固定使用 `-p dailymeal-learn -f compose.learn.yaml`。
- `image` 对应镜像选择，`build.context` 对应构建上下文，`build.dockerfile` 相对上下文选择 Dockerfile；`environment`、`ports`、服务内 `volumes` 分别对应手动运行时的环境变量、端口映射与挂载。
- 本例使用 Compose 自动创建的默认网络，API 与迁移通过服务名 `db` 访问 PostgreSQL；数据库不需要向电脑发布端口。项目名不会自动解决电脑端口冲突，所以本例明确选用 `18084`。
- 数据库健康检查在数据库容器内部执行 `pg_isready`，通过 TCP 地址 `127.0.0.1` 检查数据库是否接受连接。`service_healthy` 让迁移等待健康检查通过，`service_completed_successfully` 让 API 等待迁移成功退出；数据库就绪与业务表已经迁移是不同条件。
- 顶层 `volumes.pgdata` 声明卷，服务内挂载把它接到 PostgreSQL 17 的数据目录 `/var/lib/postgresql/data`。首次启动使用新卷初始化学习数据库，之后保留卷重建容器时复用已有数据。
- 普通 `down` 移除该项目容器与默认网络，保留本例的命名卷；`down -v` 还会删除本例声明的卷及其数据。持久化练习只执行普通 `down`。
- 运行、接口访问与数据保留已按学习者整体反馈记录完成；未收到逐条原始输出或逐项独立复述，助手未独立复测。从头编写仍待第 8 阶段练习，不由整体理解反馈推定已经完成。

已完成记录（2026-10-16，学习者整体口头确认）：

学习者针对上一轮整组练习反馈“已经完成了，相关概念也理解了”。下表记录获确认的操作与结果；不补写实际容器 ID、食物 ID、响应原文或构建耗时。Compose 命令均使用 `docker compose -p dailymeal-learn -f compose.learn.yaml` 前缀。

| 实操范围 | 获确认的命令与操作 | 结果与证据边界 |
| --- | --- | --- |
| 配置检查、构建启动与状态检查 | `config`、`up --build -d`、`ps -a`、`logs --tail=50 db migrate api` | 按整组完成反馈记录：数据库健康、迁移成功退出、API 运行；未提供实际状态列表或日志 |
| 接口访问与学习数据写入 | `curl -fsS http://127.0.0.1:18084/healthz`；POST `/api/v1/foods` 写入 `compose-demo`（`basis_unit: g`，每 100g 的 `energy_kcal: 100`）；GET `/api/v1/foods?q=compose-demo` 查询并记下 ID | 按整组完成反馈记录：健康检查、创建与查询成功；未提供 HTTP 响应原文或实际食物 ID |
| 保留卷后删除重建 | `down`；`docker volume inspect dailymeal-learn_pgdata`；`up -d`；`ps -a`；再次访问健康检查与食物查询接口 | 按整组完成反馈记录：命名卷保留，重建后仍可查询同一条学习记录；未独立核对重建前后的 ID 或响应 |

概念掌握按学习者“相关概念也理解了”的整体确认记录，覆盖服务与项目、配置与手动命令的对应关系、默认网络、启动依赖、命名卷以及 `down` 与 `down -v` 的区别。本次持久化练习执行的是普通 `down`，没有将 `down -v` 的破坏性删除操作记为已执行；逐项独立复述与从头编写仍可在综合复习中验证。

练习：

1. 回看前面手动执行过的命令，把每个参数对应到 YAML 配置。
2. 新建根目录的 `compose.learn.yaml`，构建时明确选择 `backend/Dockerfile.learn`。
3. 定义数据库、一次性迁移和 API 三个服务，并配置环境变量、端口和命名 volume。
4. 定义数据库健康检查，让迁移等待数据库就绪，再让 API 等待迁移成功。
5. 使用独立的 Compose 项目名运行这套学习服务；端口先确认未被手动练习的 API 占用。

文件写好以后，学习下面的命令：

```sh
docker compose -p dailymeal-learn -f compose.learn.yaml config
docker compose -p dailymeal-learn -f compose.learn.yaml up --build -d
docker compose -p dailymeal-learn -f compose.learn.yaml ps -a
docker compose -p dailymeal-learn -f compose.learn.yaml logs
docker compose -p dailymeal-learn -f compose.learn.yaml down
```

先用 `config` 检查解析后的配置，再启动。单纯的启动顺序不代表数据库已经准备好；要结合健康检查和依赖条件。[官方启动顺序说明](https://docs.docker.com/compose/how-tos/startup-order/)

Compose 的默认网络支持按服务名通信。该阶段使用服务名作为数据库地址，例如数据库服务叫 `db`，后端连接 `db:5432`。本阶段新建学习 volume，不默认沿用手动练习的存储。[官方 Compose 网络说明](https://docs.docker.com/compose/how-tos/networking/)

启动后使用 `ps -a` 和 `logs --tail=50 db migrate api` 检查结果。预期 `db` 为运行且健康，`migrate` 为成功退出（退出码 0），`api` 为运行状态；迁移是一次性程序，成功退出是正常结果。随后请求 `http://127.0.0.1:18084/healthz`，预期返回 `{"status":"ok"}`。`up -d` 返回与 API 已可接受请求之间可能有短暂间隔，连接失败时先结合状态和日志判断。

接口及数据保留练习（2026-10-16 已按学习者整体反馈记录完成，保留命令供复习）：`curl -f` 将 HTTP 错误视为失败，`-sS` 关闭进度但保留错误；POST 的 `-H` 声明 JSON，`--data` 提供请求体。以下写入只针对 Compose 学习数据库。

```sh
curl -fsS http://127.0.0.1:18084/healthz
curl -fsS -X POST http://127.0.0.1:18084/api/v1/foods \
  -H 'Content-Type: application/json' \
  --data '{"name":"compose-demo","basis_unit":"g","nutrition_per_100":{"energy_kcal":100}}'
curl -fsS 'http://127.0.0.1:18084/api/v1/foods?q=compose-demo'
```

记下返回的食物 ID。查询成功后执行以下重建步骤：`down` 删除项目容器和默认网络，`docker volume inspect` 只读检查保留的卷；再次 `up -d` 创建新容器并挂载同一卷。源码没有修改，本次直接使用刚才构建的镜像。

```sh
docker compose -p dailymeal-learn -f compose.learn.yaml down
docker volume inspect dailymeal-learn_pgdata
docker compose -p dailymeal-learn -f compose.learn.yaml up -d
docker compose -p dailymeal-learn -f compose.learn.yaml ps -a
curl -fsS http://127.0.0.1:18084/healthz
curl -fsS 'http://127.0.0.1:18084/api/v1/foods?q=compose-demo'
```

预期重新查询仍能找到相同 ID 的记录，因为数据库数据保存在保留的命名卷中。若某一步失败，先查看该项目日志，不将失败步骤之后的预期结果记为已完成。

参考：[Compose 应用模型](https://docs.docker.com/compose/intro/compose-application-model/)、[服务配置](https://docs.docker.com/reference/compose-file/services/)、[构建配置与路径](https://docs.docker.com/reference/compose-file/build/)、[配置检查](https://docs.docker.com/reference/cli/docker/compose/config/)、[up](https://docs.docker.com/reference/cli/docker/compose/up/)、[down 与卷删除](https://docs.docker.com/reference/cli/docker/compose/down/)、[PostgreSQL 官方镜像初始化](https://hub.docker.com/_/postgres)。

完成标准：

- [x] 能逐项说明 Compose 配置与此前手动命令的对应关系（2026-10-16 按相关概念理解的整体反馈记录，未逐项独立复述）。
- [x] 能解释三个服务的启动条件，并正常访问 API（2026-10-16 按概念理解与整组实操完成的反馈记录，未独立复测）。
- [x] 能在停止、重新创建服务后验证学习数据保留（2026-10-16 学习者整体确认完成重建实验，未提供前后响应或由助手独立复测）。
- [x] 能解释 `down` 和 `down -v` 的不同影响，以及它们作用于哪个 Compose 项目（2026-10-16 按相关概念理解的整体反馈记录，未逐项独立复述）。

## 8. 综合排错与复习

当前状态（2026-10-16）：已完成，本轮基础学习结课。前两组排错流程及命名区别已获学习者理解确认，第二组修复后的运行与健康接口已实测；未提供的故障原始日志仍不补写。学习者独立提交综合练习配置，经审阅后自行修正，提供了完整镜像构建成功输出，并确认启动与接口验证成功。助手已实测当前数据库健康、迁移成功、API 运行以及两个 GET 接口返回 HTTP 200。结合此前概念理解反馈和本次完整任务，确认基础学习目标达到；本轮未新增逐项独立复述的事实如实保留，不再将其作为追加考试或补课的门槛。

| 项目 | 记录 |
| --- | --- |
| 助手只读检查 | `docker compose -p dailymeal-learn -f compose.learn.yaml ps -a` 显示 `db` 为运行且健康、`migrate` 为 `Exited (0)`、`api` 为运行；`docker network inspect dailymeal-learn_default` 确认网络存在且包含数据库与 API |
| 镜像与端口检查 | `docker image inspect daily-meal:learn` 确认默认命令为 `/app/bin/api`，镜像环境变量中没有 `DATABASE_URL`；`lsof -nP -iTCP:18085 -sTCP:LISTEN` 未发现监听；新练习容器名 `meal-learn-debug-env` 当前未占用 |
| 原因与预期 | `config.Load()` 读取不到 `DATABASE_URL` 时返回错误；API 记录错误并执行 `os.Exit(1)`，发生在数据库连接之前。预期容器 `Exited (1)`，日志含 `DATABASE_URL is required`，容器配置中没有该变量；退出码本身不能说明具体原因 |
| 练习设计 | 用同一镜像、网络 `dailymeal-learn_default` 和端口 `127.0.0.1:18085:8080` 创建独立 API 容器；先省略变量收集证据，再次启动验证配置未变，最后删除已退出的练习容器并通过 `-e` 补齐学习数据库地址后重建 |
| 修复验证 | 按整组完成反馈记录：补齐变量后重建并检查状态、日志及 `http://127.0.0.1:18085/healthz`；未提供响应原文，不补写实际日志或输出 |
| 学习者反馈 | 2026-10-16 反馈“好了”，据上下文记录本组实操完成；随后确认已理解前述流程，按整体反馈补记概念理解。尚不勾选整个阶段的完成标准 |
| 反馈后只读检查 | 当前 context 为 `desktop-linux`，`docker ps -a` 返回空列表；未发现 `dailymeal-learn_default` 网络；教学镜像和 `dailymeal-learn_pgdata` 卷仍存在。无法复查已不存在的 `meal-learn-debug-env` 容器日志或健康接口，未据此否定学习者的完成反馈 |

本组使用的命令与参数：`docker run -d --name ... --network ... -p ... daily-meal:learn` 创建后台运行的练习容器；`docker ps -a --filter name=meal-learn-debug-env` 同时查看已退出的容器；`docker logs --tail=20` 读取最近日志；`docker inspect --format` 提取 `.State.Status`、`.State.ExitCode` 和 `.Config.Env`。`docker start` 再次启动原容器并沿用原配置，因此仍应失败。确认退出后用 `docker rm meal-learn-debug-env` 删除该练习容器，再使用相同参数并增加 `-e 'DATABASE_URL=postgres://dailymeal:dailymeal@db:5432/dailymeal_docker_learn?sslmode=disable'` 创建新容器。修复改变的是容器运行时配置，不需要重新构建镜像。

参考：[运行容器与环境变量](https://docs.docker.com/engine/containers/run/#environment-variables)、[run 参数](https://docs.docker.com/reference/cli/docker/container/run/)、[日志](https://docs.docker.com/reference/cli/docker/container/logs/)、[inspect 格式化输出](https://docs.docker.com/reference/cli/docker/inspect/)、[启动已有容器](https://docs.docker.com/reference/cli/docker/container/start/)。

第二组步骤与进展（2026-10-16）：数据库主机名写错。

- 先运行 `docker compose -p dailymeal-learn -f compose.learn.yaml up -d` 恢复学习容器与网络，并检查 `ps -a`、日志和 `http://127.0.0.1:18084/healthz`。现有镜像与命名卷仍在，使用同一项目名和配置复用该卷；未验证当前卷中的具体业务记录。
- 将连接过程按名字解析、TCP 连接、PostgreSQL 认证区分。当前服务名为 `db`；把学习 API 的连接串主机名写为 `db-typo`，预期发生名字解析失败，通常可见 `hostname resolving error`、`lookup`、`no such host` 等关键词，具体文字以实际日志为准。
- 新练习容器命名为 `meal-learn-debug-db`，接入 `dailymeal-learn_default`，映射 `127.0.0.1:18085:8080`。首次前台运行，保留退出后的容器；通过 `ps -a`、`logs`、`inspect` 核对退出原因和实际变量，并用 `docker network inspect dailymeal-learn_default` 核对数据库成员。已退出的 API 可能不再显示为活动网络端点，不据此单独推断创建时未指定网络。
- 修复时删除已退出的练习容器，把连接串主机名改回 `db`，后台重建并检查日志和健康接口。上述步骤的完整执行结果待反馈，助手没有代为创建、删除容器或验证故障。

本组新增证据：学习者已提供 `docker network inspect dailymeal-learn_default` 原始输出，网络中有 `dailymeal-learn-api-1`（`172.19.0.3`）和 `dailymeal-learn-db-1`（`172.19.0.2`）。助手随后只读执行 `docker inspect dailymeal-learn-db-1 --format '{{json .NetworkSettings.Networks}}'`，确认该网络下的 `Aliases` 为 `["dailymeal-learn-db-1","db"]`，`IPAddress` 为 `172.19.0.2`。这证明当前网络及数据库网络别名已存在，不等于本组故障定位和接口修复已经完成。

针对学习者关于连接串中 `db` 与输出中 `dailymeal-learn-db-1` 不同的疑问，补充区分：`db` 是 Compose 文件 `services.db` 定义的服务名，Compose 将其注册为网络内别名；`dailymeal-learn-db-1` 是按项目、服务及实例序号生成的容器名；`dailymeal_docker_learn` 是 PostgreSQL 内部的数据库名。当前两个网络名称都指向 `172.19.0.2`。`docker network inspect` 的 `Containers.*.Name` 列出容器名，不会在该字段中列出所有 DNS 别名；需要检查容器的网络详情查看 `Aliases`。名字解析发生在对应 Docker 网络中，不应推定电脑终端也能直接解析 `db`。学习者随后确认已理解命名区别及前述流程，故障原始日志仍未提供。

修复后复核（2026-10-16，助手只读实测）：`docker ps -a --filter name=meal-learn-debug` 显示 `meal-learn-debug-db` 为运行状态，映射 `127.0.0.1:18085:8080`；`docker logs --tail=5 meal-learn-debug-db` 包含 `INFO listening address=:8080 business_timezone=Asia/Shanghai` 及此前 `/healthz` 的 200 访问记录；`curl --max-time 5 -fsS http://127.0.0.1:18085/healthz` 返回 `{"status":"ok"}`。该检查确认当前修复后的服务可用，不替代未收到的故障日志，也不将所有阶段完成标准勾选为通过。

参考：[Compose 服务名解析与网络](https://docs.docker.com/compose/how-tos/networking/)、[Compose up](https://docs.docker.com/reference/cli/docker/compose/up/)、[PostgreSQL 认证错误](https://www.postgresql.org/docs/17/client-authentication-problems.html)。

可选参考：数据库密码认证失败（2026-10-16 曾准备，未收到执行结果；已取消必做安排，无需补做）。

- 当前 Compose 数据库健康、迁移成功退出、API 运行；上一组练习 API 使用 `18085`。助手用 `lsof -nP -iTCP:18086 -sTCP:LISTEN` 检查时未发现 `18086` 监听，当前未发现名为 `meal-learn-debug-auth` 的容器。
- 使用新容器 `meal-learn-debug-auth`，网络仍为 `dailymeal-learn_default`，端口映射为 `127.0.0.1:18086:8080`。保持数据库主机 `db`、用户 `dailymeal` 和数据库名不变，仅将 API 连接串中的密码改为 `wrong-password`，以前台方式启动并保留退出后的容器。
- 预期能解析主机名并连接 PostgreSQL，但认证失败，日志包含 `password authentication failed for user "dailymeal"`，API 退出。通过练习容器的 `ps -a`、`logs --tail=20`、`inspect` 环境变量，以及 `docker compose -p dailymeal-learn -f compose.learn.yaml logs --tail=20 db` 对照客户端与数据库端证据。具体错误文字以实际结果为准。
- 修改 API 的 `DATABASE_URL` 只改变本次连接使用的凭据，不会修改 PostgreSQL 中用户的密码。修复是在确认故障容器退出后删除它，以相同参数把连接串密码改回 `dailymeal` 并后台重建，再检查日志及 `http://127.0.0.1:18086/healthz`。
- 以上保留此前准备的讲解与练习，尚无本组操作结果。按学习者反馈，此案例及端口占用不再作为进入综合练习的前置条件。

下表作为遇到问题时查阅的排错参考，不是要求逐项完成的清单。确有必要复现某个问题时，仅在学习环境里一次改变一个条件，先收集证据，再修改配置：

| 练习问题 | 要观察什么 |
| --- | --- |
| 缺少 `DATABASE_URL` | API 的启动日志和退出信息 |
| 数据库主机名写错，或容器不在同一个网络 | 连接错误、网络成员和实际连接地址 |
| 数据库密码写错 | 区分认证失败与网络不可达 |
| 电脑上的端口被占用 | 区分 Docker 创建/启动失败与 API 自身报错 |
| 跳过迁移后访问需要业务表的接口 | 区分数据库连接成功与表结构准备完成 |
| 重建数据库时误用了新的学习 volume | 检查挂载配置，解释为何查不到原来的学习记录 |

实际排错时记录操作、预期、实际输出、原因和修复方法。优先使用 `ps -a`、`logs`、`inspect`、`network inspect` 和 `volume inspect`，根据当前证据选择需要的检查，不要求每次执行全部命令。

接下来的综合练习：由学习者独立编写一份包含 API 和迁移程序的基础 Dockerfile，再编写管理数据库、迁移和 API 的 Compose 配置，完成构建、启动及接口访问。允许查阅官方语法，助手负责审阅配置、解释设计原因，并围绕实际出现的问题补充讲解；不预先代写整份答案，不额外制造故障凑数量。

综合练习初稿审阅（2026-10-16）：

- 学习者已提交 `backend/Dockerfile.practice` 和 `compose.practice.yaml`。构建上下文 `./backend`、两个编译入口、API 与迁移共用镜像、迁移启动命令、数据库凭据与连接地址、命名卷及启动条件的设计对应练习要求。配置文本的审阅不等于运行成功。
- Dockerfile 待修正：第 2 行 `From golang 1.25.0` 应使用镜像名与标签之间的冒号；第 14 行 `go bluid` 应为 `go build`；第 16 行应为 `CMD ["/app/bin/api"]`。Dockerfile 指令大小写不是本次阻塞原因；exec 形式使用 JSON 双引号，程序绝对路径需要开头的 `/`。
- Compose 待修正：第 23 行 `depend_on` 应为 `depends_on`，第 37 行 `port` 应为 `ports`。`command: ['/app/bin/migrate', 'up']` 是合法的 YAML 列表，无须因 Dockerfile 的 JSON 引号要求而改写。
- 助手执行 `docker compose -p dailymeal-practice -f compose.practice.yaml config --quiet`，返回 `services.migrate.depend_on false schema`。仅在标准输入中替换该字段后继续校验，返回 `services.api.port false schema`；在标准输入中同时修正两个字段后，配置校验退出码为 0。原始文件未修改，该结果仅验证修改建议的 Compose 结构，不代表镜像已构建或服务已启动。
- 助手执行 `docker build --check --progress=plain -f backend/Dockerfile.practice backend`，首次受沙箱限制无法连接 Docker socket；获准访问 Engine 后，实际返回 `dockerfile parse error on line 2: FROM requires either one or three arguments`。检查停在该行，未编译程序，其他 Dockerfile 问题来自源码审阅。
- 助手执行 `lsof -nP -iTCP:18084 -sTCP:LISTEN`，发现 Docker Desktop 进程正在监听 `127.0.0.1:18084`；本次检查 `18086` 未发现监听，可作为建议的主机端口，容器端口仍为 `8080`。未停止或删除现有容器。
- 下一步：学习者修正初稿并选择可用主机端口，以独立项目名 `dailymeal-practice` 完成配置校验、实际构建、服务启动和接口访问；本次未验收上述运行结果，未确认整个阶段完成。

参考：[Dockerfile 的 FROM、exec 形式与 CMD](https://docs.docker.com/reference/dockerfile/)、[Compose 服务字段](https://docs.docker.com/reference/compose-file/services/)。

综合练习构建网络排查（2026-10-16）：

- 学习者已将基础镜像改为 `golang:1.25.0`，修正 `go build`、`CMD`、`depends_on` 和 `ports`，API 映射改为 `127.0.0.1:18086:8080`。助手重新执行 `docker compose -p dailymeal-practice -f compose.practice.yaml config --quiet`，退出码为 0。
- 学习者提供的实际构建日志显示：`load metadata for docker.io/library/golang:1.25.0` 阶段失败，具体为 `failed to fetch oauth token` 和连接 `auth.docker.io` 的 `i/o timeout`。这次未执行到依赖下载或 Go 编译；`ConsistentInstructionCasing` 是另一类大小写警告。随消息提供的 `/healthz` 200 日志没有本次练习容器来源，不能用于验收本次练习服务。
- 助手执行 `docker build --check --progress=plain -f backend/Dockerfile.practice backend`，复现连接 `199.16.156.40:443` 超时。当前使用的 `desktop-linux` 构建器正在运行；本机相关镜像列表仅显示 `daily-meal:learn`，未显示独立的 `golang` 标签，不能据此断言所有构建缓存均不存在。
- 只读检查显示：Docker Engine 的 HTTP/HTTPS 代理地址为 `http.docker.internal:3128`；macOS 系统 HTTP/HTTPS/SOCKS 代理已启用，地址为 `127.0.0.1:7890`。系统代理配置存在不代表当前命令的所有请求都会使用该代理。
- 主机普通 `curl` 请求匿名令牌接口在约 5 秒后连接超时；显式加上 `--proxy http://127.0.0.1:7890 --noproxy ''` 后，同一接口返回 HTTP 200，耗时约 3.94 秒。返回正文丢弃，未记录或保存令牌。
- 随后执行 `HTTP_PROXY=http://127.0.0.1:7890 HTTPS_PROXY=http://127.0.0.1:7890 docker build --check --progress=plain -f backend/Dockerfile.practice backend`，基础镜像元数据获取在约 8.3 秒内成功，检查只剩 `From` 与 `run` 两处大小写警告并返回退出码 1。该结果证明这次指定代理后通过了此前失败的网络步骤，不代表整个检查无警告或完整镜像已构建成功。
- 上述环境变量仅为单次命令设置；没有修改 macOS 或 Docker Desktop 代理、DNS、VPN、Dockerfile 或 Compose 文件。尚未进一步确认底层 DNS、路由或代理规则中的具体原因，也未执行完整构建、启动本次练习容器或进行接口验收。
- 下一步：学习者将 `From`、`run` 统一为 `FROM`、`RUN`，在现有代理运行时，为 `docker compose -p dailymeal-practice -f compose.practice.yaml build` 临时指定相同的 `HTTP_PROXY` 和 `HTTPS_PROXY`，完成实际构建后再启动并验证服务。

参考：[Docker Registry 令牌授权流程](https://docs.docker.com/reference/api/registry/auth/)、[Dockerfile 指令大小写](https://docs.docker.com/reference/dockerfile/#format)。

综合练习构建成功（2026-10-16）：

- 学习者执行 `HTTP_PROXY=http://127.0.0.1:7890 HTTPS_PROXY=http://127.0.0.1:7890 docker compose -p dailymeal-practice -f compose.practice.yaml build`，提供完整输出并反馈“现在好了”。日志显示 `Building 250.2s (17/17) FINISHED`，最终为 `Image daily_meal_practice Built`，导出标签为 `daily_meal_practice:latest`。
- 日志确认基础镜像下载及解包完成，`go mod download`、`go build -o bin/api ./cmd/api` 和 `go build -o bin/migrate ./cmd/migrate` 均已执行成功。此处记录的是学习者实际运行的原始输出，助手未重复执行完整构建。
- `From` 与 `run` 仍产生大小写一致性警告；本次实际构建成功，可以继续运行验证，统一指令大写保留为格式整理项。
- 助手只读执行 `docker image inspect daily_meal_practice:latest postgres:17-alpine --format '{{json .RepoTags}} {{.Os}}/{{.Architecture}}'`，确认两个标签均存在，平台均为 `linux/arm64`。`docker compose -p dailymeal-practice -f compose.practice.yaml ps -a` 当前仅有表头，尚无该项目容器。
- 下一组操作：`docker compose -p dailymeal-practice -f compose.practice.yaml up -d --no-build --pull never`，随后查看 `ps -a`、最近日志，并访问 `http://127.0.0.1:18086/healthz` 与 `/api/v1/foods`。`-d` 表示后台运行；`--no-build` 跳过构建；`--pull never` 使用已经确认存在的本地镜像，不拉取镜像。
- 预期数据库运行且健康、迁移成功退出、API 保持运行，健康接口和食物列表接口成功返回。上述启动和接口操作尚未执行或反馈，不将预期状态记为已完成，也未重新验证本次练习的数据持久化。

参考：[Compose up](https://docs.docker.com/reference/cli/docker/compose/up/)、[包含已退出容器的状态列表](https://docs.docker.com/reference/cli/docker/compose/ps/)。

综合练习启动与接口验证完成（2026-10-16）：

- 学习者对 `up -d --no-build --pull never`、`ps -a`、`logs --tail=30`、`/healthz` 与 `/api/v1/foods` 整组操作反馈“成功了”。本轮没有粘贴原始输出，先按整体反馈记录完成，再以下列助手只读检查补充当前运行证据。
- `docker compose -p dailymeal-practice -f compose.practice.yaml ps -a`：`dailymeal-practice-db-1` 为 `Up (healthy)`；`dailymeal-practice-migrate-1` 为 `Exited (0)`；`dailymeal-practice-api-1` 为 `Up`，映射 `127.0.0.1:18086->8080/tcp`。
- `docker compose -p dailymeal-practice -f compose.practice.yaml logs --tail=15 migrate`：`00001_initial.sql`、`00002_add_favorite_to_foods.sql`、`00003_add_meal_entry_requests.sql` 均为 `OK`，最终日志为 `goose: successfully migrated database to version: 3`。
- 助手使用 `curl --noproxy '*' -fsS --max-time 5` 请求本机 `18086` 端口：`/healthz` 返回 HTTP 200，正文为 `{"status":"ok"}`；`/api/v1/foods` 返回 HTTP 200，正文为 `{"items":[],"page":1,"page_size":20,"total":0}`。
- 对照结果：本次镜像构建有学习者原始成功输出，容器运行、迁移完成和接口可用有助手当前实测。构建、运行、接口访问和持久化是不同的验证范围；本次未写入业务数据，也没有删除重建容器验证该练习卷的数据保留。第 6、7 阶段的持久化练习继续按此前学习者反馈记录，不伪装为本次实测。
- 概念复述情况：本轮反馈为操作成功，未新增独立复述；不据此自动勾选剩余概念标准，也不要求重做已完成的操作。相关记录保留证据边界，不追加独立复述考核。
- 可选进阶：以后有镜像瘦身需求时，再学习多阶段构建、`COPY --from` 和运行镜像的文件取舍。此前将其列为必须继续的“下一步”不符合本轮基础课程的范围，现已取消该安排；尚未修改实践 Dockerfile 或构建新的多阶段镜像。

参考：[多阶段构建](https://docs.docker.com/build/building/multi-stage/)。

完成标准：

- [x] 理解根据状态、日志和实际配置定位问题，再修复并验证的排查流程（2026-10-16 按学习者整体理解反馈记录；独立运用情况在综合练习中观察，不按故障数量验收）。
- [x] 能从零写出基础 Dockerfile 和 Compose，而不是只会复制文件（2026-10-16 学习者自行提交初稿，经助手审阅后自行修正；配置校验通过，已提供实际构建成功输出）。
- [ ] 能说明镜像、容器、网络和存储各自的作用（相关概念在此前阶段已讲解并有整体理解反馈；本阶段未新增完整独立复述，此处保留单项证据状态，不要求为结课另行复述）。
- [ ] 能区分镜像构建成功、程序运行成功、接口可用和数据持久化已经验证（2026-10-16 已对照实际结果说明范围；本轮没有独立复述记录，此处保留单项证据状态，不作为继续补课的条件）。

## 每次学习的记录

第 0～8 阶段的已知结果已写在各阶段记录中。本轮基础课程于 2026-10-16 结课，第 8 阶段的构建有原始成功输出，运行及接口有助手实测；此前理解反馈和未单项验证的证据边界继续保留，不新增补课要求。以后确有学习需求时，可以复制下面的表格记录新练习，实际完成后及时回写。预期现象、学习者口头反馈、原始输出与助手实测分别注明。

| 项目 | 记录 |
| --- | --- |
| 日期与阶段 | |
| 这次要理解的问题 | |
| 参考的官方文档与条目 | |
| 核心概念及容易混淆的区别 | |
| 修改的文件或执行的命令 | |
| 命令与主要参数的含义 | |
| 执行前已讲清的预期结果及原因 | |
| 实际输出与观察 | |
| 用自己的话解释原因 | |
| 尚未理解的地方 | |

多阶段构建、缓存优化、镜像发布和底层原理属于可选进阶，不是本轮基础课程的结课要求，不自动安排新的必修阶段。

当前下一步：本轮 Docker 基础课程结束，将已学内容用于日常项目开发即可。以后遇到实际需求时再补充对应进阶知识，当前不安排新的必修练习、重复实验或独立复述考试。本次练习未验证的数据写入及删除重建后的持久化仍按实际证据记录，不影响本轮基础学习结课。
