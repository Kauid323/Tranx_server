package models

import "time"

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`          // 密码不返回给前端
	Email     string    `json:"email"`      // 邮箱（预留）
	Level     int       `json:"level"`      // 用户等级: 0-普通用户, 50-管理员
	Avatar    string    `json:"avatar"`     // 头像URL
	Coins     int       `json:"coins"`      // 硬币数量
	Exp       int       `json:"exp"`        // 经验值
	UserLevel int       `json:"user_level"` // 用户等级 (Lv1, Lv2...)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Follow 关注关系模型
type Follow struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`     // 关注者ID
	FollowedID int64     `json:"followed_id"` // 被关注者ID
	CreatedAt  time.Time `json:"created_at"`
}

// CheckIn 签到记录模型
type CheckIn struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	CheckDate string    `json:"check_date"` // 签到日期 YYYY-MM-DD
	CheckTime time.Time `json:"check_time"` // 签到时间
	Reward    int       `json:"reward"`     // 奖励硬币数
	CreatedAt time.Time `json:"created_at"`
}

// UserTag 用户标签模型
type UserTag struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TagName   string    `json:"tag_name"`
	TagColor  string    `json:"tag_color"` // 标签颜色
	CreatedAt time.Time `json:"created_at"`
}

// Token 认证令牌模型
type Token struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`      // RC4加密的token
	ExpiresAt time.Time `json:"expires_at"` // 过期时间
	CreatedAt time.Time `json:"created_at"`
}

// Board 板块模型
type Board struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Post 帖子模型
type Post struct {
	ID             int64     `json:"id"`
	BoardID        int64     `json:"board_id"`
	UserID         int64     `json:"user_id"` // 发布者用户ID
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Type           string    `json:"type"`            // 帖子类型: "text"(普通文本) 或 "markdown"(Markdown格式)
	Publisher      string    `json:"publisher"`       // 发布者名称
	PublishTime    time.Time `json:"publish_time"`    // 发布时间
	Coins          int       `json:"coins"`           // 投币数
	Favorites      int       `json:"favorites"`       // 收藏数
	Likes          int       `json:"likes"`           // 点赞数
	ImageURL       string    `json:"image_url"`       // 图片URL
	AttachmentURL  string    `json:"attachment_url"`  // 附件URL (预留用于APK等文件上传)
	AttachmentType string    `json:"attachment_type"` // 附件类型 (apk, zip等)
	CommentCount   int       `json:"comment_count"`   // 评论数
	ViewCount      int       `json:"view_count"`      // 浏览数
	LastReplyTime  time.Time `json:"last_reply_time"` // 最后回复时间
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Comment 评论模型
type Comment struct {
	ID          int64     `json:"id"`
	PostID      int64     `json:"post_id"`
	UserID      int64     `json:"user_id"`   // 评论者用户ID
	ParentID    *int64    `json:"parent_id"` // 父评论ID，用于楼中楼回复
	Content     string    `json:"content"`
	Publisher   string    `json:"publisher"`    // 评论者用户名
	Avatar      string    `json:"avatar"`       // 评论者头像URL
	PublishTime time.Time `json:"publish_time"` // 评论时间
	Likes       int       `json:"likes"`        // 点赞数
	Coins       int       `json:"coins"`        // 投币数
	IsAuthor    bool      `json:"is_author"`    // 是否为楼主
	Floor       int       `json:"floor"`        // 楼层号
	ReplyCount  int       `json:"reply_count"`  // 子回复数量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"` // 预留邮箱字段
	Avatar   string `json:"avatar"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SetUserLevelRequest 设置用户等级请求
type SetUserLevelRequest struct {
	Level int `json:"level" binding:"required"`
}

// CreateUserTagRequest 创建用户标签请求
type CreateUserTagRequest struct {
	UserID   int64  `json:"user_id" binding:"required"`
	TagName  string `json:"tag_name" binding:"required"`
	TagColor string `json:"tag_color"`
}

// UserStats 用户统计信息
type UserStats struct {
	FollowingCount int  `json:"following_count"` // 关注数
	FollowerCount  int  `json:"follower_count"`  // 粉丝数
	IsFollowing    bool `json:"is_following"`    // 当前用户是否关注了该用户
}

// CheckInRankItem 签到排行榜项
type CheckInRankItem struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Avatar    string    `json:"avatar"`
	CheckTime time.Time `json:"check_time"`
	Rank      int       `json:"rank"`
}

// FavoriteFolder 收藏夹模型
type FavoriteFolder struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`  // 是否公开
	ItemCount   int       `json:"item_count"` // 收藏数量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FavoriteItem 收藏夹项目
type FavoriteItem struct {
	ID        int64     `json:"id"`
	FolderID  int64     `json:"folder_id"`
	PostID    int64     `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateFavoriteFolderRequest 创建收藏夹请求
type CreateFavoriteFolderRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateFavoriteFolderRequest 更新收藏夹请求
type UpdateFavoriteFolderRequest struct {
	Name        string `json:"name" binding:"min=1,max=50"`
	Description string `json:"description" binding:"max=200"`
	IsPublic    *bool  `json:"is_public"`
}

// AddToFolderRequest 添加到收藏夹请求
type AddToFolderRequest struct {
	PostID int64 `json:"post_id" binding:"required"`
}

// PostLike 帖子点赞记录
type PostLike struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	PostID    int64     `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CommentLike 评论点赞记录
type CommentLike struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	CommentID int64     `json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ViewHistory 浏览历史
type ViewHistory struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	PostID   int64     `json:"post_id"`
	ViewedAt time.Time `json:"viewed_at"`
}

// UserDetail 用户详情（扩展信息）
type UserDetail struct {
	User           User   `json:"user"`
	Coins          int    `json:"coins"`
	FollowingCount int    `json:"following_count"`
	FollowerCount  int    `json:"follower_count"`
	PostCount      int    `json:"post_count"`
	FavoriteCount  int    `json:"favorite_count"`
	Posts          []Post `json:"posts"`
	Favorites      []Post `json:"favorites"`
}

// CreateBoardRequest 创建板块请求
type CreateBoardRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreatePostRequest 创建帖子请求
type CreatePostRequest struct {
	BoardID  int64  `json:"board_id"` // 板块ID，不传或传0则默认发到主板块(ID=1)
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Type     string `json:"type"` // 帖子类型: "text"(默认) 或 "markdown"
	ImageURL string `json:"image_url"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	PostID   int64  `json:"post_id" binding:"required"`
	ParentID *int64 `json:"parent_id"` // 可选，用于楼中楼回复
	Content  string `json:"content" binding:"required"`
}

// GetPostsQuery 获取帖子列表查询参数
type GetPostsQuery struct {
	BoardID  int64  `form:"board_id"`                                 // 板块ID
	Sort     string `form:"sort" binding:"oneof=latest reply hot ''"` // 排序方式: latest(最新发布), reply(最近回复), hot(热门)
	Page     int    `form:"page"`                                     // 页码
	PageSize int    `form:"page_size"`                                // 每页数量
}

// GetCommentsQuery 获取评论列表查询参数
type GetCommentsQuery struct {
	PostID   int64  `form:"post_id" binding:"required"`                        // 帖子ID
	Sort     string `form:"sort" binding:"oneof=default likes author desc ''"` // 排序: default(默认), likes(点赞最高), author(楼主发布), desc(倒序)
	Page     int    `form:"page"`                                              // 页码
	PageSize int    `form:"page_size"`                                         // 每页数量
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据
type PageData struct {
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     interface{} `json:"list"`
}

// App 应用信息
type App struct {
	ID            int64     `json:"id"`
	PackageName   string    `json:"package_name"`   // 应用包名（唯一标识）
	Name          string    `json:"name"`           // 应用名称
	IconURL       string    `json:"icon_url"`       // 应用图标URL
	Description   string    `json:"description"`    // 应用介绍
	Tags          string    `json:"tags"`           // 应用标签（逗号分隔）
	Rating        float64   `json:"rating"`         // 应用评分（0-5）
	RatingCount   int       `json:"rating_count"`   // 评分人数
	TotalCoins    int       `json:"total_coins"`    // 总投币数
	DownloadCount int       `json:"download_count"` // 总下载量
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AppVersion 应用版本
type AppVersion struct {
	ID            int64     `json:"id"`
	AppID         int64     `json:"app_id"`
	PackageName   string    `json:"package_name"`   // 冗余字段，方便查询
	Version       string    `json:"version"`        // 版本号
	VersionCode   int       `json:"version_code"`   // 版本代码（用于排序）
	Size          int64     `json:"size"`           // 应用大小（字节）
	DownloadURL   string    `json:"download_url"`   // 下载链接
	UpdateContent string    `json:"update_content"` // 更新内容
	Screenshots   string    `json:"screenshots"`    // 预览图URLs（JSON数组）
	UploaderID    int64     `json:"uploader_id"`    // 上传者ID
	UploaderName  string    `json:"uploader_name"`  // 上传者用户名
	IsLatest      bool      `json:"is_latest"`      // 是否最新版本
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AppListItem 应用列表项
type AppListItem struct {
	PackageName string  `json:"package_name"`
	Name        string  `json:"name"`
	IconURL     string  `json:"icon_url"`
	Version     string  `json:"version"` // 最新版本号
	Size        int64   `json:"size"`    // 最新版本大小
	Rating      float64 `json:"rating"`
}

// AppDetail 应用详情
type AppDetail struct {
	PackageName   string   `json:"package_name"`
	Name          string   `json:"name"`
	IconURL       string   `json:"icon_url"`
	Version       string   `json:"version"`
	VersionCode   int      `json:"version_code"`
	Size          int64    `json:"size"`
	Rating        float64  `json:"rating"`
	RatingCount   int      `json:"rating_count"`
	Description   string   `json:"description"`
	Screenshots   []string `json:"screenshots"`
	Tags          []string `json:"tags"`
	DownloadURL   string   `json:"download_url"`
	TotalCoins    int      `json:"total_coins"`
	DownloadCount int      `json:"download_count"`
	UploaderName  string   `json:"uploader_name"`
	UpdateContent string   `json:"update_content"`
	UpdateTime    string   `json:"update_time"`
	MainCategory  string   `json:"main_category"` // 大分类
	SubCategory   string   `json:"sub_category"`  // 小分类
}

// GetAppsQuery 获取应用列表查询参数
type GetAppsQuery struct {
	Category string `form:"category"`  // 分类筛选
	Sort     string `form:"sort"`      // 排序: rating, download, update
	Page     int    `form:"page"`      // 页码
	PageSize int    `form:"page_size"` // 每页数量
}

// GetAppDetailQuery 获取应用详情查询参数
type GetAppDetailQuery struct {
	Version string `form:"version"` // 版本号（可选，不传则返回最新版本）
}

// AppCategory 应用分类
type AppCategory struct {
	MainCategory  string   `json:"main_category"`  // 大分类
	SubCategories []string `json:"sub_categories"` // 小分类列表
}

// GetAppsByCategoryQuery 根据分类获取应用列表查询参数
type GetAppsByCategoryQuery struct {
	MainCategory string `form:"main_category" binding:"required"` // 大分类
	SubCategory  string `form:"sub_category" binding:"required"`  // 小分类
	Sort         string `form:"sort"`                             // 排序方式
	Page         int    `form:"page"`                             // 页码
	PageSize     int    `form:"page_size"`                        // 每页数量
}
