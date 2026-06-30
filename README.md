# Fight Game - 格斗游戏服务端

基于微服务架构的实时格斗游戏后端，使用 Go + go-zero 框架构建。

## 架构概览

```
┌─────────────────────────────────────────────────────┐
│                   Gateway (9001)                     │
│   HTTP + WebSocket (9000) + Push gRPC (9014)        │
│   路由器分发 + WebSocket 连接管理 + 会话管理          │
└────┬──────┬──────┬──────┬──────┬──────┬─────────────┘
     │      │      │      │      │      │
     ▼      ▼      ▼      ▼      ▼      ▼
   Auth   Match  Player  Game  Friend  Mail
   (RPC)  (RPC)  (RPC)  (RPC)  (RPC)  (RPC)
```

### 技术栈

| 组件              | 技术选型                            |
| ----------------- | ----------------------------------- |
| **框架**          | go-zero (zRPC + REST)              |
| **通信协议**      | gRPC (服务间) + WebSocket (客户端)  |
| **序列化**        | Protocol Buffers (protobuf v3)      |
| **服务发现**      | etcd                                |
| **数据库**        | MySQL (GORM) / MongoDB / Redis      |
| **认证**          | JWT (golang-jwt)                    |
| **日志/链路**     | go-zero logx + OpenTelemetry        |
| **部署基础设施**  | etcd, Prometheus, Zipkin 等集成     |

## 模块说明

### 🚪 Gateway — 网关（端口 9001 / WebSocket 9000）

统一入口，负责：
- HTTP REST 服务（go-zero rest）
- WebSocket 长连接管理（gorilla/websocket）
- 消息路由分发（基于 `WSMsgType` 枚举路由）
- Push gRPC 服务（端口 9014），供内部服务推送通知到客户端
- 会话管理、心跳保活（Ping/Pong）

### 🔐 Auth — 认证服务（fight.auth）

- 用户注册（`Register`）
- 用户登录（`Login`），签发 JWT Token
- Token 刷新（`RefreshToken`）
- 玩家信息存储于 MySQL

### 👤 Player — 玩家服务（fight.player）

- 玩家资料管理（昵称、头像、签名、等级）
- 货币系统（多种货币类型）
- 背包/资产系统
- 段位/Rating 系统（含对战统计）
- 数据存储于 MongoDB

### 🎯 Match — 匹配服务（fight.match）

基于 Redis ZSet 的竞技匹配队列：
- **竞技匹配** — 基于段位分（ELO）的渐进扩圈策略
- **娱乐匹配** — 宽松分差范围，快速匹配
- 定时扫描器（MatchScanner）周期性触发匹配
- Ticket 机制管理匹配票据

### 🎮 Game — 游戏服务（fight.game）

核心对战逻辑：
- **房间系统** — 房间生命周期管理，多阶段状态机
- **Ban/Pick 流程** — 竞技模式禁用选人阶段
- **战斗系统** — 2D 格斗战斗引擎
  - 60fps 帧同步
  - 物理引擎（重力、碰撞、击退）
  - 攻击判定（轻/重攻击、连携技）
  - 防御机制（格挡减伤）
  - 飞行道具系统
- **断线重连** — 5 秒超时判负机制
- **WebSocket 直连** — 战斗帧同步通信

### 👥 Friend — 好友服务（fight.friend）

- 好友添加/删除
- 好友请求（接收/回复）
- 玩家搜索
- 实时聊天（消息存储于 MongoDB）
- 聊天历史记录

### ✉️ Mail — 邮件服务（fight.mail）

- 邮件列表获取
- 邮件详情、标记已读
- 发送邮件、领取附件、删除邮件
- 新邮件推送通知

## 快速开始

### 环境要求

- Go 1.26+
- protoc + protoc-gen-go + protoc-gen-go-grpc
- MySQL / MongoDB / Redis
- etcd

### 启动顺序

```bash
# 1. 启动基础设施（etcd、MySQL、MongoDB、Redis）

# 2. 启动各微服务（按依赖顺序）
go run service/auth/auth.go -f service/auth/etc/auth.yaml
go run service/player/player.go -f service/player/etc/player.yaml
go run service/match/match.go -f service/match/etc/match.yaml
go run service/game/game.go -f service/game/etc/game.yaml
go run service/friend/friend.go -f service/friend/etc/friend.yaml
go run service/mail/mail.go -f service/mail/etc/mail.yaml

# 3. 启动网关
go run service/gateway/gateway.go -f service/gateway/etc/gateway.yaml
```

### 生成 Proto 代码

```bash
.\secript\proto-generate.bat
```

## 项目结构

```
api/                      # Proto 协议定义
  common/                 #   通用消息定义 + WSMsgType 枚举
  auth/                   #   认证模块协议
  match/                  #   匹配模块协议
  game/                   #   游戏模块协议
  player/                 #   玩家模块协议
  mail/                   #   邮件模块协议
  friend/                 #   好友模块协议
  gateway/                #   网关推送协议

pb/                       # Proto 生成代码
  auth/                   #   auth.pb.go / auth_grpc.pb.go
  match/                  #   ...
  game/
  player/
  mail/
  friend/
  gateway/
  common/

pkg/common/               # 通用工具包
  config/                 #   MySQL/MongoDB/Redis 初始化
  model/                  #   基础模型
  utils/                  #   工具函数（日期、ID、消息、密钥、Token）

service/                  # 微服务实现
  auth/                   #   认证服务
  friend/                 #   好友服务
  game/                   #   游戏服务
  gateway/                #   网关服务
  mail/                   #   邮件服务
  match/                  #   匹配服务
  player/                 #   玩家服务
```

每个微服务遵循统一目录结构：

```
service/{module}/
  etc/{module}.yaml       #   配置文件
  {module}.go             #   入口 main
  internal/
    config/config.go      #   配置结构体
    svc/service_context.go#   服务上下文
    server/               #   gRPC server 实现
    logic/                #   业务逻辑
    model/                #   数据模型
```

## WebSocket 消息路由

基于 protobuf 枚举 `WSMsgType`（定义于 `api/common/common.proto`）：

| 范围       | 模块     | 说明             |
| ---------- | -------- | ---------------- |
| 1000-1999  | Auth     | 登录/注册/Token  |
| 2000-2999  | Match    | 匹配队列          |
| 3000-3999  | Game     | 游戏对战          |
| 4000-4999  | Player   | 玩家信息          |
| 5000-5999  | Mail     | 邮件系统          |
| 6000-6999  | Friend   | 好友系统          |
| 7000-7999  | Chat     | 聊天系统          |
| 8000-8999  | Push     | 服务端推送        |
| 9000-9999  | System   | 心跳/错误         |

## 配置参考

各服务均使用 YAML 配置文件，通过 `-f` 参数指定。核心基础设施依赖：

- **etcd** — `127.0.0.1:2379`
- **MySQL** — 各服务独立 DataSource 配置
- **MongoDB** — 默认连接 URI 配置
- **Redis** — `127.0.0.1:6379`

配置文件统一位于 `service/{module}/etc/{module}.yaml`。
