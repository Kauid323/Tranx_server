# TaruApp - 社区服务器后台

基于 Golang 的社区论坛后台系统，提供板块管理、帖子发布、评论互动等完整功能。

## 功能特性

### 核心功能
- 🏢 **板块管理**：创建、编辑、删除板块
- 📝 **帖子系统**：发布、编辑、删除帖子，支持图片URL
- 💬 **评论系统**：多级评论，楼层标记，楼主标识
- 👍 **互动功能**：点赞、收藏、投币
- 📊 **统计功能**：浏览量、评论数、互动数据统计

### 高级特性
- 🔥 **智能排序**：
  - 帖子排序：最新发布、最近回复、热门（综合权重）
  - 评论排序：默认（楼层正序）、点赞最高、楼主发布、倒序
- 📄 **分页支持**：所有列表接口均支持分页
- 🎯 **筛选功能**：按板块筛选帖子
- ⚡ **性能优化**：数据库索引优化，高效查询

## 技术栈

- **语言**：Go 1.21+
- **框架**：Gin Web Framework
- **数据库**：SQLite3
- **API格式**：JSON

## 快速开始

### 安装依赖

```bash
go mod download
```

### 运行服务器

```bash
go run main.go
```

服务器将在 `http://localhost:4999` 启动

### 健康检查

```bash
curl http://localhost:4999/health
```

## API 文档

### 基础响应格式

所有 API 响应均为 JSON 格式：

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {}
}
```

### 板块 API

#### 1. 创建板块
- **接口**: `POST /api/boards`
- **请求体**:
```json
{
  "name": "技术交流",
  "description": "技术相关的讨论板块"
}
```

#### 2. 获取所有板块
- **接口**: `GET /api/boards`
- **响应**: 返回所有板块列表

#### 3. 获取板块详情
- **接口**: `GET /api/boards/:id`
- **响应**: 返回指定板块的详细信息

#### 4. 更新板块
- **接口**: `PUT /api/boards/:id`
- **请求体**: 同创建板块

#### 5. 删除板块
- **接口**: `DELETE /api/boards/:id`

#### 6. 获取板块统计
- **接口**: `GET /api/stats/boards/:id`
- **响应**: 返回板块的帖子数、总浏览量、总评论数

### 帖子 API

#### 1. 创建帖子
- **接口**: `POST /api/posts`
- **请求体**:
```json
{
  "board_id": 1,
  "title": "帖子标题",
  "content": "帖子内容",
  "publisher": "用户名",
  "image_url": "https://example.com/image.jpg"
}
```

#### 2. 获取帖子列表（支持筛选和排序）
- **接口**: `GET /api/posts`
- **查询参数**:
  - `board_id`: 板块ID（可选）
  - `sort`: 排序方式
    - `latest`: 最新发布
    - `reply`: 最近回复
    - `hot`: 热门（综合权重计算）
  - `page`: 页码（默认1）
  - `page_size`: 每页数量（默认20，最大100）

**示例**:
```bash
# 获取板块1的最新帖子，第1页，每页20条
GET /api/posts?board_id=1&sort=latest&page=1&page_size=20

# 获取所有板块的热门帖子
GET /api/posts?sort=hot
```

#### 3. 获取帖子详情
- **接口**: `GET /api/posts/:id`
- **说明**: 会自动增加浏览量

#### 4. 更新帖子
- **接口**: `PUT /api/posts/:id`
- **请求体**: 同创建帖子（不含board_id和publisher）

#### 5. 删除帖子
- **接口**: `DELETE /api/posts/:id`
- **说明**: 会同时删除该帖子的所有评论

#### 6. 点赞帖子
- **接口**: `POST /api/posts/:id/like`
- **响应**: 返回更新后的点赞数

#### 7. 收藏帖子
- **接口**: `POST /api/posts/:id/favorite`
- **响应**: 返回更新后的收藏数

#### 8. 投币帖子
- **接口**: `POST /api/posts/:id/coin`
- **请求体**:
```json
{
  "amount": 2
}
```
- **说明**: 投币数量1-10，默认为1

#### 9. 获取帖子统计
- **接口**: `GET /api/stats/posts/:id`
- **响应**: 返回点赞、收藏、投币、评论、浏览等统计数据

### 评论 API

#### 1. 创建评论
- **接口**: `POST /api/comments`
- **请求体**:
```json
{
  "post_id": 1,
  "content": "评论内容",
  "publisher": "用户名"
}
```
- **说明**: 会自动判断是否为楼主，自动分配楼层号

#### 2. 获取评论列表（支持多种排序）
- **接口**: `GET /api/comments`
- **查询参数**:
  - `post_id`: 帖子ID（必需）
  - `sort`: 排序方式
    - `default`: 默认（楼层正序）
    - `likes`: 点赞最高
    - `author`: 楼主发布（楼主评论优先）
    - `desc`: 倒序（楼层倒序）
  - `page`: 页码（默认1）
  - `page_size`: 每页数量（默认50，最大200）

**示例**:
```bash
# 获取帖子1的评论，按点赞排序
GET /api/comments?post_id=1&sort=likes&page=1&page_size=50

# 获取帖子1的评论，倒序查看
GET /api/comments?post_id=1&sort=desc
```

#### 3. 更新评论
- **接口**: `PUT /api/comments/:id`
- **请求体**:
```json
{
  "content": "更新后的评论内容"
}
```

#### 4. 删除评论
- **接口**: `DELETE /api/comments/:id`
- **说明**: 会自动更新帖子的评论计数

#### 5. 点赞评论
- **接口**: `POST /api/comments/:id/like`
- **响应**: 返回更新后的点赞数

## 数据模型

### 板块 (Board)
- `id`: 板块ID
- `name`: 板块名称
- `description`: 板块描述
- `created_at`: 创建时间
- `updated_at`: 更新时间

### 帖子 (Post)
- `id`: 帖子ID
- `board_id`: 所属板块ID
- `title`: 帖子标题
- `content`: 帖子内容
- `publisher`: 发布者
- `publish_time`: 发布时间
- `coins`: 投币数
- `favorites`: 收藏数
- `likes`: 点赞数
- `image_url`: 图片URL
- `comment_count`: 评论数
- `view_count`: 浏览数
- `last_reply_time`: 最后回复时间
- `created_at`: 创建时间
- `updated_at`: 更新时间

### 评论 (Comment)
- `id`: 评论ID
- `post_id`: 所属帖子ID
- `content`: 评论内容
- `publisher`: 评论者
- `publish_time`: 评论时间
- `likes`: 点赞数
- `is_author`: 是否为楼主
- `floor`: 楼层号
- `created_at`: 创建时间
- `updated_at`: 更新时间

## 扩展功能

### 热门算法
帖子热度计算公式：
```
热度 = 点赞数 × 3 + 收藏数 × 2 + 投币数 × 5 + 评论数 × 2 + 浏览数
```

### 数据库索引
系统已为以下字段创建索引以提升查询性能：
- 帖子的板块ID、发布时间、最后回复时间、点赞数
- 评论的帖子ID、点赞数、楼层号

### 分页机制
- 所有列表接口均支持分页
- 默认页码为1
- 帖子列表默认每页20条（最大100条）
- 评论列表默认每页50条（最大200条）

## 项目结构

```
TaruApp/
├── main.go              # 主程序入口
├── go.mod              # Go模块依赖
├── database/
│   └── database.go     # 数据库初始化和连接
├── models/
│   └── models.go       # 数据模型定义
├── handlers/
│   ├── board.go        # 板块处理器
│   ├── post.go         # 帖子处理器
│   └── comment.go      # 评论处理器
└── taruapp.db          # SQLite数据库文件（运行后生成）
```

## 使用示例

### 创建一个完整的论坛流程

1. **创建板块**
```bash
curl -X POST http://localhost:4999/api/boards \
  -H "Content-Type: application/json" \
  -d '{"name":"技术交流","description":"技术讨论区"}'
```

2. **发布帖子**
```bash
curl -X POST http://localhost:4999/api/posts \
  -H "Content-Type: application/json" \
  -d '{
    "board_id": 1,
    "title": "Go语言学习心得",
    "content": "分享一些Go语言的学习经验...",
    "publisher": "张三",
    "image_url": "https://example.com/go.png"
  }'
```

3. **查看帖子列表**
```bash
curl "http://localhost:4999/api/posts?board_id=1&sort=latest&page=1"
```

4. **发表评论**
```bash
curl -X POST http://localhost:4999/api/comments \
  -H "Content-Type: application/json" \
  -d '{
    "post_id": 1,
    "content": "写得不错！",
    "publisher": "李四"
  }'
```

5. **点赞帖子**
```bash
curl -X POST http://localhost:4999/api/posts/1/like
```

6. **查看评论（按点赞排序）**
```bash
curl "http://localhost:4999/api/comments?post_id=1&sort=likes"
```

## 开发说明

### 环境要求
- Go 1.21 或更高版本
- SQLite3

### 编译运行
```bash
# 开发模式
go run main.go

# 编译为可执行文件
go build -o taruapp.exe main.go

# 运行编译后的文件
./taruapp.exe
```

## 许可证

MIT License

## 作者

TaruApp Team

