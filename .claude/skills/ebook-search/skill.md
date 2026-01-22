# Ebook Search & Download Skill

搜索并下载电子书。

## 参数

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| book_title | string | 是 | 书名 |
| author | string | 否 | 作者（可选，提高匹配精度） |
| save_path | string | 否 | 保存路径（默认：桌面） |

## 使用示例

```
搜索下载《三体》
搜索下载《深入理解计算机系统》，作者：Bryant
搜索下载《Python编程》，保存到：D:\Books
```

## 实现说明

1. 使用 Open Library API 搜索电子书
2. 展示候选列表（最多 10 本，按相似度排序）
3. 用户选择后，使用 openbooks 或其他工具下载
4. 保存到桌面（或用户指定路径）

## API

- Open Library: https://openlibrary.org/search.json?q={query}
- 备用下载工具: openbooks (需单独安装)
