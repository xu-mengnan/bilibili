# 文档总览

欢迎查阅Bilibili评论爬取工具的文档！本页面提供所有文档的索引和简介。

## 📚 文档目录

### 🚀 快速开始

- **[README.md](../README.md)** - 项目主页
  - 项目简介和功能概览
  - 快速开始指南
  - 基础使用方法
  - 主要功能说明

### 📖 用户文档

#### 功能说明
- **[评论排序模式与子评论抓取](comment_sort_mode.md)**
  - 按时间/热度排序评论
  - 子评论抓取功能
  - Web界面使用教程
  - API使用示例
  - 推荐 ⭐️ 首次使用必读

#### API参考
- **[API参考文档](api-reference.md)**
  - 完整的HTTP API接口说明
  - 请求参数详解
  - 响应格式说明
  - 错误代码参考
  - 完整使用流程示例
  - 推荐 ⭐️ API开发必备

### 🛠️ 开发文档

#### 开发指南
- **[开发指南](development-guide.md)**
  - 开发环境配置
  - 项目结构说明
  - 代码规范和提交规范
  - 开发流程和最佳实践
  - 测试指南
  - 调试技巧
  - 推荐 ⭐️ 贡献代码必读

#### 故障排查
- **[故障排查指南](troubleshooting.md)**
  - 常见问题诊断
  - 错误解决方案
  - 性能优化建议
  - 调试技巧
  - 推荐 ⭐️ 遇到问题时查阅

### 📝 变更日志

- **[变更日志目录](../changelogs/README.md)** - 所有版本更新记录

#### 最新版本
- **[v1.2.0 - 子评论抓取功能](../changelogs/2026-01-12-sub-comments-feature.md)** (2026-01-12)
  - 新增子评论抓取
  - 层级化展示
  - 点击展开/折叠交互
  - Excel导出层级标识

- **[v1.1.0 - 评论排序模式功能](../changelogs/2026-01-11-comment-sort-mode.md)** (2026-01-11)
  - 新增按时间/热度排序
  - Web界面排序选择器
  - 函数式选项模式

### 🤖 AI助手文档

- **[CLAUDE.md](../CLAUDE.md)** - Claude Code工作指南
  - 项目架构说明
  - 构建和运行命令
  - 核心组件概览
  - 关键设计模式

---

## 📋 文档类型说明

### 按受众分类

| 受众 | 推荐文档 |
|------|---------|
| **普通用户** | README.md → comment_sort_mode.md → troubleshooting.md |
| **API开发者** | README.md → api-reference.md → troubleshooting.md |
| **贡献者** | README.md → development-guide.md → CLAUDE.md |
| **问题排查** | troubleshooting.md → development-guide.md |

### 按主题分类

| 主题 | 相关文档 |
|------|---------|
| **安装和运行** | README.md, development-guide.md |
| **功能使用** | comment_sort_mode.md, README.md |
| **API调用** | api-reference.md, comment_sort_mode.md |
| **代码开发** | development-guide.md, CLAUDE.md |
| **问题解决** | troubleshooting.md |
| **版本历史** | changelogs/ |

---

## 🎯 快速导航

### 我想要...

#### 开始使用项目
1. 阅读 [README.md](../README.md) 了解基本功能
2. 按照"快速开始"部分安装和运行
3. 访问Web界面开始使用

#### 通过API使用
1. 阅读 [API参考文档](api-reference.md)
2. 查看"完整使用流程示例"
3. 参考 [功能说明文档](comment_sort_mode.md) 了解参数

#### 开发新功能
1. 阅读 [开发指南](development-guide.md)
2. 了解项目结构和代码规范
3. 查看 [CLAUDE.md](../CLAUDE.md) 了解架构
4. 参考现有代码实现

#### 解决问题
1. 查看 [故障排查指南](troubleshooting.md)
2. 搜索对应的错误类型
3. 按照解决方案操作
4. 如果无法解决，提交Issue

#### 了解更新内容
1. 查看 [变更日志目录](../changelogs/README.md)
2. 阅读对应版本的变更日志
3. 了解新功能和改进

---

## 📖 阅读建议

### 新用户学习路径

```
第1天：了解项目
├─ README.md (10分钟)
└─ 运行项目并测试 (20分钟)

第2天：深入功能
├─ comment_sort_mode.md (15分钟)
├─ 测试不同排序模式 (30分钟)
└─ 尝试导出数据 (10分钟)

第3天（可选）：API使用
├─ api-reference.md (20分钟)
└─ 编写API调用脚本 (40分钟)
```

### 开发者学习路径

```
第1天：环境和架构
├─ README.md (10分钟)
├─ CLAUDE.md (15分钟)
└─ development-guide.md - 项目结构部分 (20分钟)

第2天：开发规范
├─ development-guide.md - 代码规范部分 (20分钟)
├─ 阅读核心代码 (60分钟)
└─ 运行测试 (15分钟)

第3天：实践开发
├─ development-guide.md - 开发流程部分 (15分钟)
├─ 尝试修改代码 (90分钟)
└─ troubleshooting.md - 调试部分 (20分钟)
```

---

## 🔄 文档更新

### 最近更新

- **2026-01-12**: 新增API参考文档、开发指南、故障排查指南
- **2026-01-12**: 新增v1.2.0变更日志（子评论功能）
- **2026-01-11**: 新增v1.1.0变更日志（排序模式功能）
- **2026-01-11**: 更新功能说明文档

### 文档维护

所有文档遵循以下原则：
- ✅ 保持最新：随代码更新同步更新文档
- ✅ 准确清晰：示例代码经过测试验证
- ✅ 循序渐进：从简单到复杂，便于学习
- ✅ 问题导向：围绕用户实际问题编写

---

## 📧 反馈和贡献

### 文档问题

如果发现文档中的问题：
- 错误信息
- 过时内容
- 不清晰的表述
- 缺失的内容

请通过以下方式反馈：
1. 提交Issue标注 `documentation` 标签
2. 提交Pull Request修正
3. 在相关文档中留言

### 贡献文档

欢迎贡献文档改进：
1. Fork项目
2. 创建文档分支
3. 编写或改进文档
4. 提交Pull Request

文档贡献指南参见 [开发指南](development-guide.md)。

---

## 📚 外部资源

### 相关技术文档

- [Go官方文档](https://golang.org/doc/)
- [Gin框架文档](https://gin-gonic.com/docs/)
- [Bilibili API文档](https://github.com/SocialSisterYi/bilibili-API-collect)
- [Excelize文档](https://xuri.me/excelize/)
- [Chart.js文档](https://www.chartjs.org/docs/)

### 相关项目

- [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) - Bilibili API接口文档
- [excelize](https://github.com/qax-os/excelize) - Go Excel库

---

## ❓ 常见问题

### 找不到需要的文档？

1. 使用Ctrl+F在本页搜索关键词
2. 查看 [故障排查指南](troubleshooting.md) 的目录
3. 在GitHub仓库搜索相关Issue
4. 提交Issue说明需要的文档类型

### 文档中的示例无法运行？

1. 确认Go版本 >= 1.19
2. 检查依赖是否完整：`go mod download`
3. 查看 [故障排查指南](troubleshooting.md)
4. 确认示例代码中的参数（如视频BV号）是否有效

### 想要更深入的技术细节？

1. 阅读源代码：核心逻辑在 `pkg/bilibili/` 和 `internal/`
2. 查看 [CLAUDE.md](../CLAUDE.md) 了解架构设计
3. 查看测试文件（`*_test.go`）了解使用方法
4. 在Issue中提问具体问题

---

**祝你使用愉快！如有问题，请随时查阅相关文档或提交Issue。** 📖✨
