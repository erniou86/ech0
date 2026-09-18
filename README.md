# Ech0

极简自托管发布平台：碎片记录、标签、搜索、RSS。单二进制部署，内存占用极低。
**适用人群**：追求轻量自托管的个人写作者、极客与数字花园爱好者，希望用最小成本长期运行个人记录服务的用户。

## 功能特性

- 碎片 CRUD（文字/链接/待办）
- 标签分类与全文搜索
- RSS 2.0 输出
- 内置极简 Web UI（零前端构建）
- SQLite 嵌入式存储，单文件数据库

## 技术栈与目录结构

**技术栈**：Go 1.22 / SQLite（mattn/go-sqlite3）/ 内置零构建 Web UI

```
ech0/
├── main.go        # 入口（HTTP 服务 + SQLite 存储）
├── main_test.go   # 单元测试
├── go.mod
├── Dockerfile
└── LICENSE
```

## 快速开始

```bash
# 本地运行
go run main.go
# http://localhost:8080

# 构建单二进制
go build -o ech0 .
./ech0

# 运行测试
go test -v ./...
```

**Docker 部署**

```bash
docker build -t ech0 .
docker run -p 8080:8080 -v ech0-data:/data ech0
```

**环境变量**

| 变量 | 默认 | 说明 |
|------|------|------|
| `PORT` | 8080 | 监听端口 |
| `DB_PATH` | ./data.db | SQLite 文件路径 |

**部署说明**：可编译为单一可执行文件部署到任意 Linux 服务器/VPS，或使用 Docker 容器化运行；SQLite 单文件数据库便于备份迁移。

## 验证状态

引用 AI Factory 全套件验证报告（[VERIFICATION.md](../../VERIFICATION.md)，2026-09-18）：

- 仓库提供 go.mod / main.go / main_test.go / Dockerfile / README / LICENSE，代码完整
- 本机未安装 Go 工具链，未做本地编译；在具备 Go 1.22+ 的环境可直接 `go test ./...` 验证（main_test.go 含单元测试）
- Docker 部署路径：`docker build -t ech0 . && docker run -p 8080:8080 ech0`

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/notes?q=&tag=&page=&size=` | 列表（搜索/标签/分页） |
| POST | `/api/notes` | 创建 `{content, tags}` |
| GET | `/api/notes/:id` | 详情 |
| DELETE | `/api/notes/:id` | 删除 |
| GET | `/api/rss` | RSS feed |
| GET | `/health` | 健康检查 |

## License

MIT License，详见 [LICENSE](LICENSE)。本项目代码与文档由 AI 辅助生成，仅供参考与学习使用。

## 支持项目

如果这个项目对你有帮助，欢迎赞助支持持续开发：

[![PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat-square&logo=paypal)](https://paypal.me/Junlong439)
