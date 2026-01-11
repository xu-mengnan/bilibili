# 故障排查指南

本文档帮助你快速诊断和解决Bilibili评论爬取工具中的常见问题。

## 目录

- [启动问题](#启动问题)
- [API请求问题](#api请求问题)
- [爬取任务问题](#爬取任务问题)
- [认证问题](#认证问题)
- [数据导出问题](#数据导出问题)
- [Web界面问题](#web界面问题)
- [性能问题](#性能问题)
- [网络问题](#网络问题)

---

## 启动问题

### 问题: 程序无法启动

**症状**:
```
panic: runtime error: ...
```

**可能原因和解决方案**:

1. **端口被占用**
   ```
   Error: listen tcp :8080: bind: address already in use
   ```

   **解决方案**:
   ```bash
   # 查找占用8080端口的进程
   # Linux/Mac
   lsof -i :8080

   # Windows
   netstat -ano | findstr :8080

   # 终止进程或更改端口
   export PORT=3000
   go run ./cmd/app
   ```

2. **Go版本过低**
   ```
   go: go.mod requires go >= 1.19
   ```

   **解决方案**:
   ```bash
   # 检查Go版本
   go version

   # 升级Go到1.19或更高版本
   # 访问 https://golang.org/dl/
   ```

3. **依赖缺失**
   ```
   package github.com/gin-gonic/gin: cannot find package
   ```

   **解决方案**:
   ```bash
   go mod download
   go mod tidy
   ```

### 问题: 静态文件404

**症状**: 访问 `http://localhost:8080` 显示404或样式丢失

**解决方案**:
1. 确认 `static/` 目录存在且包含必要文件
2. 检查工作目录是否正确
   ```bash
   # 应该在项目根目录运行
   pwd  # 或 Windows: cd
   ls static/  # 确认static目录存在
   ```

---

## API请求问题

### 问题: "Invalid request" 错误

**症状**:
```json
{
  "error": "Invalid request: Key: 'ScrapeRequest.VideoID' Error:Field validation for 'VideoID' failed on the 'required' tag"
}
```

**原因**: 必填参数缺失或格式错误

**解决方案**:
```bash
# 确保包含所有必填参数
curl -X POST http://localhost:8080/api/comments/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "video_id": "BV1xx411c7mu",  # 必填
    "page_limit": 2
  }'
```

### 问题: "Invalid sort_mode" 错误

**症状**:
```json
{
  "error": "Invalid sort_mode: must be 'time' or 'hot'"
}
```

**原因**: `sort_mode` 参数值错误

**解决方案**:
```json
{
  "sort_mode": "time"  // 或 "hot"
}
```

### 问题: "task not found" 错误

**症状**:
```json
{
  "error": "task not found"
}
```

**原因**:
- 任务ID错误
- 任务已过期（服务器重启会清空内存中的任务）

**解决方案**:
1. 确认任务ID正确
2. 重新启动爬取任务
3. 考虑实现任务持久化（存储到数据库）

---

## 爬取任务问题

### 问题: 爬取任务一直"running"不完成

**症状**: 任务状态一直显示 "running"，进度不更新

**可能原因**:

1. **网络连接问题**

   **诊断**:
   ```bash
   # 测试Bilibili API连通性
   curl https://api.bilibili.com/x/web-interface/view?bvid=BV1xx411c7mu
   ```

   **解决方案**: 检查网络连接，可能需要代理

2. **视频ID无效**

   **症状**: 日志显示 "invalid video id" 或类似错误

   **解决方案**: 确认视频BV号正确，视频未被删除

3. **API限流**

   **症状**: 进度停滞不前，日志显示429错误

   **解决方案**:
   ```json
   {
     "delay_ms": 1000,  // 增加延迟
     "page_limit": 2    // 减少页数
   }
   ```

4. **程序崩溃**

   **诊断**: 查看控制台是否有panic或错误信息

   **解决方案**: 重启程序，减小 `page_limit`

### 问题: 评论数量比预期少

**可能原因**:

1. **评论确实较少**: 视频评论本身就不多
2. **分页限制**: `page_limit` 设置过小
3. **API权限**: 无认证可能看不到所有评论

**解决方案**:
```json
{
  "page_limit": 10,              // 增加页数
  "auth_type": "cookie",         // 使用认证
  "cookie": "your_sessdata"
}
```

### 问题: 子评论没有抓取到

**症状**: 主评论有回复数（rcount > 0），但 `replies` 为空

**可能原因**:

1. **未开启子评论抓取**
   ```json
   {
     "include_replies": true  // 确保设置为true
   }
   ```

2. **子评论API失败**: 查看日志是否有错误

   **解决方案**: 增加延迟，使用认证

3. **子评论已被删除**: 评论计数未更新但实际已删除

### 问题: 爬取速度太慢

**症状**: 每页需要很长时间

**原因**: 延迟设置过大或网络慢

**解决方案**:
```json
{
  "delay_ms": 300,          // 减小延迟（但不要低于300）
  "include_replies": false  // 关闭子评论抓取
}
```

---

## 认证问题

### 问题: Cookie认证失败

**症状**:
```json
{
  "error": "authentication failed"
}
```

**可能原因和解决方案**:

1. **SESSDATA过期**
   - Cookie有效期通常为1-3个月
   - 重新登录获取新的SESSDATA

2. **SESSDATA格式错误**
   - 确保复制完整的SESSDATA值
   - 不要包含空格或换行符

3. **获取方法**:
   ```
   1. 打开 bilibili.com 并登录
   2. F12 打开开发者工具
   3. Application -> Cookies -> bilibili.com
   4. 找到 SESSDATA 并复制 Value 列的内容
   ```

### 问题: 使用Cookie后仍然看不到更多评论

**原因**:
- Cookie可能无效
- 账号可能被限制

**诊断**:
```bash
# 使用Cookie测试API
curl 'https://api.bilibili.com/x/web-interface/nav' \
  -H 'Cookie: SESSDATA=your_sessdata'

# 检查返回的 data.isLogin 是否为 true
```

---

## 数据导出问题

### 问题: 导出失败

**症状**:
```json
{
  "error": "Failed to export: ..."
}
```

**可能原因**:

1. **任务未完成**
   ```json
   {
     "error": "Task not completed yet"
   }
   ```

   **解决方案**: 等待任务完成再导出

2. **磁盘空间不足**

   **诊断**:
   ```bash
   df -h  # Linux/Mac
   # 检查可用空间
   ```

3. **权限问题**

   **解决方案**:
   ```bash
   # 确保exports目录可写
   mkdir -p exports
   chmod 755 exports
   ```

### 问题: Excel中文乱码

**症状**: 在Excel中打开CSV文件，中文显示为乱码

**原因**: Excel默认使用系统编码，不是UTF-8

**解决方案**:

**方法1**: 使用Excel导入
```
1. Excel -> 数据 -> 从文本/CSV
2. 选择文件
3. 文件原始格式选择 "UTF-8"
4. 点击加载
```

**方法2**: 导出为Excel格式（推荐）
```json
{
  "format": "xlsx"  // 使用xlsx代替csv
}
```

### 问题: 导出文件找不到

**症状**: 导出成功但找不到文件

**解决方案**:
```bash
# 检查exports目录
ls -la exports/

# 文件命名格式: export_{时间戳}_{随机}.{扩展名}
# 例如: export_20260112_103045_abc123.xlsx
```

### 问题: 子评论层级不正确

**症状**: Excel中子评论的"层级"列显示不正确

**原因**: 可能是导出逻辑bug

**诊断**:
- 检查主评论是否标记为"主评论"
- 检查子评论是否标记为"└ 回复 (L1)"

**解决方案**: 如果持续出现，请报告issue

---

## Web界面问题

### 问题: 页面无法加载

**症状**: 白屏或一直加载中

**诊断步骤**:

1. **检查浏览器控制台**（F12）
   - 查看Console是否有JavaScript错误
   - 查看Network是否有失败的请求

2. **检查网络**
   ```bash
   # 测试API是否可访问
   curl http://localhost:8080/api/hello
   ```

3. **检查静态文件**
   ```bash
   # 确认静态文件存在
   ls static/js/
   ls static/css/
   ```

### 问题: 界面功能不正常

**症状**: 按钮点击无反应、数据不更新等

**解决方案**:

1. **清除浏览器缓存**
   - Chrome: Ctrl+Shift+Delete
   - 或者强制刷新: Ctrl+F5

2. **检查浏览器兼容性**
   - 推荐使用最新版Chrome、Firefox、Edge
   - Safari可能存在兼容性问题

3. **检查JavaScript错误**
   - F12 -> Console
   - 查看是否有错误信息

### 问题: 图表不显示

**症状**: 评论时间分布和点赞数分布图表显示空白

**可能原因**:

1. **Chart.js未加载**

   **诊断**: F12 -> Console，查看是否有 "Chart is not defined" 错误

   **解决方案**: 检查 `index.html` 中的CDN链接
   ```html
   <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
   ```

2. **数据为空**

   **原因**: 爬取的评论数太少

   **解决方案**: 增加 `page_limit`

### 问题: 子评论展开/折叠不工作

**症状**: 点击箭头图标没有反应

**诊断**:
1. F12 -> Console 查看是否有错误
2. 确认评论有子评论（replies不为空）

**解决方案**:
- 强制刷新页面（Ctrl+F5）
- 检查 `app.js` 是否正确加载

---

## 性能问题

### 问题: 内存占用过高

**症状**: 程序占用大量内存

**可能原因**:

1. **爬取数据量太大**

   **解决方案**:
   ```json
   {
     "page_limit": 5  // 减少页数限制
   }
   ```

2. **任务未清理**

   **说明**: 所有任务数据存储在内存中

   **临时解决方案**: 定期重启程序

   **长期解决方案**: 实现任务过期清理机制

### 问题: CPU占用过高

**可能原因**:

1. **频繁的API请求**

   **解决方案**: 增加 `delay_ms`

2. **大量数据处理**

   **正常现象**: 导出大量评论时会短暂占用较高CPU

### 问题: 响应缓慢

**症状**: API请求响应时间长

**诊断**:
```bash
# 测试API响应时间
time curl http://localhost:8080/api/comments/result/{task_id}
```

**可能原因**:

1. **数据量大**: 任务包含大量评论

   **解决方案**: 使用 `limit` 参数限制返回数量
   ```bash
   curl "http://localhost:8080/api/comments/result/{task_id}?limit=100"
   ```

2. **磁盘I/O**: 导出文件时写入慢

   **解决方案**: 使用SSD，或减少导出数据量

---

## 网络问题

### 问题: "connection refused" 错误

**症状**:
```
dial tcp 127.0.0.1:8080: connect: connection refused
```

**原因**: 服务器未启动或端口错误

**解决方案**:
```bash
# 确认服务器已启动
ps aux | grep app

# 确认端口正确
netstat -an | grep 8080
```

### 问题: "timeout" 错误

**症状**:
```
context deadline exceeded
```

**可能原因**:

1. **网络不稳定**

   **解决方案**: 检查网络连接

2. **Bilibili API响应慢**

   **解决方案**: 增加超时时间（需要修改代码）
   ```go
   client := &http.Client{
       Timeout: 30 * time.Second,  // 增加超时
   }
   ```

3. **需要代理**

   **解决方案**: 配置HTTP代理
   ```bash
   export HTTP_PROXY=http://proxy:port
   export HTTPS_PROXY=http://proxy:port
   ```

### 问题: "429 Too Many Requests" 错误

**症状**: 日志显示429状态码

**原因**: 触发Bilibili API限流

**解决方案**:

1. **增加延迟**
   ```json
   {
     "delay_ms": 1000  // 增加到1秒
   }
   ```

2. **使用认证**
   ```json
   {
     "auth_type": "cookie",
     "cookie": "your_sessdata"
   }
   ```

3. **减少并发**
   - 避免同时运行多个爬取任务

4. **等待一段时间**
   - 被限流后等待5-10分钟再试

---

## 调试技巧

### 启用详细日志

修改 `cmd/app/main.go`:
```go
import "log"

func main() {
    // 启用详细日志
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    // ... 其他代码
}
```

在关键位置添加日志:
```go
log.Printf("Debug: oid=%d, page=%d, comments=%d\n", oid, page, len(comments))
```

### 使用调试器

```bash
# 使用Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试主程序
dlv debug ./cmd/app

# 设置断点
(dlv) break main.main
(dlv) break pkg/bilibili/comment.go:100

# 运行
(dlv) continue

# 查看变量
(dlv) print variableName
```

### 抓包分析

使用工具如Wireshark、Fiddler、Charles分析HTTP请求:

```bash
# 或使用curl查看完整请求
curl -v https://api.bilibili.com/x/v2/reply/main?oid=123456&type=1
```

---

## 常见错误代码

| 错误代码 | 说明 | 解决方案 |
|---------|------|---------|
| -400 | 请求错误 | 检查参数格式 |
| -403 | 访问权限不足 | 使用Cookie认证 |
| -404 | 资源不存在 | 检查视频ID是否正确 |
| -412 | 请求被拦截 | 增加延迟，使用认证 |
| -509 | 请求过于频繁 | 增加delay_ms，减少page_limit |
| -799 | 服务器内部错误 | 稍后重试 |

---

## 获取帮助

如果以上方案都无法解决问题：

1. **查看日志**: 查看程序输出的错误日志
2. **搜索Issue**: 在GitHub Issues中搜索类似问题
3. **提交Issue**: 提供以下信息：
   - Go版本: `go version`
   - 操作系统: Windows/Linux/Mac
   - 错误信息: 完整的错误日志
   - 复现步骤: 如何触发错误
   - 请求参数: API请求的参数

4. **查看文档**:
   - [README.md](../README.md)
   - [API参考文档](api-reference.md)
   - [开发指南](development-guide.md)

---

## 预防问题

### 最佳实践

1. **合理设置参数**
   ```json
   {
     "page_limit": 2,      // 从小数值开始测试
     "delay_ms": 500,      // 给足够的延迟
     "include_replies": false  // 先不抓子评论
   }
   ```

2. **使用认证**
   - Cookie认证可以获得更高的限额
   - 减少被限流的可能性

3. **监控任务**
   - 定期检查任务进度
   - 及时发现异常情况

4. **保存数据**
   - 及时导出爬取结果
   - 避免数据丢失（服务器重启会清空内存数据）

5. **定期更新**
   - 关注项目更新
   - 及时更新依赖

---

**记住**: 大多数问题都可以通过增加延迟、使用认证、减少请求量来解决！
