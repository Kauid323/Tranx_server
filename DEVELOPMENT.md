# TaruApp 开发指南

## 项目结构说明

```
TaruApp/
├── main.go                 # 主程序入口，路由配置
├── go.mod                  # Go 模块依赖管理
├── go.sum                  # 依赖版本锁定文件
├── seed_data.go            # 示例数据生成器
├── start.bat               # Windows 启动脚本
├── start.sh                # Linux/Mac 启动脚本
├── .env.example            # 环境变量配置示例
├── .gitignore              # Git 忽略文件配置
├── README.md               # 项目说明文档
├── API_TESTS.md            # API 测试文档
├── DEVELOPMENT.md          # 开发文档（本文件）
│
├── config/                 # 配置模块
│   └── config.go           # 配置加载和管理
│
├── database/               # 数据库模块
│   └── database.go         # 数据库初始化、表创建、索引
│
├── models/                 # 数据模型模块
│   └── models.go           # 数据结构定义、请求/响应模型
│
├── handlers/               # 处理器模块（业务逻辑）
│   ├── board.go            # 板块相关处理器
│   ├── post.go             # 帖子相关处理器
│   └── comment.go          # 评论相关处理器
│
├── middleware/             # 中间件模块
│   └── middleware.go       # 日志、CORS、限流、错误处理
│
└── utils/                  # 工具函数模块
    └── utils.go            # 通用工具函数

```

## 开发环境搭建

### 1. 安装 Go
下载并安装 Go 1.21 或更高版本：https://golang.org/dl/

### 2. 克隆项目（或创建项目）
```bash
# 如果是新项目
mkdir TaruApp
cd TaruApp

# 初始化 Go 模块
go mod init TaruApp
```

### 3. 安装依赖
```bash
go mod download
```

### 4. 配置环境变量（可选）
```bash
# 复制环境变量示例文件
cp .env.example .env

# 编辑 .env 文件，根据需要修改配置
```

### 5. 运行项目
```bash
# 直接运行
go run main.go

# 或使用启动脚本
# Windows
start.bat

# Linux/Mac
chmod +x start.sh
./start.sh
```

## 开发流程

### 1. 添加新的 API 端点

#### 步骤 1: 定义数据模型（models/models.go）
```go
// 请求模型
type CreateItemRequest struct {
    Name string `json:"name" binding:"required"`
    // ...
}

// 响应模型
type Item struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
    // ...
}
```

#### 步骤 2: 创建处理器函数（handlers/item.go）
```go
func CreateItem(c *gin.Context) {
    var req models.CreateItemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.Response{
            Code:    400,
            Message: "请求参数错误: " + err.Error(),
        })
        return
    }
    
    // 业务逻辑...
    
    c.JSON(http.StatusOK, models.Response{
        Code:    200,
        Message: "创建成功",
        Data:    item,
    })
}
```

#### 步骤 3: 注册路由（main.go）
```go
items := api.Group("/items")
{
    items.POST("", handlers.CreateItem)
    items.GET("", handlers.GetItems)
    // ...
}
```

### 2. 数据库操作

#### 查询单条记录
```go
var item Item
err := database.DB.QueryRow(
    "SELECT id, name FROM items WHERE id = ?", 
    id,
).Scan(&item.ID, &item.Name)

if err == sql.ErrNoRows {
    // 记录不存在
}
```

#### 查询多条记录
```go
rows, err := database.DB.Query(
    "SELECT id, name FROM items WHERE category = ?",
    category,
)
defer rows.Close()

var items []Item
for rows.Next() {
    var item Item
    rows.Scan(&item.ID, &item.Name)
    items = append(items, item)
}
```

#### 插入记录
```go
result, err := database.DB.Exec(
    "INSERT INTO items (name) VALUES (?)",
    name,
)
id, _ := result.LastInsertId()
```

#### 更新记录
```go
_, err := database.DB.Exec(
    "UPDATE items SET name = ? WHERE id = ?",
    name, id,
)
```

#### 删除记录
```go
_, err := database.DB.Exec(
    "DELETE FROM items WHERE id = ?",
    id,
)
```

### 3. 添加中间件

在 `middleware/middleware.go` 中添加新的中间件：
```go
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理
        
        c.Next()
        
        // 后置处理
    }
}
```

在 `main.go` 中使用：
```go
r.Use(middleware.CustomMiddleware())
```

## 代码规范

### 1. 命名规范
- **包名**：小写，简短，不使用下划线
- **文件名**：小写，使用下划线分隔（如：user_handler.go）
- **函数名**：大驼峰（公开）或小驼峰（私有）
- **变量名**：小驼峰
- **常量名**：大驼峰或全大写+下划线

### 2. 注释规范
```go
// CreateBoard 创建板块
// 参数：
//   - c: Gin 上下文
// 返回：
//   - JSON 响应，包含创建的板块 ID
func CreateBoard(c *gin.Context) {
    // ...
}
```

### 3. 错误处理
```go
// 检查错误
if err != nil {
    c.JSON(http.StatusInternalServerError, models.Response{
        Code:    500,
        Message: "操作失败: " + err.Error(),
    })
    return
}

// 检查记录是否存在
if err == sql.ErrNoRows {
    c.JSON(http.StatusNotFound, models.Response{
        Code:    404,
        Message: "记录不存在",
    })
    return
}
```

### 4. 响应格式
统一使用 `models.Response` 结构：
```go
c.JSON(http.StatusOK, models.Response{
    Code:    200,
    Message: "操作成功",
    Data:    data,
})
```

## 测试

### 1. 生成测试数据
```bash
go run seed_data.go
```

### 2. 单元测试
创建测试文件（如：`handlers/board_test.go`）：
```go
package handlers

import (
    "testing"
)

func TestCreateBoard(t *testing.T) {
    // 测试逻辑
}
```

运行测试：
```bash
go test ./...
```

### 3. API 测试
参考 `API_TESTS.md` 文档，使用 curl 或 Postman 进行测试。

## 性能优化

### 1. 数据库优化
- 为常用查询字段添加索引
- 使用预编译语句避免 SQL 注入
- 批量操作使用事务

### 2. 缓存策略
- 对热门数据使用内存缓存
- 设置合理的缓存过期时间

### 3. 分页优化
- 限制最大分页大小
- 使用游标分页代替偏移分页（大数据量）

## 部署

### 1. 编译
```bash
# Windows
go build -o taruapp.exe main.go

# Linux/Mac
go build -o taruapp main.go
```

### 2. 运行
```bash
# Windows
taruapp.exe

# Linux/Mac
./taruapp
```

### 3. 后台运行（Linux）
```bash
nohup ./taruapp > taruapp.log 2>&1 &
```

### 4. 使用 systemd（Linux）
创建服务文件 `/etc/systemd/system/taruapp.service`：
```ini
[Unit]
Description=TaruApp Community Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/taruapp
ExecStart=/opt/taruapp/taruapp
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl start taruapp
sudo systemctl enable taruapp
```

## 常见问题

### 1. 端口被占用
修改 `.env` 文件中的 `SERVER_PORT` 配置，或设置环境变量：
```bash
export SERVER_PORT=5000
```

### 2. 数据库文件权限问题
确保应用有读写数据库文件的权限：
```bash
chmod 666 taruapp.db
```

### 3. CORS 跨域问题
在 `.env` 中设置：
```
ENABLE_CORS=true
```

### 4. 依赖下载失败
设置 Go 代理：
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

## 扩展功能建议

### 1. 用户认证系统
- JWT Token 认证
- 用户注册/登录
- 权限管理

### 2. 文件上传
- 图片上传到本地/云存储
- 文件大小限制
- 格式验证

### 3. 搜索功能
- 全文搜索
- 标签搜索
- 高级筛选

### 4. 通知系统
- 评论通知
- 点赞通知
- 系统公告

### 5. 数据统计
- 用户活跃度
- 内容热度分析
- 数据导出

### 6. 缓存层
- Redis 集成
- 热点数据缓存
- Session 管理

### 7. 消息队列
- 异步任务处理
- 邮件发送
- 数据同步

## 参考资源

- [Gin 官方文档](https://gin-gonic.com/docs/)
- [Go 官方文档](https://golang.org/doc/)
- [SQLite 文档](https://www.sqlite.org/docs.html)
- [Go 数据库操作](https://golang.org/pkg/database/sql/)

## 联系方式

如有问题或建议，请联系开发团队。

