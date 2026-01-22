# 变更摘要 - 全新设计版本与 V2 API

## 日期
2026-01-23

## 描述

创建了一个全新的设计版本，采用 "Cyber-Aqua Glass" 美学风格，并配套开发了专用的 V2 API 接口。新版本与老版本并存，用户可以自由切换对比。

### 设计理念
- **美学风格**: Cyber-Aqua Glass（赛博-水溶玻璃）
- **核心色调**: 青色/蓝绿色渐变替代原版紫色
- **视觉效果**: 玻璃拟态 + 动态网格背景 + 发光阴影
- **设计原则**: 不受老版本兼容性限制，采用最合理的设计方案

### 主要变更

#### 1. 新设计系统

**配色方案:**
```css
/* 主色调 - 青色渐变 */
--primary-light: #22D3EE
--primary: #06B6D4
--primary-dark: #0891B2
--primary-darker: #0E7490

/* 玻璃拟态效果 */
--glass-bg: rgba(255, 255, 255, 0.85)
--glass-border: rgba(255, 255, 255, 0.3)
--glass-blur: blur(20px)
```

**动态背景:**
- 4 个径向渐变网格
- 20 秒循环的平移动画
- 多色渐变：青色/紫色/粉色混合

**字体系统:**
- 显示字体: Outfit (Google Fonts)
- 正文字体: Plus Jakarta Sans (Google Fonts)

**动画系统:**
| 动画 | 时长 | 用途 |
|------|------|------|
| fadeInUp | 0.6s | 页面元素进入 |
| shimmer | 3s | Logo 光泽效果 |
| meshMove | 20s | 背景网格移动 |
| pulse | 1.5s | 运行状态指示 |
| spin | 0.8s | 加载旋转 |

#### 2. V2 API 设计

创建专用的 V2 API (`/api/v2/*`)，采用更简洁的设计：

| 特性 | 老API | 新API (V2) |
|------|---------------|-----------------|
| 响应格式 | `{data: {...}}` | 直接返回数组/对象 |
| 任务列表 | `{tasks: [...]}` | `[...]` |
| 模板列表 | `{templates: [...]}` | `[...]` |
| SSE格式 | `event: content\ndata: "JSON"` | `data: 文本内容` |
| 完成信号 | `event: done\ndata: ` | `data: [DONE]` |

**新端点:**
```
GET  /api/v2/tasks          获取所有任务（直接数组）
GET  /api/v2/tasks/:id      获取单个任务详情
GET  /api/v2/templates      获取所有模板（直接数组）
POST /api/v2/analyze-stream 流式分析（简化SSE）
POST /api/v2/preview        预览Prompt
```

#### 3. 页面布局

**侧边栏:**
- 260px 宽度，固定定位
- 玻璃拟态效果（半透明 + 背景模糊）
- Logo 带光动画效果
- 导航项悬停时滑动 + 渐变背景

**评论爬取页面:**
- Hero 区域：大标题 + 渐变文字
- 输入卡片：大输入框 + 图标装饰
- 进度卡片：3列统计 + 动画进度条
- 结果卡片：图表 + 可展开评论卡片

**任务管理页面:**
- 4列统计卡片：图标 + 渐变数值
- 任务列表：状态徽章 + 元信息
- Modal 详情：分组信息展示

**AI分析页面:**
- 左侧配置面板：步骤编号卡片
- 右侧结果面板：实时预览 + Markdown渲染
- 流式显示：纯文本进度 + 自动滚动

#### 4. 交互改进

**按钮效果:**
- 主按钮：渐变背景 + 光晕阴影
- 悬停：向上抬起 + 加强光晕
- 点击：缩小动画反馈

**卡片效果:**
- 悬停：轻微抬起 + 阴影加深
- 玻璃边框：半透明白色

**进度条:**
- 渐变填充
- 流光动画（白色条纹从左到右）

**状态徽章:**
- 运行中：黄色背景 + 脉冲圆点
- 已完成：绿色背景 + 静态圆点
- 失败：红色背景 + 静态圆点

#### 5. 评论卡片设计

**主评论:**
- 圆形头像 + 紫色边框
- 用户名 + 等级徽章（渐变背景）
- 点赞数（红心图标）+ 时间
- 内容文本
- 展开按钮（箭头图标 + 回复数）

**子评论:**
- 左侧紫色竖线
- 较小头像
- 嵌套式布局

#### 6. 图表样式

**时间分布图:**
- 面积图（渐变填充）
- 圆点标记 + 白边
- 隐藏网格线

**点赞分布图:**
- 柱状图
- 渐变色柱体（随高度变化）
- 圆角顶部

## 修改的文件

### 新建文件

**后端:**
- `internal/handlers/v2_api.go` - V2 API 处理器（约 350 行）

**前端 - CSS:**
- `static-new/css/style.css` - 主设计系统（约 700 行）
- `static-new/css/app.css` - 评论爬取样式（约 350 行）
- `static-new/css/analysis.css` - AI分析样式（约 500 行）
- `static-new/css/tasks.css` - 任务管理样式（约 370 行）

**前端 - JS:**
- `static-new/js/api.js` - V2 API 客户端（约 185 行）
- `static-new/js/app.js` - 评论爬取逻辑（约 440 行）
- `static-new/js/analysis.js` - AI分析逻辑（约 400 行）
- `static-new/js/tasks.js` - 任务管理逻辑（约 290 行）
- `static-new/js/charts.js` - 图表渲染（约 120 行）

**前端 - HTML:**
- `static-new/index-new.html` - 评论爬取页面
- `static-new/analysis-new.html` - AI分析页面
- `static-new/tasks-new.html` - 任务管理页面

**文档:**
- `docs/NEW_DESIGN.md` - 新版设计完整文档

### 修改文件

**后端:**
- `api/api.go` - 添加 V2 路由组和静态文件路由

## 影响分析

### 正面影响
1. **视觉吸引力大幅提升**: 现代化的玻璃拟态设计
2. **代码结构更清晰**: V2 API 不受历史包袱限制
3. **API 响应更简洁**: 直接返回数组，减少解构操作
4. **SSE 解析更简单**: 不需要处理 event 类型
5. **用户体验更好**: 流畅动画 + 实时反馈

### 无破坏性变更
- 老版本 (`/static/*`) 完全保留
- 老 API (`/api/*`) 完全保留
- 新老版本可以并存使用
- 后端向后兼容

### 兼容性
- 现代浏览器（Chrome 90+、Firefox 88+、Safari 14+、Edge 90+）
- 支持响应式布局（平板、移动端）

## 修复问题

### API 数据解构问题
**问题**: 老版 API 返回 `{tasks: [...]}` 需要手动解构
**解决**: V2 API 直接返回数组 `[...]`

### SSE 格式复杂度
**问题**: 老版 SSE 需要 `event:` + `data:` + JSON 编码
**解决**: V2 SSE 简化为 `data: 内容` 格式

### 完成信号识别
**问题**: 老版完成信号是 `event: done`，前端解析复杂
**解决**: V2 完成信号是 `data: [DONE]`，易于识别

### 字段访问问题
**问题**: CommentData 没有 Level 字段
**解决**: 使用 `Member.LevelInfo.CurrentLevel`

### 类型匹配问题
**问题**: Ctime 是 int 类型，格式化函数需要 int64
**解决**: 添加类型转换 `time.Unix(int64(ts), 0)`

## 技术细节

### SSE 流式格式对比

**老版 API SSE:**
```
event: content
data: "文本内容"

event: done
data:

event: error
data: "错误信息"
```

**V2 API SSE:**
```
data: 文本内容\n\n
data: [DONE]\n\n
data: [ERROR] 错误信息\n\n
```

### 时间戳格式化

V2 API 实现了更友好的时间显示：
- 1小时内: "刚刚"
- 24小时内: "X小时前"
- 30天内: "X天前"
- 1年内: "X个月前"
- 超过1年: "2026-01-02 15:04"

## 访问地址

| 页面 | 老版本 | 新版本 |
|------|--------|--------|
| 评论爬取 | `http://localhost:8080/` | `http://localhost:8080/index-new.html` |
| 任务管理 | `http://localhost:8080/tasks` | `http://localhost:8080/tasks-new.html` |
| AI 分析 | `http://localhost:8080/analysis` | `http://localhost:8080/analysis-new.html` |

## 版本信息

- **设计版本**: v2 - Cyber-Aqua Glass
- **API 版本**: v2
- **CSS 架构**: 玻璃拟态 + CSS 变量
- **响应式**: 是
- **暗色模式**: 否（待添加）

## 后续优化建议

1. **深色模式**: 添加暗色主题支持
2. **主题切换**: 允许用户切换配色方案
3. **侧边栏折叠**: 移动端自适应
4. **加载骨架屏**: 改善首屏加载体验
5. **更多预设模板**: 扩展 AI 分析模板库
6. **导出 PDF**: 支持将分析结果导出为 PDF
7. **PWA 支持**: 添加离线使用能力
8. **快捷键**: 添加键盘快捷键支持

## 相关文档

- [新版设计完整文档](../docs/NEW_DESIGN.md)
- [老版设计变更日志](./2026-01-18-ui-redesign-saas-console.md)
- [AI分析功能变更](./2026-01-17-ai-analysis-feature.md)
