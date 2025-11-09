# Bilibili

这是一个使用Go语言开发的项目。

## 目录结构

```
.
├── api/              # API相关代码
├── cmd/              # 程序入口
│   └── app/          # 主应用程序
├── configs/          # 配置文件
├── deployments/      # 部署相关文件
├── docs/             # 文档
├── examples/         # 示例代码
├── go.mod            # Go模块文件
├── internal/         # 私有应用程序代码
│   ├── handlers/     # HTTP处理器
│   └── services/     # 业务逻辑层
├── pkg/              # 可被外部引用的公共代码
│   ├── file/         # 文件处理（Excel、CSV）
│   └── utils/        # 工具类
└── scripts/          # 脚本文件
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