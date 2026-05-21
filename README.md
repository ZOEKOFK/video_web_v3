# video_web_v3

一个基于 Go 语言构建的视频社交平台后端服务。

## 功能特性

- **用户服务**: 用户注册、登录、信息管理
- **视频服务**: 视频上传、播放、管理
- **社交服务**: 关注、粉丝、社交互动
- **聊天服务**: WebSocket 实时通信、离线消息推送

## 技术栈

- **框架**: Hertz
- **语言**: Go 1.26+
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+
- **对象存储**: MinIO
- **RPC**: gRPC
- **服务发现**: Consul

## 快速开始

```bash
# 克隆项目
git clone https://github.com/ZOEKOFK/video_web_v3.git
cd video_web_v3

# 启动依赖服务
docker-compose up -d

# 启动 API Gateway
cd api_gateway && go run .

# 启动 App Service（新终端）
cd app && go run .
```

## 目录结构

```
video_web_v3/
├── api_gateway/           # API 网关
│   ├── biz/handler/       # HTTP 处理器
│   ├── client/            # gRPC 客户端
│   ├── idl/               # Proto 定义
│   ├── my_jwt/            # JWT 认证
│   └── router/            # 路由配置
└── app/                   # 后端应用
    ├── adapter/           # 适配器层
    │   ├── grpc/          # gRPC 服务实现
    │   ├── persistence/   # 数据持久化
    │   └── consul/        # 服务发现
    ├── domain/            # 领域层
    │   ├── model/         # 数据模型
    │   ├── repository/    # 仓储接口
    │   └── service_logic/ # 业务逻辑
    ├── usecase/           # 用例层
    └── idl/               # Proto 定义
```

