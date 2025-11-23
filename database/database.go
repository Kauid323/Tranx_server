package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB 初始化数据库
func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite", "./taruapp.db")
	if err != nil {
		return err
	}

	// 测试连接
	if err = DB.Ping(); err != nil {
		return err
	}

	// 创建表
	if err = createTables(); err != nil {
		return err
	}

	log.Println("数据库初始化成功")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// TableDefinition 表定义结构
type TableDefinition struct {
	Name   string
	SQL    string
	Repair func() error // 可选的修复函数
}

// createTables 创建数据表
func createTables() error {
	log.Println("开始检查和创建数据表...")

	// 定义所有表的结构
	tables := []TableDefinition{
		{
			Name: "users",
			SQL: `CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT NOT NULL UNIQUE,
				password TEXT NOT NULL,
				email TEXT,
				level INTEGER DEFAULT 0,
				avatar TEXT,
				coins INTEGER DEFAULT 0,
				exp INTEGER DEFAULT 0,
				user_level INTEGER DEFAULT 1,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
			Repair: repairUsersTable,
		},
		{
			Name: "follows",
			SQL: `CREATE TABLE IF NOT EXISTS follows (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				followed_id INTEGER NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, followed_id),
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (followed_id) REFERENCES users(id)
			);`,
		},
		{
			Name: "check_ins",
			SQL: `CREATE TABLE IF NOT EXISTS check_ins (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				check_date TEXT NOT NULL,
				check_time DATETIME NOT NULL,
				reward INTEGER DEFAULT 50,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, check_date),
				FOREIGN KEY (user_id) REFERENCES users(id)
			);`,
		},
		{
			Name: "tokens",
			SQL: `CREATE TABLE IF NOT EXISTS tokens (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				token TEXT NOT NULL UNIQUE,
				expires_at DATETIME NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id)
			);`,
		},
		{
			Name: "user_tags",
			SQL: `CREATE TABLE IF NOT EXISTS user_tags (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				tag_name TEXT NOT NULL,
				tag_color TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id)
			);`,
		},
		{
			Name: "boards",
			SQL: `CREATE TABLE IF NOT EXISTS boards (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				description TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
			Repair: repairBoardsTable,
		},
		{
			Name: "posts",
			SQL: `CREATE TABLE IF NOT EXISTS posts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				board_id INTEGER NOT NULL,
				user_id INTEGER NOT NULL,
				title TEXT NOT NULL,
				content TEXT NOT NULL,
				type TEXT DEFAULT 'text',
				publisher TEXT NOT NULL,
				publish_time DATETIME DEFAULT CURRENT_TIMESTAMP,
				coins INTEGER DEFAULT 0,
				favorites INTEGER DEFAULT 0,
				likes INTEGER DEFAULT 0,
				image_url TEXT,
				attachment_url TEXT,
				attachment_type TEXT,
				comment_count INTEGER DEFAULT 0,
				view_count INTEGER DEFAULT 0,
				last_reply_time DATETIME DEFAULT CURRENT_TIMESTAMP,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (board_id) REFERENCES boards(id),
				FOREIGN KEY (user_id) REFERENCES users(id)
			);`,
			Repair: repairPostsTable,
		},
		{
			Name: "comments",
			SQL: `CREATE TABLE IF NOT EXISTS comments (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				post_id INTEGER NOT NULL,
				user_id INTEGER NOT NULL,
				parent_id INTEGER,
				content TEXT NOT NULL,
				publisher TEXT NOT NULL,
				publish_time DATETIME DEFAULT CURRENT_TIMESTAMP,
				likes INTEGER DEFAULT 0,
				coins INTEGER DEFAULT 0,
				is_author BOOLEAN DEFAULT 0,
				floor INTEGER NOT NULL,
				reply_count INTEGER DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (post_id) REFERENCES posts(id),
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (parent_id) REFERENCES comments(id)
			);`,
			Repair: repairCommentsTable,
		},
		{
			Name: "favorite_folders",
			SQL: `CREATE TABLE IF NOT EXISTS favorite_folders (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				is_public BOOLEAN DEFAULT 0,
				item_count INTEGER DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id)
			);`,
		},
		{
			Name: "favorite_items",
			SQL: `CREATE TABLE IF NOT EXISTS favorite_items (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				folder_id INTEGER NOT NULL,
				post_id INTEGER NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(folder_id, post_id),
				FOREIGN KEY (folder_id) REFERENCES favorite_folders(id),
				FOREIGN KEY (post_id) REFERENCES posts(id)
			);`,
		},
		{
			Name: "post_likes",
			SQL: `CREATE TABLE IF NOT EXISTS post_likes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				post_id INTEGER NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, post_id),
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (post_id) REFERENCES posts(id)
			);`,
		},
		{
			Name: "comment_likes",
			SQL: `CREATE TABLE IF NOT EXISTS comment_likes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				comment_id INTEGER NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id, comment_id),
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (comment_id) REFERENCES comments(id)
			);`,
		},
		{
			Name: "view_histories",
			SQL: `CREATE TABLE IF NOT EXISTS view_histories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				post_id INTEGER NOT NULL,
				viewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (post_id) REFERENCES posts(id)
			);`,
		},
		{
			Name: "apps",
			SQL: `CREATE TABLE IF NOT EXISTS apps (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				package_name TEXT UNIQUE NOT NULL,
				name TEXT NOT NULL,
				icon_url TEXT,
				description TEXT,
				tags TEXT,
				main_category TEXT,
				sub_category TEXT,
				channel TEXT,
				share_desc TEXT,
				developer_name TEXT,
				ad_level TEXT,
				payment_type TEXT,
				operation_type TEXT,
				rating REAL DEFAULT 0,
				rating_count INTEGER DEFAULT 0,
				total_coins INTEGER DEFAULT 0,
				download_count INTEGER DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
			Repair: repairAppsTable,
		},
		{
			Name: "app_versions",
			SQL: `CREATE TABLE IF NOT EXISTS app_versions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				app_id INTEGER NOT NULL,
				package_name TEXT NOT NULL,
				version TEXT NOT NULL,
				version_code INTEGER NOT NULL,
				size INTEGER NOT NULL,
				download_url TEXT NOT NULL,
				update_content TEXT,
				screenshots TEXT,
				uploader_id INTEGER NOT NULL,
				uploader_name TEXT NOT NULL,
				is_latest BOOLEAN DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(package_name, version),
				FOREIGN KEY (app_id) REFERENCES apps(id),
				FOREIGN KEY (uploader_id) REFERENCES users(id)
			);`,
		},
		{
			Name: "app_upload_tasks",
			SQL: `CREATE TABLE IF NOT EXISTS app_upload_tasks (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				package_name TEXT NOT NULL,
				name TEXT NOT NULL,
				icon_url TEXT NOT NULL,
				version TEXT NOT NULL,
				version_code INTEGER NOT NULL,
				size INTEGER NOT NULL,
				channel TEXT NOT NULL,
				main_category TEXT NOT NULL,
				sub_category TEXT NOT NULL,
				screenshots TEXT,
				description TEXT,
				share_desc TEXT,
				update_content TEXT,
				developer_name TEXT NOT NULL,
				ad_level TEXT NOT NULL,
				payment_type TEXT NOT NULL,
				operation_type TEXT NOT NULL,
				download_url TEXT NOT NULL,
				status TEXT DEFAULT 'pending',
				reject_reason TEXT,
				reviewer_id INTEGER,
				review_time DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (reviewer_id) REFERENCES users(id)
			);`,
		},
	}

	// 检查并创建每个表
	for _, table := range tables {
		exists, err := tableExists(table.Name)
		if err != nil {
			log.Printf("检查表 %s 是否存在时出错: %v", table.Name, err)
			continue
		}

		if !exists {
			log.Printf("表 %s 不存在，正在创建...", table.Name)
			if err := createTable(table); err != nil {
				log.Printf("创建表 %s 失败: %v", table.Name, err)
				return err
			}
			log.Printf("✓ 表 %s 创建成功", table.Name)
		} else {
			log.Printf("✓ 表 %s 已存在", table.Name)
			// 如果表存在但可能缺少字段，执行修复
			if table.Repair != nil {
				if err := table.Repair(); err != nil {
					log.Printf("修复表 %s 时出错: %v", table.Name, err)
				}
			}
		}
	}

	// 创建索引
	if err := createIndexes(); err != nil {
		return err
	}

	// 确保默认数据存在
	if err := ensureDefaultData(); err != nil {
		return err
	}

	log.Println("数据表检查和创建完成")
	return nil
}

// tableExists 检查表是否存在
func tableExists(tableName string) (bool, error) {
	var count int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
		tableName,
	).Scan(&count)
	return count > 0, err
}

// createTable 创建单个表
func createTable(table TableDefinition) error {
	_, err := DB.Exec(table.SQL)
	return err
}

// createIndexes 创建索引
func createIndexes() error {
	log.Println("开始创建索引...")
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_token ON tokens(token);`,
		`CREATE INDEX IF NOT EXISTS idx_user_tags_user_id ON user_tags(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_follows_user_id ON follows(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_follows_followed_id ON follows(followed_id);`,
		`CREATE INDEX IF NOT EXISTS idx_check_ins_user_id ON check_ins(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_check_ins_check_date ON check_ins(check_date);`,
		`CREATE INDEX IF NOT EXISTS idx_check_ins_check_time ON check_ins(check_time);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_board_id ON posts(board_id);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_publish_time ON posts(publish_time DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_last_reply_time ON posts(last_reply_time DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_likes ON posts(likes DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_likes ON comments(likes DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_floor ON comments(floor);`,
		`CREATE INDEX IF NOT EXISTS idx_favorite_folders_user_id ON favorite_folders(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_favorite_items_folder_id ON favorite_items(folder_id);`,
		`CREATE INDEX IF NOT EXISTS idx_favorite_items_post_id ON favorite_items(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_likes_user_id ON post_likes(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_likes_post_id ON post_likes(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comment_likes_user_id ON comment_likes(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comment_likes_comment_id ON comment_likes(comment_id);`,
		`CREATE INDEX IF NOT EXISTS idx_view_histories_user_id ON view_histories(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_view_histories_post_id ON view_histories(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_view_histories_viewed_at ON view_histories(viewed_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_apps_package_name ON apps(package_name);`,
		`CREATE INDEX IF NOT EXISTS idx_apps_rating ON apps(rating DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_apps_download_count ON apps(download_count DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_app_versions_app_id ON app_versions(app_id);`,
		`CREATE INDEX IF NOT EXISTS idx_app_versions_package_name ON app_versions(package_name);`,
		`CREATE INDEX IF NOT EXISTS idx_app_versions_version_code ON app_versions(version_code DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_app_versions_is_latest ON app_versions(is_latest);`,
		`CREATE INDEX IF NOT EXISTS idx_app_upload_tasks_user_id ON app_upload_tasks(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_app_upload_tasks_status ON app_upload_tasks(status);`,
		`CREATE INDEX IF NOT EXISTS idx_app_upload_tasks_package_name ON app_upload_tasks(package_name);`,
		`CREATE INDEX IF NOT EXISTS idx_app_upload_tasks_created_at ON app_upload_tasks(created_at DESC);`,
	}

	for _, index := range indexes {
		if _, err := DB.Exec(index); err != nil {
			log.Printf("创建索引失败: %v", err)
		}
	}
	log.Println("✓ 索引创建完成")
	return nil
}

// ensureDefaultData 确保默认数据存在
func ensureDefaultData() error {
	log.Println("检查默认数据...")

	// 创建默认主板块（如果不存在）
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM boards WHERE id = 1").Scan(&count)
	if err == nil && count == 0 {
		_, err = DB.Exec(
			"INSERT INTO boards (id, name, description) VALUES (1, '综合讨论', '默认主板块，所有话题都可以在这里讨论')",
		)
		if err != nil {
			log.Printf("创建默认主板块失败: %v", err)
		} else {
			log.Println("✓ 创建默认主板块成功")
		}
	}

	return nil
}

// 表修复函数
func repairUsersTable() error {
	// 检查并添加可能缺失的字段
	columns := []struct {
		name         string
		definition   string
		defaultValue string
	}{
		{"coins", "INTEGER DEFAULT 0", "0"},
		{"exp", "INTEGER DEFAULT 0", "0"},
		{"user_level", "INTEGER DEFAULT 1", "1"},
	}

	for _, col := range columns {
		if !columnExists("users", col.name) {
			log.Printf("为users表添加字段: %s", col.name)
			_, err := DB.Exec(fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", col.name, col.definition))
			if err != nil {
				log.Printf("添加字段 %s 失败: %v", col.name, err)
			} else {
				log.Printf("✓ 字段 %s 添加成功", col.name)
			}
		}
	}
	return nil
}

func repairPostsTable() error {
	// 检查并添加可能缺失的字段
	columns := []struct {
		name         string
		definition   string
		defaultValue string
	}{
		{"type", "TEXT DEFAULT 'text'", "'text'"},
		{"user_id", "INTEGER", ""},
		{"attachment_url", "TEXT", ""},
		{"attachment_type", "TEXT", ""},
	}

	for _, col := range columns {
		if !columnExists("posts", col.name) {
			log.Printf("为posts表添加字段: %s", col.name)
			_, err := DB.Exec(fmt.Sprintf("ALTER TABLE posts ADD COLUMN %s %s", col.name, col.definition))
			if err != nil {
				log.Printf("添加字段 %s 失败: %v", col.name, err)
			} else {
				log.Printf("✓ 字段 %s 添加成功", col.name)
			}
		}
	}
	return nil
}

func repairCommentsTable() error {
	// 检查并添加可能缺失的字段
	columns := []struct {
		name         string
		definition   string
		defaultValue string
	}{
		{"parent_id", "INTEGER", ""},
		{"coins", "INTEGER DEFAULT 0", "0"},
		{"reply_count", "INTEGER DEFAULT 0", "0"},
		{"user_id", "INTEGER", ""},
	}

	for _, col := range columns {
		if !columnExists("comments", col.name) {
			log.Printf("为comments表添加字段: %s", col.name)
			_, err := DB.Exec(fmt.Sprintf("ALTER TABLE comments ADD COLUMN %s %s", col.name, col.definition))
			if err != nil {
				log.Printf("添加字段 %s 失败: %v", col.name, err)
			} else {
				log.Printf("✓ 字段 %s 添加成功", col.name)
			}
		}
	}
	return nil
}

func repairBoardsTable() error {
	// boards表通常不需要修复，但可以在这里添加逻辑
	return nil
}

// columnExists 检查字段是否存在
func columnExists(tableName, columnName string) bool {
	rows, err := DB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			continue
		}

		if name == columnName {
			return true
		}
	}
	return false
}

// repairAppsTable 修复apps表
func repairAppsTable() error {
	// 检查并添加可能缺失的字段
	columns := []struct {
		name         string
		definition   string
		defaultValue string
	}{
		{"main_category", "TEXT", ""},
		{"sub_category", "TEXT", ""},
		{"channel", "TEXT", ""},
		{"share_desc", "TEXT", ""},
		{"developer_name", "TEXT", ""},
		{"ad_level", "TEXT", ""},
		{"payment_type", "TEXT", ""},
		{"operation_type", "TEXT", ""},
	}

	for _, col := range columns {
		if !columnExists("apps", col.name) {
			log.Printf("为apps表添加字段: %s", col.name)
			_, err := DB.Exec(fmt.Sprintf("ALTER TABLE apps ADD COLUMN %s %s", col.name, col.definition))
			if err != nil {
				log.Printf("添加字段 %s 失败: %v", col.name, err)
			} else {
				log.Printf("✓ 字段 %s 添加成功", col.name)
			}
		}
	}
	return nil
}
