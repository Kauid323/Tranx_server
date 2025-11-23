# TaruApp API 使用说明

## 概述

TaruApp 是一个基于 Golang 的社区服务器后台，运行在端口 **4999**。所有API响应均为JSON格式。

## 认证机制

- 除了注册和登录API外，**所有API都需要在请求头中携带Token**
- Token 使用 RC4 加密
- Token 有效期为 30 天
- 请求头格式：`Token: <your_token>`

## 用户系统

### 用户等级
- **0**: 普通用户
- **50**: 管理员

### 用户标签
- 管理员可以给用户添加标签
- 标签包含名称和颜色

---

## API 端点

### 1. 用户认证（无需Token）

#### 1.1 用户注册
```http
POST /api/auth/register
Content-Type: application/json
```

**请求体：**
```json
{
  "username": "testuser",      // 必填，3-20个字符
  "password": "12345678",      // 必填，至少8位
  "email": "user@example.com", // 可选，预留邮箱字段
  "avatar": "http://..."       // 可选，头像URL
}
```

**响应：**
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "testuser"
  }
}
```

#### 1.2 用户登录
```http
POST /api/auth/login
Content-Type: application/json
```

**请求体：**
```json
{
  "username": "testuser",
  "password": "12345678"
}
```

**响应：**
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "RC4加密的token字符串",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "user@example.com",
      "level": 0,
      "avatar": "http://...",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "expires_at": "2024-01-31T00:00:00Z"
  }
}
```

---

### 2. 用户信息（需要Token）

#### 2.1 获取当前用户信息
```http
GET /api/me
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取当前用户信息成功",
  "data": {
    "user": { ... },
    "tags": [ ... ]
  }
}
```

#### 2.2 获取指定用户信息
```http
GET /api/users/:id
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取用户信息成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "user@example.com",
      "level": 0,
      "avatar": "http://...",
      "coins": 150,
      "created_at": "2024-01-01T00:00:00Z"
    },
    "tags": [...]
  }
}
```

#### 2.3 获取用户统计信息
```http
GET /api/users/:id/stats
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取用户统计成功",
  "data": {
    "following_count": 10,
    "follower_count": 25,
    "is_following": true
  }
}
```

#### 2.4 获取所有用户列表
```http
GET /api/users?page=1&page_size=20
Token: <your_token>
```

**查询参数：**
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大100）

**响应：**
```json
{
  "code": 200,
  "message": "获取用户列表成功",
  "data": {
    "total": 50,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "id": 1,
        "username": "user1",
        "email": "user1@example.com",
        "level": 50,
        "avatar": "http://...",
        "coins": 200,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      },
      {
        "id": 2,
        "username": "user2",
        "email": "user2@example.com",
        "level": 0,
        "avatar": "http://...",
        "coins": 150,
        "created_at": "2024-01-02T00:00:00Z",
        "updated_at": "2024-01-02T00:00:00Z"
      }
    ]
  }
}
```

**注意：** 用户列表按用户ID升序排列，ID最小的排在第一个

#### 2.5 退出登录
```http
POST /api/logout
Token: <your_token>
```

---

### 3. 关注系统（需要Token）

#### 3.1 关注用户
```http
POST /api/follow/:id
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "关注成功"
}
```

#### 3.2 取消关注用户
```http
DELETE /api/follow/:id
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "取消关注成功"
}
```

#### 3.3 获取关注列表
```http
GET /api/follow/:id/following?page=1&page_size=20
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取关注列表成功",
  "data": {
    "total": 10,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "id": 2,
        "username": "user2",
        "avatar": "http://...",
        "level": 0,
        "coins": 200,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### 3.4 获取粉丝列表
```http
GET /api/follow/:id/followers?page=1&page_size=20
Token: <your_token>
```

---

### 4. 签到系统（需要Token）

#### 4.1 每日签到
```http
POST /api/checkin
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "签到成功",
  "data": {
    "reward": 50,
    "total_coins": 150,
    "check_time": "2024-01-01T08:30:00Z"
  }
}
```

**注意：** 每天只能签到一次，每次签到奖励50硬币，每天0点刷新签到状态

#### 4.2 获取签到状态
```http
GET /api/checkin/status
Token: <your_token>
```

**响应（已签到）：**
```json
{
  "code": 200,
  "message": "今天已签到",
  "data": {
    "checked_in": true,
    "can_check": false,
    "check_time": "2024-01-01T08:30:00Z",
    "reward": 50
  }
}
```

**响应（未签到）：**
```json
{
  "code": 200,
  "message": "今天未签到",
  "data": {
    "checked_in": false,
    "can_check": true
  }
}
```

#### 4.3 获取签到排行榜
```http
GET /api/checkin/rank?page=1&page_size=100
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取签到排行榜成功",
  "data": {
    "total": 50,
    "page": 1,
    "page_size": 100,
    "list": [
      {
        "user_id": 1,
        "username": "earlybird",
        "avatar": "http://...",
        "check_time": "2024-01-01T00:00:15Z",
        "rank": 1
      },
      {
        "user_id": 2,
        "username": "user2",
        "avatar": "http://...",
        "check_time": "2024-01-01T00:01:30Z",
        "rank": 2
      }
    ]
  }
}
```

**注意：** 排行榜按签到时间排序，签到越早排名越靠前

#### 4.4 获取用户签到历史
```http
GET /api/checkin/history/:id?page=1&page_size=30
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取签到历史成功",
  "data": {
    "total": 15,
    "page": 1,
    "page_size": 30,
    "list": [
      {
        "id": 1,
        "user_id": 1,
        "check_date": "2024-01-01",
        "check_time": "2024-01-01T08:30:00Z",
        "reward": 50,
        "created_at": "2024-01-01T08:30:00Z"
      }
    ]
  }
}
```

---

### 5. 板块管理（需要Token）

**注意：** 系统有一个默认主板块（ID=1，名称：综合讨论），用于接收所有未指定板块的帖子。

#### 5.1 创建板块
```http
POST /api/boards/create
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "name": "技术讨论",
  "description": "讨论各种技术问题"
}
```

#### 5.2 获取所有板块
```http
GET /api/boards/list
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取板块列表成功",
  "data": [
    {
      "id": 1,
      "name": "综合讨论",
      "description": "默认主板块，所有话题都可以在这里讨论",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "name": "技术讨论",
      "description": "讨论各种技术问题",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 5.3 获取板块详情
```http
GET /api/boards/:id
Token: <your_token>
```

#### 5.4 更新板块
```http
PUT /api/boards/:id
Token: <your_token>
Content-Type: application/json
```

#### 5.5 删除板块
```http
DELETE /api/boards/:id
Token: <your_token>
```

---

### 6. 帖子管理（需要Token）

#### 6.1 创建帖子
```http
POST /api/posts/create
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "board_id": 1,
  "title": "帖子标题",
  "content": "帖子内容",
  "type": "text",             // 可选，帖子类型："text"(普通文本，默认) 或 "markdown"(Markdown格式)
  "image_url": "http://..."   // 可选
}
```

**注意：** 发布者信息从Token中自动获取

#### 6.2 获取帖子列表
```http
GET /api/posts/list?board_id=1&sort=latest&page=1&page_size=20
Token: <your_token>
```

**查询参数：**
- `board_id`: 板块ID（可选）
- `sort`: 排序方式
  - `latest`: 最新发布（默认）
  - `reply`: 最近回复
  - `hot`: 热门（综合点赞、收藏、投币等）
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大100）

**响应：**
```json
{
  "code": 200,
  "message": "获取帖子列表成功",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "id": 1,
        "board_id": 1,
        "user_id": 1,
        "title": "帖子标题",
        "content": "帖子内容",
        "type": "text",              // 帖子类型："text" 或 "markdown"
        "publisher": "testuser",
        "publish_time": "2024-01-01T00:00:00Z",
        "coins": 10,
        "favorites": 5,
        "likes": 20,
        "image_url": "http://...",
        "attachment_url": "",        // 预留字段，用于APK等文件
        "attachment_type": "",       // 附件类型
        "comment_count": 15,
        "view_count": 100,
        "last_reply_time": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
```

#### 6.3 获取帖子详情
```http
GET /api/posts/:id
Token: <your_token>
```

#### 6.4 更新帖子
```http
PUT /api/posts/:id
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "title": "更新后的标题",
  "content": "更新后的内容",
  "type": "markdown",        // 可选，帖子类型："text" 或 "markdown"
  "image_url": "http://..."  // 可选
}
```

**权限说明：** 只有帖子作者本人才能编辑自己的帖子

#### 6.5 删除帖子
```http
DELETE /api/posts/:id
Token: <your_token>
```

**权限说明：** 
- 只有帖子作者本人才能删除自己的帖子
- 删除帖子会级联删除：
  - 该帖子的所有评论（包括楼中楼回复）
  - 所有相关的点赞记录
  - 所有相关的收藏记录
  - 所有相关的浏览历史记录

**响应：**
```json
{
  "code": 200,
  "message": "删除帖子成功"
}
```

**错误响应：**
```json
{
  "code": 403,
  "message": "无权删除此帖子，只能删除自己的帖子"
}
```

#### 6.6 点赞/取消点赞帖子（切换功能）
```http
POST /api/posts/:id/like
Token: <your_token>
```

**功能说明：**
- 如果用户未点赞该帖子，则执行点赞操作
- 如果用户已点赞该帖子，则执行取消点赞操作
- 支持用户给自己的帖子点赞

**响应：**
```json
{
  "code": 200,
  "message": "点赞成功",  // 或 "取消点赞成功"
  "data": {
    "likes": 15,        // 帖子当前总点赞数
    "is_liked": true    // 当前用户是否已点赞该帖子
  }
}
```

#### 6.7 收藏帖子
```http
POST /api/posts/:id/favorite
Token: <your_token>
```

#### 6.8 投币帖子
```http
POST /api/posts/:id/coin
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "amount": 1   // 投币数量，1-10
}
```

**功能说明：**
- 投币者的硬币会被扣除
- 帖子作者会获得相应数量的硬币
- 不能给自己的帖子投币
- 投币数量范围：1-10
- 硬币不足时无法投币

**响应：**
```json
{
  "code": 200,
  "message": "投币成功",
  "data": {
    "coins": 15,        // 帖子当前总投币数
    "user_coins": 135   // 投币者剩余硬币数
  }
}
```

---

### 7. 评论管理（需要Token）

#### 7.1 创建评论
```http
POST /api/comments/create
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "post_id": 1,
  "parent_id": null,      // 可选，楼中楼回复时填写父评论ID
  "content": "评论内容"
}
```

**注意：** 
- 评论者信息从Token中自动获取，系统会自动判断是否为楼主
- 支持楼中楼回复：设置 `parent_id` 为父评论ID即可回复指定评论
- 顶级评论的 `parent_id` 为 `null`，楼中楼回复的楼层号为0

#### 7.2 获取评论列表（顶级评论）
```http
GET /api/comments/list?post_id=1&sort=default&page=1&page_size=50
Token: <your_token>
```

**查询参数：**
- `post_id`: 帖子ID（必填）
- `sort`: 排序方式
  - `default`: 默认正序（按楼层）
  - `likes`: 点赞最高
  - `author`: 楼主发布优先
  - `desc`: 倒序（按楼层倒序）
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认50，最大200）

**说明：** 此API只返回顶级评论（parent_id为null的评论），每个评论包含 `reply_count` 字段表示子回复数量

**响应示例：**
```json
{
  "code": 200,
  "message": "获取评论列表成功",
  "data": {
    "total": 20,
    "page": 1,
    "page_size": 50,
    "list": [
      {
        "id": 1,
        "post_id": 1,
        "user_id": 2,
        "parent_id": null,
        "content": "这是一条顶级评论",
        "publisher": "testuser",
        "avatar": "http://example.com/avatar.jpg",
        "publish_time": "2024-01-01T12:30:00Z",
        "likes": 5,
        "coins": 2,
        "is_author": false,
        "floor": 1,
        "reply_count": 3
      }
    ]
  }
}
```

#### 7.3 获取评论的子回复列表
```http
GET /api/comments/:id/replies?page=1&page_size=20
Token: <your_token>
```

**查询参数：**
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大100）

**说明：** 获取指定评论的所有子回复，按时间正序排列

#### 7.4 更新评论
```http
PUT /api/comments/:id
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "content": "更新后的评论内容"
}
```

**权限说明：** 只有评论作者本人才能编辑自己的评论

#### 7.5 删除评论
```http
DELETE /api/comments/:id
Token: <your_token>
```

**权限说明：** 
- 只有评论作者本人才能删除自己的评论
- 删除评论会级联删除：
  - 该评论的所有子回复（楼中楼回复）
  - 所有相关的点赞记录
  - 自动更新父评论的回复数（如果是楼中楼回复）
  - 自动更新帖子的评论总数

**响应：**
```json
{
  "code": 200,
  "message": "删除评论成功",
  "data": {
    "deleted_replies": 3  // 同时删除的子回复数量
  }
}
```

**错误响应：**
```json
{
  "code": 403,
  "message": "无权删除此评论，只能删除自己的评论"
}
```

#### 7.6 点赞/取消点赞评论（切换功能）
```http
POST /api/comments/:id/like
Token: <your_token>
```

**功能说明：**
- 如果用户未点赞该评论，则执行点赞操作
- 如果用户已点赞该评论，则执行取消点赞操作
- 支持用户给自己的评论点赞
- 点赞记录存储在 `comment_likes` 表

**响应：**
```json
{
  "code": 200,
  "message": "点赞成功",  // 或 "取消点赞成功"
  "data": {
    "likes": 8,         // 评论当前总点赞数
    "is_liked": true    // 当前用户是否已点赞该评论
  }
}
```

#### 7.7 投币评论
```http
POST /api/comments/:id/coin
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "amount": 2
}
```

**功能说明：**
- 投币者的硬币会被扣除
- 评论作者会获得相应数量的硬币
- 不能给自己的评论投币
- 投币数量范围：1-10
- 硬币不足时无法投币

**响应：**
```json
{
  "code": 200,
  "message": "投币成功",
  "data": {
    "coins": 15,        // 评论当前总投币数
    "user_coins": 135   // 投币者剩余硬币数
  }
}
```

---

### 8. 管理员功能（需要Token + 管理员权限）

#### 8.1 设置用户等级
```http
PUT /api/admin/users/:id/level
Token: <admin_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "level": 50   // 0=普通用户, 50=管理员
}
```

#### 8.2 创建用户标签
```http
POST /api/admin/users/tags
Token: <admin_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "user_id": 1,
  "tag_name": "活跃用户",
  "tag_color": "#FF5733"
}
```

#### 8.3 删除用户标签
```http
DELETE /api/admin/users/tags/:id
Token: <admin_token>
```

---

### 9. 统计信息（需要Token）

#### 9.1 获取板块统计
```http
GET /api/stats/boards/:id
Token: <your_token>
```

#### 9.2 获取帖子统计
```http
GET /api/stats/posts/:id
Token: <your_token>
```

---

## 错误码说明

- **200**: 成功
- **400**: 请求参数错误
- **401**: 未授权（Token无效或未提供）
- **403**: 权限不足
- **404**: 资源不存在
- **500**: 服务器内部错误

---

## 文件上传预留字段

帖子模型中已预留以下字段用于文件上传功能：
- `attachment_url`: 附件URL（如APK文件）
- `attachment_type`: 附件类型（如"apk", "zip"等）

后续可扩展文件上传API。

---

## 使用示例

### 完整流程示例

1. **注册用户**
```bash
curl -X POST http://localhost:4999/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"12345678"}'
```

2. **登录获取Token**
```bash
curl -X POST http://localhost:4999/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"12345678"}'
```

3. **使用Token访问API**
```bash
curl -X GET http://localhost:4999/api/boards/list \
  -H "Token: <your_token_here>"
```

4. **创建帖子**
```bash
# 创建普通文本帖子
curl -X POST http://localhost:4999/api/posts/create \
  -H "Token: <your_token_here>" \
  -H "Content-Type: application/json" \
  -d '{"board_id":1,"title":"测试帖子","content":"这是内容","type":"text"}'

# 创建Markdown格式帖子
curl -X POST http://localhost:4999/api/posts/create \
  -H "Token: <your_token_here>" \
  -H "Content-Type: application/json" \
  -d '{"board_id":1,"title":"Markdown帖子","content":"# 标题\n这是**加粗**文本","type":"markdown"}'
```

---

## 注意事项

1. 所有时间格式均为 ISO 8601 标准
2. Token 需要保存好，有效期30天
3. 密码至少8位，注册时会自动验证
4. 用户名长度3-20个字符
5. 所有需要认证的API都必须在请求头中携带Token
6. 管理员操作需要用户等级为50
7. 签到每天只能一次，每天0点刷新，每次奖励50硬币和25经验
8. 签到排行榜按当天签到时间排序，越早排名越靠前
9. 硬币系统用于投币帖子等功能
10. 关注/粉丝功能支持分页查询
11. 发帖子每次奖励5经验，不限制次数
12. 帖子支持两种类型：普通文本(text)和Markdown格式(markdown)，默认为text

---

## 用户等级系统

### 等级计算规则

用户等级基于经验值（exp）计算，采用平方根公式：

**等级公式：** `Lv = floor(sqrt(exp / 100)) + 1`

### 等级对应经验值表

| 等级 | 所需总经验值 | 该等级经验范围 |
|------|------------|--------------|
| Lv1  | 0          | 0 - 99       |
| Lv2  | 100        | 100 - 399    |
| Lv3  | 400        | 400 - 899    |
| Lv4  | 900        | 900 - 1599   |
| Lv5  | 1600       | 1600 - 2499  |
| Lv6  | 2500       | 2500 - 3599  |
| Lv7  | 3600       | 3600 - 4899  |
| Lv8  | 4900       | 4900 - 6399  |
| Lv9  | 6400       | 6400 - 8099  |
| Lv10 | 8100       | 8100+        |

### 获取经验值的方式

1. **每日签到**：+25 经验值
   - 每天只能签到一次
   - 每天0点刷新

2. **发布帖子**：+5 经验值
   - 不限制次数
   - 每发布一个帖子即可获得

### 用户信息中的等级字段

所有返回用户信息的API都包含以下字段：

```json
{
  "id": 1,
  "username": "testuser",
  "level": 0,           // 权限等级：0-普通用户, 50-管理员
  "user_level": 5,      // 用户等级：Lv1, Lv2, Lv3...
  "exp": 1800,          // 当前总经验值
  "coins": 200,         // 硬币数量
  "avatar": "http://...",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### 示例

**签到响应示例：**
```json
{
  "code": 200,
  "message": "签到成功",
  "data": {
    "reward_coins": 50,
    "reward_exp": 25,
    "total_coins": 150,
    "total_exp": 425,
    "user_level": 3,
    "check_time": "2024-01-01T08:30:00Z"
  }
}
```

**发帖响应示例：**
```json
{
  "code": 200,
  "message": "创建帖子成功",
  "data": {
    "id": 123,
    "reward_exp": 5,
    "total_exp": 430,
    "user_level": 3
  }
}
```

### 升级提示

当用户的经验值达到下一等级所需经验时，系统会自动更新用户等级（user_level）。客户端可以根据返回的 `user_level` 字段判断是否升级，并显示升级动画。

---

## 收藏夹和浏览历史

### 收藏夹功能

用户可以创建多个收藏夹，每个收藏夹可以包含多个帖子。

#### 创建收藏夹
```http
POST /api/folders/create
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "name": "技术文章",
  "description": "收藏的技术相关文章",
  "is_public": true
}
```

**响应：**
```json
{
  "code": 200,
  "message": "创建收藏夹成功",
  "data": {
    "folder_id": 1
  }
}
```

#### 获取我的收藏夹列表
```http
GET /api/folders/my
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取收藏夹列表成功",
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "name": "技术文章",
      "description": "收藏的技术相关文章",
      "is_public": true,
      "item_count": 10,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 获取用户的收藏夹列表
```http
GET /api/folders/user/:id
Token: <your_token>
```

**说明：** 
- 查看自己的收藏夹：显示所有收藏夹（包括私密）
- 查看别人的收藏夹：只显示公开的收藏夹

#### 更新收藏夹
```http
PUT /api/folders/:id
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "name": "新名称",
  "description": "新描述",
  "is_public": false
}
```

#### 删除收藏夹
```http
DELETE /api/folders/:id
Token: <your_token>
```

**说明：** 删除收藏夹会同时删除其中的所有收藏项

#### 添加帖子到收藏夹
```http
POST /api/folders/:id/posts
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "post_id": 123
}
```

**响应：**
```json
{
  "code": 200,
  "message": "添加收藏成功"
}
```

#### 从收藏夹移除帖子
```http
DELETE /api/folders/:id/posts/:post_id
Token: <your_token>
```

#### 获取收藏夹中的帖子列表
```http
GET /api/folders/:id/posts?page=1&page_size=20
Token: <your_token>
```

**响应：**
```json
{
  "code": 200,
  "message": "获取收藏夹帖子成功",
  "data": {
    "folder": {
      "id": 1,
      "user_id": 1,
      "name": "技术文章",
      "description": "收藏的技术相关文章",
      "is_public": true,
      "item_count": 10
    },
    "posts": {
      "total": 10,
      "page": 1,
      "page_size": 20,
      "list": [...]
    }
  }
}
```

---

### 浏览历史

#### 获取浏览历史
```http
GET /api/history?page=1&page_size=20
Token: <your_token>
```

**说明：** 
- 浏览历史会在查看帖子详情时自动记录
- 按最后浏览时间排序
- 自动去重，只保留最后一次浏览记录

**响应：**
```json
{
  "code": 200,
  "message": "获取浏览历史成功",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "post": {...},
        "viewed_at": "2024-01-01T10:30:00Z"
      }
    ]
  }
}
```

---

## 点赞和投币功能

### 帖子点赞

#### 点赞帖子
```http
POST /api/posts/:id/like
Token: <your_token>
```

**说明：** 
- 每个用户对每个帖子只能点赞一次
- 点赞后会记录到 `post_likes` 表

**响应：**
```json
{
  "code": 200,
  "message": "点赞成功",
  "data": {
    "likes": 125
  }
}
```

#### 取消点赞帖子
```http
DELETE /api/posts/:id/like
Token: <your_token>
```

### 评论点赞和投币

#### 点赞评论
```http
POST /api/comments/:id/like
Token: <your_token>
```

**说明：** 
- 每个用户对每个评论只能点赞一次
- 点赞后会记录到 `comment_likes` 表

#### 投币评论
```http
POST /api/comments/:id/coin
Token: <your_token>
Content-Type: application/json
```

**请求体：**
```json
{
  "amount": 2
}
```

**说明：**
- 投币数量范围：1-10
- 会消耗用户的硬币
- 硬币不足时无法投币

**响应：**
```json
{
  "code": 200,
  "message": "投币成功",
  "data": {
    "coins": 15,
    "user_coins": 135
  }
}
```

---

## 用户详情API

### 获取用户详情
```http
GET /api/users/:id/detail
Token: <your_token>
```

**说明：** 
此API返回用户的完整信息，包括：
- 用户基本信息（含硬币、经验、等级）
- 关注数和粉丝数
- 发布的帖子数和收藏数
- 最近5个收藏夹
- 最近发布的10个帖子
- 最近收藏的10个帖子

**响应：**
```json
{
  "code": 200,
  "message": "获取用户详情成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "user@example.com",
      "level": 0,
      "user_level": 5,
      "exp": 1800,
      "coins": 200,
      "avatar": "http://...",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "coins": 200,
    "following_count": 10,
    "follower_count": 25,
    "post_count": 50,
    "favorite_count": 30,
    "folders": [
      {
        "id": 1,
        "name": "技术文章",
        "item_count": 15,
        ...
      }
    ],
    "posts": [
      {
        "id": 123,
        "title": "最新发布的帖子",
        ...
      }
    ],
    "favorites": [
      {
        "id": 456,
        "title": "收藏的帖子",
        ...
      }
    ]
  }
}
```

---

## 15. 应用市场 API

### 15.1 获取所有大分类
```http
GET /api/apps/categories
```

**说明：**
- 获取应用市场所有的大分类列表
- 不需要登录

**响应：**
```json
{
  "code": 200,
  "message": "获取大分类成功",
  "data": [
    "动作冒险",
    "休闲益智",
    "影音视听",
    "实用工具",
    "聊天社交",
    "图书阅读",
    "时尚购物",
    "摄影摄像",
    "学习教育",
    "旅行交通",
    "金融理财",
    "娱乐消遣",
    "新闻资讯",
    "居家生活",
    "体育运动",
    "医疗健康",
    "效率办公",
    "玩机",
    "定制系统应用"
  ]
}
```

### 15.2 获取指定大分类下的小分类
```http
GET /api/apps/subcategories?main_category=动作冒险
```

**查询参数：**
- `main_category` (必填): 大分类名称

**响应：**
```json
{
  "code": 200,
  "message": "获取小分类成功",
  "data": {
    "main_category": "动作冒险",
    "sub_categories": [
      "跑酷闯关",
      "网游RPG",
      "赛车体育",
      "飞行空战",
      "动作枪战",
      "格斗快打"
    ]
  }
}
```

### 15.3 根据分类获取应用列表
```http
GET /api/apps/category?main_category=动作冒险&sub_category=网游RPG
```

**查询参数：**
- `main_category` (必填): 大分类名称
- `sub_category` (必填): 小分类名称
- `sort` (可选): 排序方式
  - `rating`: 按评分排序
  - `download`: 按下载量排序（默认）
  - `update`: 按更新时间排序
- `page` (可选): 页码，默认1
- `page_size` (可选): 每页数量，默认20，最大100

**响应：**
```json
{
  "code": 200,
  "message": "获取分类应用列表成功",
  "data": {
    "total": 50,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "package_name": "com.example.rpg",
        "name": "示例RPG游戏",
        "icon_url": "https://example.com/icon.png",
        "version": "2.1.0",
        "size": 52428800,
        "rating": 4.8
      }
    ]
  }
}
```

### 15.4 获取应用列表
```http
GET /api/apps
```

**查询参数：**
- `category` (可选): 分类筛选（如：游戏、工具、社交等）
- `sort` (可选): 排序方式
  - `rating`: 按评分排序
  - `download`: 按下载量排序（默认）
  - `update`: 按更新时间排序
- `page` (可选): 页码，默认1
- `page_size` (可选): 每页数量，默认20，最大100

**响应：**
```json
{
  "code": 200,
  "message": "获取应用列表成功",
  "data": {
    "total": 150,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "package_name": "com.example.app",
        "name": "示例应用",
        "icon_url": "https://example.com/icon.png",
        "version": "1.2.3",
        "size": 10485760,  // 字节
        "rating": 4.5
      }
    ]
  }
}
```

### 15.5 获取应用详情
```http
GET /api/apps/:package_name
```

**路径参数：**
- `package_name`: 应用包名

**查询参数：**
- `version` (可选): 指定版本号，不传则返回最新版本

**响应：**
```json
{
  "code": 200,
  "message": "获取应用详情成功",
  "data": {
    "package_name": "com.example.app",
    "name": "示例应用",
    "icon_url": "https://example.com/icon.png",
    "version": "1.2.3",
    "version_code": 10203,
    "size": 10485760,
    "rating": 4.5,
    "rating_count": 1234,
    "description": "这是一个示例应用的详细介绍...",
    "screenshots": [
      "https://example.com/screenshot1.png",
      "https://example.com/screenshot2.png"
    ],
    "tags": ["工具", "效率", "免费"],
    "download_url": "https://example.com/app-v1.2.3.apk",
    "total_coins": 5678,
    "download_count": 12345,
    "uploader_name": "developer123",
    "update_content": "1. 修复了一些bug\n2. 优化了性能\n3. 新增了XX功能",
    "update_time": "2024-01-15 10:30:00",
    "main_category": "动作冒险",
    "sub_category": "网游RPG"
  }
}
```

### 15.6 给应用投币
```http
POST /api/apps/:package_name/coin
Token: <your_token>
Content-Type: application/json
```

**路径参数：**
- `package_name`: 应用包名

**请求体：**
```json
{
  "coins": 5  // 投币数量，1-10
}
```

**响应：**
```json
{
  "code": 200,
  "message": "投币成功，投了5个硬币",
  "data": {
    "total_coins": 5683  // 应用当前总投币数
  }
}
```

### 15.7 记录应用下载
```http
POST /api/apps/:package_name/download
```

**路径参数：**
- `package_name`: 应用包名

**说明：**
- 此API用于记录应用下载次数
- 不需要登录即可调用
- 每次调用会将应用的下载计数+1

**响应：**
```json
{
  "code": 200,
  "message": "下载记录成功"
}
```

---

## 应用市场功能说明

### 数据结构

1. **应用表 (apps)**
   - 存储应用的基本信息
   - 包含包名、名称、图标、描述、标签、评分、投币数、下载量等

2. **应用版本表 (app_versions)**
   - 存储应用的各个版本信息
   - 包含版本号、版本代码、大小、下载链接、更新内容、截图、上传者等
   - 支持多版本管理，标记最新版本

### 功能特点

1. **应用列表**
   - 支持分类筛选
   - 多种排序方式（评分、下载量、更新时间）
   - 分页展示
   - 显示最新版本信息

2. **应用详情**
   - 完整的应用信息展示
   - 支持查看指定版本或最新版本
   - 包含应用截图、标签、更新内容等
   - 显示投币数和下载量

3. **投币功能**
   - 用户可以给喜欢的应用投币支持
   - 投币数量限制1-10个
   - 需要登录且硬币充足

4. **下载统计**
   - 自动记录应用下载次数
   - 不需要登录即可统计

