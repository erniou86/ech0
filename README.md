# Ech0

极简自托管发布平台：碎片记录、标签、搜索、RSS。单二进制部署，内存占用极低。

## 功能
- 碎片 CRUD（文字/链接/待办）
- 标签分类与全文搜索
- RSS 2.0 输出
- 内置极简 Web UI（零前端构建）
- SQLite 嵌入式存储，单文件数据库

## 快速开始
```bash
# 本地运行
go run main.go
# http://localhost:8080

# 构建单二进制
go build -o ech0 .
./ech0
```

## Docker
```bash
docker build -t ech0 .
docker run -p 8080:8080 -v ech0-data:/data ech0
```

## 测试
```bash
go test -v ./...
```

## API
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/notes?q=&tag=&page=&size=` | 列表（搜索/标签/分页） |
| POST | `/api/notes` | 创建 `{content, tags}` |
| GET | `/api/notes/:id` | 详情 |
| DELETE | `/api/notes/:id` | 删除 |
| GET | `/api/rss` | RSS feed |
| GET | `/health` | 健康检查 |

## 环境变量
| 变量 | 默认 | 说明 |
|------|------|------|
| `PORT` | 8080 | 监听端口 |
| `DB_PATH` | ./data.db | SQLite 文件路径 |

## License
MIT
