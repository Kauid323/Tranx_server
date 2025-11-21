# 快速开始指南

## 🚀 快速启动

### Windows 用户
1. 双击运行 `start.bat`
2. 访问 http://localhost:4999/health 测试服务器

### Linux/Mac 用户
```bash
chmod +x start.sh
./start.sh
```

## 📝 生成测试数据

```bash
go run seed_data.go
```

这将创建：
- 5 个示例板块
- 20+ 个示例帖子
- 100+ 个示例评论

## 🔍 测试 API

### 使用 PowerShell（Windows）
```powershell
# 健康检查
Invoke-RestMethod -Uri "http://localhost:4999/health" -Method Get

# 获取所有板块
Invoke-RestMethod -Uri "http://localhost:4999/api/boards" -Method Get

# 获取帖子列表
Invoke-RestMethod -Uri "http://localhost:4999/api/posts?sort=hot" -Method Get
```

### 使用 curl（Linux/Mac）
```bash
# 健康检查
curl http://localhost:4999/health

# 获取所有板块
curl http://localhost:4999/api/boards

# 获取帖子列表
curl "http://localhost:4999/api/posts?sort=hot"
```

## 📖 完整文档

- `README.md` - 项目介绍和 API 文档
- `API_TESTS.md` - 详细的 API 测试示例
- `DEVELOPMENT.md` - 开发指南和代码规范

## 🎯 核心功能

✅ 板块管理（创建、编辑、删除）
✅ 帖子发布（支持图片 URL）
✅ 评论系统（楼层、楼主标识）
✅ 互动功能（点赞、收藏、投币）
✅ 多种排序（最新、最热、最近回复）
✅ 分页支持
✅ JSON API 响应

## 🔧 配置

服务器默认运行在端口 4999。如需修改，可以：

1. 设置环境变量：
```bash
# Windows PowerShell
$env:SERVER_PORT="5000"

# Linux/Mac
export SERVER_PORT=5000
```

2. 或创建配置文件（参考 config.env.example）

## 📊 项目结构

```
TaruApp/
├── main.go              # 主程序
├── config/              # 配置模块
├── database/            # 数据库模块
├── models/              # 数据模型
├── handlers/            # API 处理器
├── middleware/          # 中间件
├── utils/               # 工具函数
├── start.bat           # Windows 启动脚本
├── start.sh            # Linux/Mac 启动脚本
└── seed_data.go        # 测试数据生成器
```

## ❓ 问题排查

### 端口被占用
修改端口号（见上方配置说明）

### 依赖下载失败
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

### 数据库文件权限问题
```bash
chmod 666 taruapp.db
```

## 📞 获取帮助

查看详细文档：
- API 使用：`README.md`
- 开发指南：`DEVELOPMENT.md`
- 测试示例：`API_TESTS.md`

