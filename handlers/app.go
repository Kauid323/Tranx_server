package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"TaruApp/database"
	"TaruApp/models"

	"github.com/gin-gonic/gin"
)

// 定义应用分类数据
var appCategories = map[string][]string{
	"动作冒险":   {"跑酷闯关", "网游RPG", "赛车体育", "飞行空战", "动作枪战", "格斗快打"},
	"休闲益智":   {"休闲创意", "棋牌桌游", "模拟经营", "战争策略", "塔防迷宫", "儿童益智"},
	"影音视听":   {"视频", "音乐", "直播", "电台", "播放器"},
	"实用工具":   {"系统", "安全", "浏览器", "输入法", "小工具"},
	"聊天社交":   {"聊天", "婚恋", "通讯", "交友", "社区"},
	"图书阅读":   {"听书", "漫画", "电子书", "小说", "杂志"},
	"时尚购物":   {"电商", "团购", "海淘", "导购", "时尚"},
	"摄影摄像":   {"美图", "相机", "图片分享", "相册", "视频"},
	"学习教育":   {"外语", "考试", "教育", "育儿", "驾考"},
	"旅行交通":   {"用车", "地图", "旅游", "酒店", "票务", "公交地铁"},
	"金融理财":   {"银行", "股票", "基金", "记账", "支付", "贷款"},
	"娱乐消遣":   {"搞怪", "消遣", "星座运势", "笑话"},
	"新闻资讯":   {"新闻", "资讯", "科技", "热点", "头条"},
	"居家生活":   {"闹钟", "查违章", "天气日历", "美食", "电影票", "房产家居"},
	"体育运动":   {"健身", "计步", "球类", "直播"},
	"医疗健康":   {"减肥", "经期", "养生", "孕育", "美容", "医疗"},
	"效率办公":   {"办公", "邮箱", "笔记", "云盘", "日程"},
	"玩机":     {"系统", "调度", "美化", "其他"},
	"定制系统应用": {"OPPO", "真我", "华为", "荣耀", "小米", "VIVO", "三星", "一加", "金立", "LG", "海信", "夏普", "摩托罗拉", "谷歌Google", "iQOO", "红魔", "魅族", "TCL", "百度", "小辣椒", "Fairphone", "Nothing", "努比亚", "索尼", "诺基亚", "黑鲨", "联想", "威图Vertu", "华硕", "酷派", "飞利浦", "乐视", "朵唯", "FreemeOS", "HTC", "柔宇", "黑莓BlackBerry", "AGM", "8848", "鼎桥", "ROG", "中兴", "其他"},
}

// GetApps 获取应用列表
func GetApps(c *gin.Context) {
	var query models.GetAppsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建查询SQL
	baseQuery := `
		SELECT DISTINCT 
			a.package_name, a.name, a.icon_url, a.rating,
			v.version, v.size
		FROM apps a
		INNER JOIN app_versions v ON a.id = v.app_id AND v.is_latest = 1
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(DISTINCT a.id) FROM apps a WHERE 1=1`
	args := []interface{}{}
	countArgs := []interface{}{}

	// 分类筛选
	if query.Category != "" {
		baseQuery += " AND a.tags LIKE ?"
		countQuery += " AND a.tags LIKE ?"
		categoryArg := "%" + query.Category + "%"
		args = append(args, categoryArg)
		countArgs = append(countArgs, categoryArg)
	}

	// 排序
	switch query.Sort {
	case "rating":
		baseQuery += " ORDER BY a.rating DESC, a.download_count DESC"
	case "download":
		baseQuery += " ORDER BY a.download_count DESC"
	case "update":
		baseQuery += " ORDER BY v.created_at DESC"
	default:
		baseQuery += " ORDER BY a.download_count DESC"
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, query.PageSize, offset)

	// 查询总数
	var total int
	if err := database.DB.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询应用总数失败: " + err.Error(),
		})
		return
	}

	// 查询列表
	rows, err := database.DB.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询应用列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	apps := []models.AppListItem{}
	for rows.Next() {
		var app models.AppListItem
		err := rows.Scan(
			&app.PackageName, &app.Name, &app.IconURL, &app.Rating,
			&app.Version, &app.Size,
		)
		if err != nil {
			continue
		}
		apps = append(apps, app)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取应用列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
			List:     apps,
		},
	})
}

// GetAppDetail 获取应用详情
func GetAppDetail(c *gin.Context) {
	packageName := c.Param("package_name")
	if packageName == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "应用包名不能为空",
		})
		return
	}

	var query models.GetAppDetailQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 首先获取应用基本信息
	var app models.App
	var mainCategory, subCategory sql.NullString
	err := database.DB.QueryRow(
		`SELECT id, package_name, name, icon_url, description, tags, main_category, sub_category,
			rating, rating_count, total_coins, download_count 
		FROM apps WHERE package_name = ?`,
		packageName,
	).Scan(
		&app.ID, &app.PackageName, &app.Name, &app.IconURL, &app.Description,
		&app.Tags, &mainCategory, &subCategory, &app.Rating, &app.RatingCount,
		&app.TotalCoins, &app.DownloadCount,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "应用不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询应用信息失败: " + err.Error(),
		})
		return
	}

	// 查询版本信息
	var versionQuery string
	var versionArgs []interface{}

	if query.Version != "" {
		// 查询指定版本
		versionQuery = `
			SELECT version, version_code, size, download_url, update_content, 
				screenshots, uploader_name, created_at
			FROM app_versions 
			WHERE app_id = ? AND version = ?
		`
		versionArgs = []interface{}{app.ID, query.Version}
	} else {
		// 查询最新版本
		versionQuery = `
			SELECT version, version_code, size, download_url, update_content, 
				screenshots, uploader_name, created_at
			FROM app_versions 
			WHERE app_id = ? AND is_latest = 1
			ORDER BY version_code DESC
			LIMIT 1
		`
		versionArgs = []interface{}{app.ID}
	}

	var version models.AppVersion
	var screenshotsJSON string
	err = database.DB.QueryRow(versionQuery, versionArgs...).Scan(
		&version.Version, &version.VersionCode, &version.Size,
		&version.DownloadURL, &version.UpdateContent, &screenshotsJSON,
		&version.UploaderName, &version.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "应用版本不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询版本信息失败: " + err.Error(),
		})
		return
	}

	// 解析截图JSON
	var screenshots []string
	if screenshotsJSON != "" {
		if err := json.Unmarshal([]byte(screenshotsJSON), &screenshots); err != nil {
			screenshots = []string{}
		}
	}

	// 解析标签
	tags := []string{}
	if app.Tags != "" {
		tags = strings.Split(app.Tags, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// 构建响应
	detail := models.AppDetail{
		PackageName:   app.PackageName,
		Name:          app.Name,
		IconURL:       app.IconURL,
		Version:       version.Version,
		VersionCode:   version.VersionCode,
		Size:          version.Size,
		Rating:        app.Rating,
		RatingCount:   app.RatingCount,
		Description:   app.Description,
		Screenshots:   screenshots,
		Tags:          tags,
		DownloadURL:   version.DownloadURL,
		TotalCoins:    app.TotalCoins,
		DownloadCount: app.DownloadCount,
		UploaderName:  version.UploaderName,
		UpdateContent: version.UpdateContent,
		UpdateTime:    version.CreatedAt.Format("2006-01-02 15:04:05"),
		MainCategory:  mainCategory.String,
		SubCategory:   subCategory.String,
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取应用详情成功",
		Data:    detail,
	})
}

// CreateApp 创建应用（管理员功能，可选实现）
func CreateApp(c *gin.Context) {
	// TODO: 实现应用创建逻辑
	c.JSON(http.StatusNotImplemented, models.Response{
		Code:    501,
		Message: "功能未实现",
	})
}

// UpdateApp 更新应用信息（管理员功能，可选实现）
func UpdateApp(c *gin.Context) {
	// TODO: 实现应用更新逻辑
	c.JSON(http.StatusNotImplemented, models.Response{
		Code:    501,
		Message: "功能未实现",
	})
}

// UploadAppVersion 上传应用版本
func UploadAppVersion(c *gin.Context) {
	// TODO: 实现版本上传逻辑
	c.JSON(http.StatusNotImplemented, models.Response{
		Code:    501,
		Message: "功能未实现",
	})
}

// CoinApp 给应用投币
func CoinApp(c *gin.Context) {
	packageName := c.Param("package_name")
	userID, _ := c.Get("user_id")

	var req struct {
		Coins int `json:"coins" binding:"required,min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查应用是否存在
	var appID int64
	err := database.DB.QueryRow(
		"SELECT id FROM apps WHERE package_name = ?",
		packageName,
	).Scan(&appID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "应用不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询应用失败: " + err.Error(),
		})
		return
	}

	// 检查用户硬币是否足够
	var userCoins int
	err = database.DB.QueryRow(
		"SELECT coins FROM users WHERE id = ?",
		userID,
	).Scan(&userCoins)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询用户信息失败: " + err.Error(),
		})
		return
	}

	if userCoins < req.Coins {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: fmt.Sprintf("硬币不足，当前硬币: %d", userCoins),
		})
		return
	}

	// 开始事务
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	// 扣除用户硬币
	_, err = tx.Exec(
		"UPDATE users SET coins = coins - ? WHERE id = ?",
		req.Coins, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}

	// 增加应用投币数
	_, err = tx.Exec(
		"UPDATE apps SET total_coins = total_coins + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		req.Coins, appID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "投币失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的投币数
	var totalCoins int
	database.DB.QueryRow("SELECT total_coins FROM apps WHERE id = ?", appID).Scan(&totalCoins)

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: fmt.Sprintf("投币成功，投了%d个硬币", req.Coins),
		Data: gin.H{
			"total_coins": totalCoins,
		},
	})
}

// DownloadApp 记录应用下载
func DownloadApp(c *gin.Context) {
	packageName := c.Param("package_name")

	// 增加下载计数
	_, err := database.DB.Exec(
		"UPDATE apps SET download_count = download_count + 1 WHERE package_name = ?",
		packageName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "记录下载失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "下载记录成功",
	})
}

// GetMainCategories 获取所有大分类
func GetMainCategories(c *gin.Context) {
	mainCategories := make([]string, 0, len(appCategories))
	for category := range appCategories {
		mainCategories = append(mainCategories, category)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取大分类成功",
		Data:    mainCategories,
	})
}

// GetSubCategories 获取指定大分类下的小分类
func GetSubCategories(c *gin.Context) {
	mainCategory := c.Query("main_category")
	if mainCategory == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "大分类参数不能为空",
		})
		return
	}

	subCategories, exists := appCategories[mainCategory]
	if !exists {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "大分类不存在",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取小分类成功",
		Data: models.AppCategory{
			MainCategory:  mainCategory,
			SubCategories: subCategories,
		},
	})
}

// GetAppsByCategory 根据分类获取应用列表
func GetAppsByCategory(c *gin.Context) {
	var query models.GetAppsByCategoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证分类是否存在
	subCategories, exists := appCategories[query.MainCategory]
	if !exists {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "大分类不存在",
		})
		return
	}

	// 验证小分类是否存在
	subCategoryExists := false
	for _, sub := range subCategories {
		if sub == query.SubCategory {
			subCategoryExists = true
			break
		}
	}
	if !subCategoryExists {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "小分类不存在",
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 构建查询SQL
	baseQuery := `
		SELECT DISTINCT 
			a.package_name, a.name, a.icon_url, a.rating,
			v.version, v.size
		FROM apps a
		INNER JOIN app_versions v ON a.id = v.app_id AND v.is_latest = 1
		WHERE a.main_category = ? AND a.sub_category = ?
	`
	countQuery := `SELECT COUNT(DISTINCT a.id) FROM apps a WHERE a.main_category = ? AND a.sub_category = ?`
	args := []interface{}{query.MainCategory, query.SubCategory}
	countArgs := []interface{}{query.MainCategory, query.SubCategory}

	// 排序
	switch query.Sort {
	case "rating":
		baseQuery += " ORDER BY a.rating DESC, a.download_count DESC"
	case "download":
		baseQuery += " ORDER BY a.download_count DESC"
	case "update":
		baseQuery += " ORDER BY v.created_at DESC"
	default:
		baseQuery += " ORDER BY a.download_count DESC"
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, query.PageSize, offset)

	// 查询总数
	var total int
	if err := database.DB.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询应用总数失败: " + err.Error(),
		})
		return
	}

	// 查询列表
	rows, err := database.DB.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "查询应用列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	apps := []models.AppListItem{}
	for rows.Next() {
		var app models.AppListItem
		err := rows.Scan(
			&app.PackageName, &app.Name, &app.IconURL, &app.Rating,
			&app.Version, &app.Size,
		)
		if err != nil {
			continue
		}
		apps = append(apps, app)
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    200,
		Message: "获取分类应用列表成功",
		Data: models.PageData{
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
			List:     apps,
		},
	})
}
