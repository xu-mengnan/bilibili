# Bilibili

这是一个使用Go语言开发的项目。

## 目录结构

```
.
├── api/              # API相关代码
├── cmd/              # 程序入口
│   └── app/          # 主应用程序
├── configs/          # 配置文件
├── examples/         # 示例代码
├── go.mod            # Go模块文件
├── internal/         # 私有应用程序代码
│   ├── handlers/     # HTTP处理器
│   └── services/     # 业务逻辑层
└── pkg/              # 可被外部引用的公共代码
    ├── bilibili/     # Bilibili API接口
    ├── file/         # 文件处理（Excel、CSV）
    └── utils/        # 工具类
```

## 快速开始

### 构建项目

```bash
go build -o bin/app ./cmd/app
```

### 运行项目

```bash
go run ./cmd/app
```

或者

```bash
./bin/app
```

## API接口

- `GET /` - 欢迎页面
- `GET /hello` - 简单的问候接口
- `GET /user/{id}` - 根据ID获取用户信息

## Bilibili API功能

项目提供了访问Bilibili公开API的功能：

- 获取视频信息：使用`pkg/bilibili/video.go`
- 获取用户信息：使用`pkg/bilibili/user.go`
- 获取评论信息：使用`pkg/bilibili/comment.go`

### 示例

获取视频评论数据：
```go
// 获取视频评论 (oid为视频aid, pn为页码, ps为每页数量, next为游标)
comments, err := bilibili.GetComments(123456, 1, 20, 0)
if err != nil {
    log.Fatal("获取评论失败:", err)
}
```

获取用户信息：
```go
// 获取用户信息 (mid为用户ID)
user, err := bilibili.GetUser(123456)
if err != nil {
    log.Fatal("获取用户信息失败:", err)
}
```

获取视频信息：
```go
// 通过BVID获取视频信息
video, err := bilibili.GetVideoByBVID("BV1xx411c7mu")
if err != nil {
    log.Fatal("获取视频信息失败:", err)
}
```

## HTTP客户端优化

为了提高代码复用性和维护性，项目引入了统一的HTTP客户端：

- 统一的HTTP请求处理：使用`pkg/bilibili/client.go`
- 所有Bilibili API调用都通过该客户端发送请求
- 统一设置User-Agent、超时等参数

## 文件处理功能

项目提供了处理Excel和CSV文件的功能：

- Excel文件读写：使用`pkg/file/excel.go`
- CSV文件读写：使用`pkg/file/csv.go`

### 示例

运行文件处理示例：
```bash
go run examples/file_example.go
```

这将创建并读取示例的Excel和CSV文件。

## 配置

配置文件位于 `configs/config.json`，可以修改服务器端口和其他配置项。