# SkillsHub 私有仓库平台
## 产品需求文档（PRD）

**文档版本：** v1.0.0  
**创建日期：** 2026-03-20  
**文档状态：** 草稿  
**产品负责人：** 待定  
**技术栈：** Node.js / TypeScript  

---

## 目录

1. [产品概述](#1-产品概述)
2. [背景与问题定义](#2-背景与问题定义)
3. [目标用户与角色](#3-目标用户与角色)
4. [产品架构总览](#4-产品架构总览)
5. [核心功能模块](#5-核心功能模块)
6. [安全机制设计](#6-安全机制设计)
7. [API 接口规范](#7-api-接口规范)
8. [数据模型设计](#8-数据模型设计)
9. [部署与运维](#9-部署与运维)
10. [非功能性需求](#10-非功能性需求)
11. [里程碑与迭代计划](#11-里程碑与迭代计划)
12. [风险与约束](#12-风险与约束)
13. [附录](#13-附录)

---

## 1. 产品概述

### 1.1 产品名称

**SkillsHub** —— 企业级 AI Agent Skill 私有仓库平台

### 1.2 一句话定义

SkillsHub 是一个面向企业团队的 AI Agent Skill 全生命周期管理平台，提供私有仓库托管、开源 Skill 自动引入与安全扫描、团队权限管控、以及与 Claude Code 深度集成的能力，确保团队成员只能使用经过安全审核的 Skill。

### 1.3 核心价值主张

| 价值维度 | 描述 |
|----------|------|
| **安全可控** | 杜绝未经审查的 Skill 进入开发环境，防止供应链攻击 |
| **效率提升** | 自动从开源仓库引入并扫描，减少人工干预 |
| **合规审计** | 完整记录 Skill 使用与安装行为，满足企业合规需求 |
| **团队协作** | 统一 Skill 版本，消除成员间环境不一致问题 |

---

## 2. 背景与问题定义

### 2.1 背景

随着 AI Agent（以 Claude Code 为代表）在企业开发团队中的大规模普及，Skill 生态系统快速扩张。Skill 本质上是**注入到 AI 会话上下文中的可执行指令集**，通过 `SKILL.md` 文件及配套脚本驱动 Claude 完成特定任务。

Claude Code 的 Skill 加载机制如下：

```
启动时扫描 ~/.claude/skills/ 目录
    ↓
读取各 SKILL.md 的 YAML frontmatter（name + description）
    ↓
注入 System Prompt 的 <available_skills> 块
    ↓
用户发言时 LLM 推断匹配 Skill
    ↓
加载 SKILL.md 全文 + 执行配套脚本
```

### 2.2 核心问题

**问题一：供应链安全风险**  
开发者从 GitHub、公共 Skill 市场等渠道随意安装 Skill，无任何审查机制。恶意 Skill 可通过以下方式危害企业：

- **提示词注入**：在 SKILL.md 中嵌入越权指令，劫持 Claude 行为
- **数据外泄**：通过工具调用将上下文中的敏感代码/配置发送到外部服务器
- **权限提升**：诱导 Claude 执行超出业务范围的系统操作

**问题二：环境不一致**  
不同成员使用不同版本的同一 Skill，导致相同提示词产生不同结果，难以协作和排查问题。

**问题三：缺乏审计追踪**  
无法知道团队成员安装了哪些 Skill、谁在使用、使用频率如何，出现问题后无法溯源。

**问题四：开源 Skill 引入效率低**  
手动评估开源 Skill 的安全性耗时耗力，导致团队要么放弃使用有价值的 Skill，要么绕过安全程序直接使用。

### 2.3 解决思路

建立**私有仓库 + 自动扫描入库**的双层机制：

1. **强制私有仓库**：团队成员的 Claude Code 只允许从企业私有仓库安装 Skill
2. **开源自动引入**：当私有仓库没有所需 Skill 时，自动从开源仓库拉取 → 安全扫描 → 审核通过后入库

---

## 3. 目标用户与角色

### 3.1 用户角色定义

#### 角色一：普通开发者（Developer）
- **典型场景**：日常使用 Claude Code 进行代码开发，需要安装各种 Skill 辅助工作
- **核心诉求**：
  - 能快速找到所需 Skill
  - 安装流程简单，不影响开发效率
  - 不用关心安全审查细节
- **权限范围**：搜索、安装已审核 Skill；申请引入新 Skill

#### 角色二：安全审核员（Security Reviewer）
- **典型场景**：对从开源仓库引入的 Skill 进行人工审核
- **核心诉求**：
  - 清晰的扫描报告，快速定位风险点
  - 高效的审批工作流
  - 可追溯的审核记录
- **权限范围**：查看扫描报告；审批/拒绝 Skill 入库；添加审核备注

#### 角色三：平台管理员（Admin）
- **典型场景**：管理整个 SkillsHub 平台，制定安全策略
- **核心诉求**：
  - 全局可见的使用统计和安全态势
  - 灵活的策略配置
  - 成员和权限管理
- **权限范围**：所有功能 + 系统配置 + 用户管理 + 策略制定

#### 角色四：CI/CD 系统（System）
- **典型场景**：通过 API 自动触发 Skill 安装、同步操作
- **权限范围**：API Token 授权的特定接口

### 3.2 用户旅程

**场景：开发者发现私有仓库没有所需 Skill**

```
开发者在 Claude Code 中使用 /install-skill <name>
    ↓
Claude Code 请求 SkillsHub API
    ↓
私有仓库未找到 → 自动触发开源仓库搜索
    ↓
找到候选 Skill → 自动拉取并进入安全扫描队列
    ↓
扫描完成：
  ├── 低风险 → 自动通知审核员 → 审核通过后通知开发者
  └── 高风险 → 阻断 + 通知安全团队
    ↓
审核通过 → Skill 入库 → 开发者安装成功
```

---

## 4. 产品架构总览

### 4.1 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    开发者工作环境                             │
│  ┌──────────────┐    ┌─────────────────────────────────┐   │
│  │  Claude Code  │───▶│  SkillsHub CLI / Agent Plugin   │   │
│  └──────────────┘    └───────────────┬─────────────────┘   │
└──────────────────────────────────────│─────────────────────┘
                                       │ HTTPS
┌──────────────────────────────────────▼─────────────────────┐
│                    SkillsHub 服务端                          │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐   │
│  │  API Gateway │  │  Web 管理台  │  │  后台任务队列    │   │
│  │  (Express)   │  │  (React)    │  │  (Bull/Redis)    │   │
│  └──────┬──────┘  └──────┬──────┘  └────────┬─────────┘   │
│         │                │                   │              │
│  ┌──────▼──────────────────────────────────▼─┐             │
│  │              核心业务服务层                  │             │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐ │             │
│  │  │  仓库服务 │  │  扫描服务 │  │  审核服务 │ │             │
│  │  └──────────┘  └──────────┘  └──────────┘ │             │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐ │             │
│  │  │  用户服务 │  │  通知服务 │  │  审计服务 │ │             │
│  │  └──────────┘  └──────────┘  └──────────┘ │             │
│  └────────────────────────┬────────────────────┘             │
│                           │                                  │
│  ┌────────────────────────▼────────────────────┐            │
│  │                  数据层                       │            │
│  │  ┌──────────┐  ┌──────────┐  ┌────────────┐ │            │
│  │  │PostgreSQL │  │  Redis   │  │  MinIO/S3  │ │            │
│  │  │(元数据)   │  │(缓存/队列)│  │(Skill文件) │ │            │
│  │  └──────────┘  └──────────┘  └────────────┘ │            │
│  └─────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────┘
          │
          │ 外部集成
          ▼
┌─────────────────────┐    ┌─────────────────────┐
│  开源 Skill 仓库      │    │  通知服务             │
│  (GitHub / npm      │    │  (Slack / Email /    │
│   Registry 等)       │    │   企业微信)           │
└─────────────────────┘    └─────────────────────┘
```

### 4.2 技术栈选型

| 层次 | 技术 | 理由 |
|------|------|------|
| 运行时 | Node.js 20 LTS + TypeScript 5 | 团队熟悉，生态完善，类型安全 |
| Web 框架 | Express.js + Zod 参数校验 | 轻量灵活，中间件生态丰富 |
| 数据库 | PostgreSQL 15 | 关系型数据，事务支持，JSON 字段灵活 |
| 缓存/队列 | Redis 7 + BullMQ | 高性能缓存，任务队列首选 |
| 文件存储 | MinIO（私有化）/ AWS S3 | Skill 文件存储，支持版本管理 |
| ORM | Prisma | 类型安全，迁移管理 |
| 认证 | JWT + OIDC（企业 SSO 集成） | 支持企业统一登录 |
| 管理台前端 | React 18 + Ant Design | 开发效率高，组件完善 |
| 容器化 | Docker + Docker Compose | 简化部署，环境一致性 |
| 日志 | Winston + ELK Stack | 结构化日志，便于审计查询 |

---

## 5. 核心功能模块

### 5.1 模块一：私有仓库管理

#### 5.1.1 Skill 上传与发布

**功能描述**  
管理员或授权用户可将内部开发的 Skill 上传到私有仓库，支持版本管理。

**上传流程**

```
用户提交 Skill 文件包（ZIP）
    ↓
系统校验包结构：
  必须包含：SKILL.md（含合法 frontmatter）
  可选包含：脚本文件、配置文件、README.md
    ↓
自动扫描（同开源引入流程）
    ↓
扫描通过 → 发布到私有仓库
扫描失败 → 返回详细报告，等待修复
```

**Skill 包结构规范**

```
skill-name/
├── SKILL.md              # 必须，包含 YAML frontmatter
├── README.md             # 推荐，用户可见说明
├── scripts/              # 可选，配套执行脚本
│   ├── main.py
│   └── utils.js
├── examples/             # 可选，使用示例
└── .skillhub.yaml        # 必须，SkillsHub 元数据配置文件
```

**.skillhub.yaml 规范**

```yaml
# .skillhub.yaml
name: "pdf-processor"           # 与 SKILL.md frontmatter 一致
version: "1.2.0"                # 语义化版本
category: "document"            # 分类标签
author: "internal-team"         # 作者
permissions:                    # 申请的权限声明（用于扫描对比）
  - file_read
  - file_write
  - network: none               # none | internal | external
external_calls: []              # 声明的外部网络调用列表
```

#### 5.1.2 版本管理

- 遵循语义化版本（SemVer）：`major.minor.patch`
- 支持同时维护多个大版本
- 提供版本对比功能（diff SKILL.md 内容）
- 支持将特定版本设为 `latest` / `stable` / `deprecated`
- 版本锁定：团队可锁定到特定版本，禁止自动升级

#### 5.1.3 Skill 目录与搜索

**搜索能力**

- 关键词全文搜索（搜索 name、description、README）
- 按分类、标签、作者、安全评分过滤
- 按下载量、评分、更新时间排序
- 支持模糊匹配（"pdf" 可找到 "pdf-extractor"、"pdf-converter" 等）

**Skill 详情页展示内容**

| 字段 | 说明 |
|------|------|
| 基本信息 | 名称、版本、描述、作者、更新时间 |
| 安全评分 | A/B/C/D/F 五级评分 + 扫描报告摘要 |
| 权限声明 | 该 Skill 申请的工具调用权限列表 |
| 使用统计 | 安装次数、活跃用户数、周使用量趋势 |
| 版本历史 | 所有版本列表 + 变更日志 |
| 依赖关系 | 依赖的其他 Skill 或外部服务 |
| 安装命令 | 一键复制安装命令 |

---

### 5.2 模块二：开源 Skill 自动引入

#### 5.2.1 触发机制

**触发方式一：CLI 主动搜索**

```bash
# 开发者在 Claude Code 中
/install-skill excel-processor

# SkillsHub CLI 执行：
# 1. 查询私有仓库 → 未找到
# 2. 自动触发开源搜索流程
# 3. 返回："正在从开源仓库搜索 excel-processor，预计 5 分钟完成安全扫描，完成后通知您"
```

**触发方式二：管理台手动申请**

管理员在 Web 管理台输入 Skill 名称或 GitHub URL，手动触发引入流程。

**触发方式三：API 触发**

```http
POST /api/v1/skills/import
Authorization: Bearer <token>
Content-Type: application/json

{
  "source": "github",
  "url": "https://github.com/org/skill-name",
  "requestedBy": "user-id",
  "priority": "normal"
}
```

#### 5.2.2 开源仓库搜索策略

**搜索源优先级**（按可信度排序）

1. **官方 Anthropic Skills Marketplace**（如存在）
2. **GitHub** —— 按 stars、更新时间、关键词匹配度综合排序
3. **npm Registry** —— 搜索 `@claude-skill/` 命名空间下的包
4. **企业内部配置的白名单镜像源**

**候选 Skill 评估指标**（自动打分）

```
综合可信度评分 = 
  仓库 Stars 权重(20%) +
  最近更新时间权重(15%) +
  维护者账号可信度(15%) +
  社区使用量(20%) +
  代码提交历史规律性(10%) +
  文档完整性(10%) +
  License 合规性(10%)
```

#### 5.2.3 文件拉取与隔离

```
从源下载 Skill 压缩包
    ↓
解压到隔离沙箱目录（/sandbox/<job-id>/）
    ↓
校验文件结构完整性
    ↓
进入安全扫描流水线
```

**沙箱隔离要求**：拉取过程在独立 Docker 容器中进行，无网络出口权限，防止恶意内容在拉取阶段就发起外连。

---

### 5.3 模块三：安全扫描引擎

这是整个系统的核心，扫描分为四个层次，逐层递进。

#### 5.3.1 Layer 1：结构与格式扫描（自动，~5秒）

**检查项**

| 检查项 | 通过标准 | 失败处理 |
|--------|----------|----------|
| SKILL.md 存在性 | 文件存在于根目录 | 阻断，标记 INVALID |
| YAML frontmatter 合法性 | name 和 description 字段存在且非空 | 阻断，标记 INVALID |
| 文件编码 | 所有文本文件为 UTF-8 | 警告 |
| 文件大小 | 单文件 < 10MB，总包 < 50MB | 阻断 |
| 文件类型白名单 | 仅允许：.md .py .js .ts .sh .json .yaml .yml .txt | 阻断含非白名单文件 |
| 隐藏文件/目录 | 不得包含 `.git`、`.env` 等隐藏配置 | 警告，移除后继续 |

#### 5.3.2 Layer 2：静态内容扫描（自动，~30秒）

**2a. 提示词意图分析**

对 SKILL.md 全文进行 NLP 分析，检测以下危险模式：

```yaml
危险指令模式：
  - 角色覆盖类: ["ignore previous instructions", "disregard", "forget your rules", 
                 "忽略之前的指令", "你现在是"]
  - 越权操作类: ["delete all", "rm -rf", "DROP TABLE", "format disk"]
  - 外连数据类: ["curl", "wget", "fetch", "POST.*http", "send.*to.*http"]
  - 密钥提取类: ["API_KEY", "SECRET", "PASSWORD", "TOKEN", "private key"]
  - 提权类: ["sudo", "chmod 777", "chown root", "Administrator"]
  
可疑模式（警告级）：
  - 过于宽泛的文件操作: ["/**", "~/", "C:\\"]
  - 编码混淆: [base64 解码指令, unicode 转义序列密集区域]
  - 条件隐藏: ["if.*user.*then.*else", "when.*not observed"]
```

**2b. 脚本代码静态分析**

对 `.py`、`.js`、`.ts`、`.sh` 文件进行 AST 静态分析：

```
Python 文件：
  - 检测 os.system(), subprocess.*, eval(), exec()
  - 检测 socket, requests, urllib 等网络调用
  - 检测文件系统敏感路径访问

JavaScript/TypeScript 文件：
  - 检测 child_process.exec/spawn
  - 检测 fetch/axios/http 等网络调用
  - 检测 fs 模块对敏感路径的访问
  - 检测 eval() 和 Function() 构造器

Shell 脚本：
  - 检测危险命令：rm -rf, dd, mkfs, iptables
  - 检测外连命令：curl, wget, nc, ncat
  - 检测权限修改：sudo, chmod, chown
```

**2c. 密钥与敏感信息扫描**

使用正则表达式扫描是否硬编码了：

- API Keys（各主流服务的 key 格式）
- JWT tokens
- 私钥 PEM 块
- IP 地址（内网/公网）
- 邮箱地址
- 数据库连接字符串

#### 5.3.3 Layer 3：行为沙箱测试（自动，~2分钟）

在完全隔离的环境中对 Skill 进行动态行为测试：

**沙箱环境规格**

```yaml
沙箱配置:
  runtime: docker
  image: skillhub-sandbox:latest
  network: none          # 完全断网
  memory: 256m
  cpu: 0.5
  timeout: 120s
  readonly_rootfs: true
  allowed_paths:         # 仅允许读写以下路径
    - /sandbox/skill/    # Skill 文件本身
    - /sandbox/output/   # 输出目录
    - /tmp/              # 临时文件
```

**测试用例集**

```
Test 1: 基本加载测试
  - 模拟 Claude Code 加载 SKILL.md
  - 验证不产生意外的文件创建/修改

Test 2: 模拟任务执行
  - 提供标准测试输入
  - 监控系统调用（strace）
  - 检测是否有意外的网络尝试

Test 3: 边界输入测试
  - 空输入、超长输入、特殊字符输入
  - 验证不崩溃、不产生意外输出

Test 4: 权限探测测试
  - 尝试访问沙箱外路径
  - 验证是否有越权行为
```

**监控指标**

| 指标 | 阈值 | 触发动作 |
|------|------|----------|
| 网络连接尝试次数 | > 0 | 高风险标记 |
| 文件创建/修改（沙箱外） | > 0 | 阻断 |
| 子进程数量 | > 10 | 警告 |
| 内存使用 | > 200MB | 警告 |
| 执行超时 | > 60s | 警告 |
| 系统调用异常 | 出现禁止调用 | 高风险标记 |

#### 5.3.4 Layer 4：供应链溯源分析（自动，~1分钟）

```
来源仓库分析:
  - 仓库创建时间（< 30天 → 高风险加权）
  - 维护者账号注册时间和活跃度
  - commit 历史规律性（突然大量提交 → 可疑）
  - fork 关系（是否 fork 自可信仓库但内容被修改）
  - 已知恶意仓库黑名单比对

License 合规检查:
  - 识别 License 类型（MIT/Apache/GPL/无License）
  - 无 License → 警告（版权风险）
  - GPL → 警告（传染性）
  - 禁止商业使用条款 → 警告
```

#### 5.3.5 综合评分与风险等级

```
综合风险评分计算:

每个检查项有对应权重和分值，汇总得出 0-100 分
分值越低风险越高

风险等级划分:
  A级（85-100分）：低风险，可自动审核通过
  B级（70-84分） ：轻微风险，建议人工复核
  C级（50-69分） ：中等风险，必须人工审核
  D级（30-49分） ：高风险，需安全团队专项审查
  F级（0-29分）  ：严重风险，自动拒绝入库

自动处理规则（可由管理员配置）:
  A级 → 可配置为自动入库
  B级 → 自动通知审核员，24小时内处理
  C级 → 通知安全审核员，48小时内处理
  D级 → 通知安全团队，冻结等待专项审查
  F级 → 自动拒绝，通知申请者和安全团队
```

---

### 5.4 模块四：人工审核工作台

#### 5.4.1 审核队列

**界面功能**

- 待审核列表（按优先级、风险等级排序）
- 审核任务分配（支持认领和指派）
- SLA 倒计时（B级24h，C级48h）
- 批量操作（批量通过低风险项目）

**审核详情页展示**

```
审核详情页包含以下 Tab：

[概览] 基本信息、综合评分、风险摘要、申请者信息

[扫描报告] 
  - Layer 1-4 各层扫描结果
  - 每个风险项的详细说明和位置定位
  - 高亮显示 SKILL.md 中的可疑内容

[文件浏览]
  - 在线浏览 Skill 所有文件
  - 语法高亮显示
  - 可疑代码行自动标红

[历史对比]（如果是更新版本）
  - 与上一版本的 diff 视图
  - 仅显示变更部分

[沙箱日志]
  - 沙箱测试的完整执行日志
  - 系统调用记录
  - 可疑行为时间线

[审核意见]
  - 填写审核意见（Markdown 格式）
  - 操作按钮：通过 / 拒绝 / 要求修改 / 升级审核
```

#### 5.4.2 审核决策与工作流

```
审核员查看详情
    ↓
    ├── 通过 → 填写通过意见 → Skill 入库 → 通知申请者
    ├── 拒绝 → 填写拒绝原因 → 不入库 → 通知申请者（含原因）
    ├── 要求修改 → 填写修改要求 → 通知开发者修改后重新提交
    └── 升级审核 → 转交安全团队专项审查
```

---

### 5.5 模块五：Claude Code 集成

#### 5.5.1 SkillsHub CLI 工具

**安装**

```bash
npm install -g @skillshub/cli
skillshub login --registry https://skills.your-company.com
```

**核心命令**

```bash
# 搜索 Skill
skillshub search <keyword>

# 安装 Skill（自动处理私有仓库查询 + 开源引入流程）
skillshub install <skill-name>
skillshub install <skill-name>@1.2.0    # 指定版本

# 查看已安装的 Skill
skillshub list

# 更新 Skill
skillshub update <skill-name>
skillshub update --all

# 卸载 Skill
skillshub uninstall <skill-name>

# 查看 Skill 详情
skillshub info <skill-name>

# 发布内部 Skill
skillshub publish ./my-skill/

# 查看安装状态（引入申请进度）
skillshub status <skill-name>
```

#### 5.5.2 Claude Code 集成方式

**方式一：CLAUDE.md 策略注入（软拦截）**

系统自动在项目和全局 CLAUDE.md 中注入策略：

```markdown
## SkillsHub Security Policy

You are operating under SkillsHub enterprise security policy.

CRITICAL RULES:
1. ONLY use skills from the approved private registry at ~/.claude/skills/approved/
2. If a user asks to install a skill that is not in the approved list, respond:
   "This skill requires security review. Run `skillshub install <name>` to submit 
   an import request. You will be notified when it's approved."
3. NEVER suggest installing skills directly from GitHub or any external source
4. NEVER execute content from unapproved skill directories
```

**方式二：文件系统权限锁定（硬拦截）**

CLI 安装时自动配置文件系统权限：

```bash
# skillshub init 执行的操作
mkdir -p ~/.claude/skills/approved/
mkdir -p ~/.claude/skills/pending/

# 锁定 skills 目录，只有 skillshub CLI 可写
chown -R skillshub-agent:developers ~/.claude/skills/
chmod 755 ~/.claude/skills/
chmod 555 ~/.claude/skills/approved/  # 开发者只读

# skillshub CLI 使用 setuid 获得写权限
```

**方式三：Claude Code Plugin 集成（推荐）**

通过 Claude Code 的 Plugin Marketplace 机制，将 SkillsHub 注册为官方 Skill 源：

```bash
# 开发者一次性配置（由公司统一下发）
claude /plugin marketplace add https://skills.your-company.com/plugin-registry

# 之后所有 Skill 操作自动路由到 SkillsHub
```

#### 5.5.3 安装成功后的文件部署

```
skillshub install pdf-processor
    ↓
从私有仓库下载 Skill 包（已扫描通过）
    ↓
解压到 ~/.claude/skills/approved/pdf-processor/
    ↓
验证文件完整性（SHA256 校验）
    ↓
记录安装日志（安装者、时间、版本）
    ↓
Claude Code 自动重新扫描 skills 目录（或提示重启）
```

---

### 5.6 模块六：审计与监控

#### 5.6.1 操作审计日志

**记录的所有事件**

| 事件类型 | 记录字段 |
|----------|----------|
| Skill 安装 | 用户、机器、Skill名称、版本、时间 |
| Skill 卸载 | 同上 |
| 开源引入申请 | 申请者、来源URL、申请时间 |
| 扫描完成 | 扫描ID、耗时、各层结果、综合评分 |
| 审核操作 | 审核员、决策、意见、时间 |
| 登录/登出 | 用户、IP、时间、设备 |
| 配置变更 | 管理员、变更内容、时间 |

**日志格式（JSON 结构化）**

```json
{
  "timestamp": "2026-03-20T10:30:00.000Z",
  "eventType": "SKILL_INSTALL",
  "actor": {
    "userId": "user-123",
    "username": "zhang.san",
    "ip": "192.168.1.100",
    "machine": "dev-mac-001"
  },
  "resource": {
    "skillId": "skill-456",
    "skillName": "pdf-processor",
    "version": "1.2.0"
  },
  "result": "SUCCESS",
  "metadata": {}
}
```

#### 5.6.2 监控大盘（管理员视图）

**实时指标**

- 当日安装次数
- 待审核队列深度 + 平均等待时间
- 被阻断的高风险引入次数
- 各风险等级 Skill 占比

**趋势报表**

- Skill 使用趋势（周/月）
- 最受欢迎的 Skill Top 10
- 安全事件统计
- 审核员工作量统计

**告警规则**

```yaml
告警配置:
  - name: 高风险Skill引入激增
    condition: F级Skill数量 > 3 in 1小时
    severity: critical
    notify: security-team-channel

  - name: 未授权Skill安装尝试
    condition: 检测到 ~/.claude/skills/ 非法写入
    severity: high
    notify: security-team-channel

  - name: 审核SLA超时
    condition: B级Skill等待 > 24h OR C级Skill等待 > 48h
    severity: medium
    notify: reviewer-channel
```

---

## 6. 安全机制设计

### 6.1 多层拦截防线

```
防线 1（最外层）：网络层
  Claude Code 出站请求只允许到企业内部 SkillsHub 域名
  通过企业网络策略/防火墙实现

防线 2：CLI 强制路由
  覆盖 Claude Code 的默认 Skill 安装行为
  所有安装请求强制路由到 SkillsHub API

防线 3：文件系统权限
  ~/.claude/skills/approved/ 目录对开发者只读
  只有 skillshub-agent 进程可写入

防线 4：CLAUDE.md 策略
  向 Claude 注入拒绝未授权 Skill 的指令
  作为语义层软拦截

防线 5（最内层）：inotify 监控
  监控 skills 目录的文件变化
  非法写入立即告警 + 可选自动删除
```

### 6.2 认证与授权

**认证机制**

- 支持企业 SSO（OIDC/SAML）
- 本地账号（仅用于无 SSO 场景）
- API Token（供 CI/CD 系统使用，支持细粒度权限）
- CLI 登录使用 OAuth Device Flow，无需明文密码

**权限矩阵**

| 操作 | Developer | Reviewer | Admin |
|------|-----------|----------|-------|
| 搜索/浏览 Skill | ✅ | ✅ | ✅ |
| 安装已审核 Skill | ✅ | ✅ | ✅ |
| 申请引入开源 Skill | ✅ | ✅ | ✅ |
| 发布内部 Skill | ❌ | ❌ | ✅ |
| 审核 Skill | ❌ | ✅ | ✅ |
| 查看扫描详情 | ❌ | ✅ | ✅ |
| 管理用户权限 | ❌ | ❌ | ✅ |
| 修改安全策略 | ❌ | ❌ | ✅ |
| 查看审计日志 | 仅自己 | 仅自己 | 全部 |

### 6.3 数据安全

- Skill 文件存储加密（AES-256）
- 传输层 TLS 1.3
- 敏感配置（数据库密码等）通过环境变量或 Vault 管理，不写入代码
- 审计日志不可删除（append-only 存储策略）

---

## 7. API 接口规范

### 7.1 通用规范

- **Base URL**：`https://skills.your-company.com/api/v1`
- **认证**：`Authorization: Bearer <jwt-token>`
- **内容类型**：`Content-Type: application/json`
- **错误格式**：

```json
{
  "error": {
    "code": "SKILL_NOT_FOUND",
    "message": "Skill 'unknown-skill' 不存在",
    "details": {}
  }
}
```

### 7.2 核心接口列表

#### Skills 查询

```http
# 搜索 Skill
GET /skills?q={keyword}&category={cat}&minScore={score}&page={n}&limit={n}

# 获取 Skill 详情
GET /skills/{skillId}

# 获取 Skill 版本列表
GET /skills/{skillId}/versions

# 下载 Skill 文件包
GET /skills/{skillId}/download?version={ver}
```

#### Skills 安装管理

```http
# 安装 Skill（由 CLI 调用）
POST /skills/{skillId}/install
Body: { "version": "1.2.0", "targetPath": "/home/user/.claude/skills/" }

# 查询安装状态
GET /installs/{installId}

# 卸载记录
DELETE /skills/{skillId}/install
```

#### 开源引入

```http
# 申请引入开源 Skill
POST /imports
Body: {
  "source": "github | npm | url",
  "identifier": "https://github.com/org/skill-name",
  "requestNote": "用于处理 PDF 报告生成"
}

# 查询引入申请状态
GET /imports/{importId}

# 取消引入申请
DELETE /imports/{importId}
```

#### 扫描结果

```http
# 获取扫描报告
GET /scans/{scanId}

# 获取扫描报告（Skill 维度）
GET /skills/{skillId}/scans/latest
```

#### 审核（Reviewer/Admin）

```http
# 获取待审核列表
GET /reviews?status=pending&assignee=me

# 提交审核决策
PUT /reviews/{reviewId}
Body: {
  "decision": "approve | reject | request_changes",
  "comment": "审核意见..."
}
```

#### 管理接口（Admin）

```http
# 发布内部 Skill
POST /skills
Content-Type: multipart/form-data
Body: skill_package (ZIP file)

# 更新安全策略配置
PUT /settings/security-policy

# 获取审计日志
GET /audit-logs?startTime=&endTime=&userId=&eventType=
```

### 7.3 Webhook 事件

SkillsHub 支持向企业系统推送事件通知：

```json
// Webhook 请求体格式
{
  "event": "scan.completed",
  "timestamp": "2026-03-20T10:30:00Z",
  "data": {
    "skillName": "excel-processor",
    "scanId": "scan-789",
    "riskLevel": "B",
    "score": 78,
    "requiresReview": true
  }
}
```

**支持的事件类型**

- `skill.published` —— 新 Skill 发布
- `import.requested` —— 开源引入申请提交
- `scan.completed` —— 扫描完成
- `review.completed` —— 审核完成
- `skill.installed` —— Skill 被安装
- `security.alert` —— 安全告警触发

---

## 8. 数据模型设计

### 8.1 核心数据表

#### skills 表

```sql
CREATE TABLE skills (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(100) NOT NULL UNIQUE,
  display_name  VARCHAR(200),
  description   TEXT NOT NULL,
  category      VARCHAR(50),
  tags          TEXT[],
  source_type   VARCHAR(20) NOT NULL,  -- 'internal' | 'opensource'
  source_url    TEXT,
  author        VARCHAR(200),
  license       VARCHAR(50),
  status        VARCHAR(20) NOT NULL DEFAULT 'active',  -- 'active' | 'deprecated' | 'blocked'
  install_count INTEGER DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### skill_versions 表

```sql
CREATE TABLE skill_versions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  skill_id        UUID NOT NULL REFERENCES skills(id),
  version         VARCHAR(20) NOT NULL,
  changelog       TEXT,
  file_path       TEXT NOT NULL,     -- MinIO/S3 路径
  file_hash       VARCHAR(64) NOT NULL,  -- SHA256
  is_latest       BOOLEAN DEFAULT false,
  is_stable       BOOLEAN DEFAULT false,
  scan_id         UUID REFERENCES scans(id),
  published_by    UUID REFERENCES users(id),
  published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(skill_id, version)
);
```

#### scans 表

```sql
CREATE TABLE scans (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  skill_version_id UUID REFERENCES skill_versions(id),
  status          VARCHAR(20) NOT NULL,  -- 'pending' | 'running' | 'completed' | 'failed'
  risk_level      CHAR(1),              -- 'A' | 'B' | 'C' | 'D' | 'F'
  score           SMALLINT,             -- 0-100
  layer1_result   JSONB,
  layer2_result   JSONB,
  layer3_result   JSONB,
  layer4_result   JSONB,
  summary         TEXT,
  started_at      TIMESTAMPTZ,
  completed_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### reviews 表

```sql
CREATE TABLE reviews (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  scan_id         UUID NOT NULL REFERENCES scans(id),
  assignee_id     UUID REFERENCES users(id),
  status          VARCHAR(20) NOT NULL DEFAULT 'pending',
  decision        VARCHAR(20),   -- 'approved' | 'rejected' | 'escalated'
  comment         TEXT,
  due_at          TIMESTAMPTZ,
  reviewed_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### audit_logs 表

```sql
CREATE TABLE audit_logs (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_type  VARCHAR(50) NOT NULL,
  actor_id    UUID,
  actor_meta  JSONB,    -- IP、设备信息
  resource    JSONB,    -- 操作对象信息
  result      VARCHAR(20),
  metadata    JSONB,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  -- 注意：此表不设 UPDATE/DELETE 权限
);
```

---

## 9. 部署与运维

### 9.1 部署架构

**最小化部署（小型团队 < 50人）**

```yaml
# docker-compose.yml 关键服务
services:
  api:
    image: skillshub/api:latest
    ports: ["3000:3000"]
    environment:
      - DATABASE_URL=postgresql://...
      - REDIS_URL=redis://redis:6379
      - MINIO_ENDPOINT=minio:9000

  web:
    image: skillshub/web:latest
    ports: ["80:80"]

  postgres:
    image: postgres:15-alpine
    volumes: ["postgres_data:/var/lib/postgresql/data"]

  redis:
    image: redis:7-alpine

  minio:
    image: minio/minio:latest
    command: server /data
    volumes: ["minio_data:/data"]

  scanner-worker:
    image: skillshub/scanner:latest
    # 扫描工作进程，独立部署

  sandbox:
    image: skillshub/sandbox:latest
    # 沙箱执行环境，最高安全隔离
    security_opt: ["no-new-privileges:true"]
    cap_drop: ["ALL"]
    network_mode: none
```

**生产环境部署（大型团队 > 200人）**

- API 服务：3 个节点，Nginx 负载均衡
- Scanner Worker：2 个节点，支持并发扫描
- 数据库：PostgreSQL 主从复制
- Redis：Sentinel 高可用模式
- MinIO：分布式模式 4 节点

### 9.2 企业统一部署方案

**通过 MDM / 配置管理工具（Ansible/Chef）统一下发**

```bash
# 下发到开发者机器的配置脚本
#!/bin/bash
SKILLSHUB_REGISTRY="https://skills.your-company.com"

# 1. 安装 CLI
npm install -g @skillshub/cli@latest

# 2. 配置注册中心
skillshub config set registry $SKILLSHUB_REGISTRY

# 3. SSO 登录（自动打开浏览器）
skillshub login --sso

# 4. 初始化文件系统保护
skillshub init --enforce

# 5. 注入全局 CLAUDE.md 策略
skillshub policy inject --global
```

### 9.3 环境变量配置

```bash
# .env.production
NODE_ENV=production
PORT=3000

# 数据库
DATABASE_URL=postgresql://user:pass@host:5432/skillshub

# Redis
REDIS_URL=redis://:password@host:6379

# 文件存储
STORAGE_TYPE=s3           # s3 | minio | local
S3_BUCKET=skillshub-files
S3_REGION=ap-east-1

# 认证
JWT_SECRET=<random-256bit>
OIDC_ISSUER=https://sso.your-company.com
OIDC_CLIENT_ID=skillshub
OIDC_CLIENT_SECRET=<secret>

# 扫描配置
SCAN_SANDBOX_TIMEOUT=120000   # ms
SCAN_MAX_CONCURRENT=5
AUTO_APPROVE_MIN_SCORE=85     # A级自动通过阈值

# 通知
SLACK_WEBHOOK_URL=https://hooks.slack.com/...
EMAIL_SMTP_HOST=smtp.your-company.com

# 外部 Skill 源
GITHUB_TOKEN=<pat>            # 提高 GitHub API 速率限制
ALLOWED_IMPORT_SOURCES=github,npm  # 允许的外部源
```

---

## 10. 非功能性需求

### 10.1 性能需求

| 指标 | 目标值 |
|------|--------|
| API 响应时间（P95） | < 300ms |
| Skill 搜索响应时间 | < 200ms |
| Layer 1-2 扫描耗时 | < 60s |
| Layer 3 沙箱测试耗时 | < 120s |
| 完整扫描流水线 | < 5min |
| 并发扫描任务数 | ≥ 10 |
| 系统可用性 | 99.5% |

### 10.2 安全需求

- 所有通信强制 HTTPS（TLS 1.2+）
- 沙箱环境完全网络隔离
- 审计日志不可篡改
- 敏感数据静态加密
- 定期安全渗透测试（季度）
- 依赖包定期漏洞扫描

### 10.3 可扩展性需求

- 扫描规则支持插件化扩展（自定义检测规则）
- 支持接入更多开源 Skill 源（通过 Source Adapter 接口）
- 通知渠道可扩展（Slack / 企业微信 / 钉钉 / Email）
- 扫描工作节点支持水平扩展

---

## 11. 里程碑与迭代计划

### Phase 1：MVP（第 1-4 周）

**目标**：核心流程跑通，最小可用

- [x] 私有仓库基础 CRUD（上传、存储、下载）
- [x] Layer 1 + Layer 2 静态扫描
- [x] 基础审核工作台（通过/拒绝）
- [x] SkillsHub CLI（install / list / search）
- [x] 文件系统权限锁定
- [x] 基础 Web 管理台
- [x] JWT 认证

**验收标准**：开发者可通过 CLI 安装私有仓库中的 Skill，未经审核的 Skill 无法被安装。

### Phase 2：安全强化（第 5-8 周）

**目标**：完善安全扫描能力

- [ ] Layer 3 沙箱动态测试
- [ ] Layer 4 供应链溯源分析
- [ ] 综合风险评分模型
- [ ] 开源自动引入流程
- [ ] 审核 SLA 管理 + 超时告警
- [ ] CLAUDE.md 策略注入

**验收标准**：开源 Skill 引入全流程跑通，沙箱扫描可正确识别高风险行为。

### Phase 3：企业级特性（第 9-12 周）

**目标**：满足大型团队需求

- [ ] SSO / OIDC 集成
- [ ] 审计日志与合规报告
- [ ] 监控大盘与告警
- [ ] Webhook 事件推送
- [ ] 批量部署脚本（MDM 支持）
- [ ] 版本锁定与团队策略管理
- [ ] 高可用部署方案

**验收标准**：通过企业安全审计，支持 200+ 人团队稳定使用。

### Phase 4：智能化（第 13-16 周）

**目标**：提升自动化程度，降低人工审核负担

- [ ] 基于历史数据的扫描规则自学习
- [ ] 相似 Skill 推荐（避免重复引入）
- [ ] 自动生成审核摘要（AI 辅助）
- [ ] Skill 使用行为分析与异常检测
- [ ] 开放 API 生态（第三方扫描插件）

---

## 12. 风险与约束

### 12.1 技术风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| 沙箱逃逸漏洞 | 低 | 高 | 使用 gVisor / seccomp，定期安全审计 |
| 扫描误报率高影响效率 | 中 | 中 | 持续调优规则，提供人工复核通道 |
| 提示词注入检测不完全 | 高 | 高 | 多层防御，不依赖单一检测手段 |
| 开源仓库 API 限流 | 中 | 低 | 缓存搜索结果，使用认证 Token |

### 12.2 运营约束

- 审核员需要具备基本的安全知识，需提供培训
- 引入流程平均耗时 5-30 分钟，紧急需求需有快速通道
- 沙箱测试消耗较多服务器资源，需预留计算资源

### 12.3 已知局限性

- **语义层扫描局限**：无法 100% 检测所有形式的提示词注入（语言的歧义性）
- **行为测试覆盖率**：沙箱测试用例难以覆盖所有运行场景
- **私有/混淆代码**：对故意混淆的恶意脚本检测能力有限

---

## 13. 附录

### 13.1 术语表

| 术语 | 定义 |
|------|------|
| Skill | Claude Code 中用于扩展 AI 能力的功能模块，核心是 SKILL.md 文件 |
| SKILL.md | Skill 的核心描述文件，包含 YAML frontmatter 和指令正文 |
| Frontmatter | SKILL.md 头部的 YAML 块，包含 name 和 description |
| 提示词注入 | 通过 Skill 内容操纵 AI 执行非预期指令的攻击手段 |
| 供应链攻击 | 通过污染依赖（Skill）来间接攻击使用者的手段 |
| SLA | Service Level Agreement，本文中指审核完成的时间承诺 |
| gVisor | Google 开源的容器沙箱运行时，提供更强的内核级隔离 |

### 13.2 参考资料

- Claude Code 官方文档：https://docs.anthropic.com/en/docs/claude-code/overview
- Anthropic Skills 机制说明：https://docs.claude.com
- OWASP LLM Top 10（LLM 安全最佳实践）
- npm 私有仓库搭建参考（Verdaccio）
- gVisor 容器安全隔离文档

### 13.3 变更记录

| 版本 | 日期 | 变更说明 | 作者 |
|------|------|----------|------|
| v1.0.0 | 2026-03-20 | 初版文档 | — |
