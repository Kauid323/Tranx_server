package database

import (
	"database/sql"
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

// createTables 创建数据表
func createTables() error {
	// 创建用户表
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
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
	);`

	// 创建关注关系表
	followsTable := `
	CREATE TABLE IF NOT EXISTS follows (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		followed_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, followed_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (followed_id) REFERENCES users(id)
	);`

	// 创建签到记录表
	checkInsTable := `
	CREATE TABLE IF NOT EXISTS check_ins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		check_date TEXT NOT NULL,
		check_time DATETIME NOT NULL,
		reward INTEGER DEFAULT 50,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, check_date),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	// 创建Token表
	tokensTable := `
	CREATE TABLE IF NOT EXISTS tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	// 创建用户标签表
	userTagsTable := `
	CREATE TABLE IF NOT EXISTS user_tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		tag_name TEXT NOT NULL,
		tag_color TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	// 创建板块表
	boardsTable := `
	CREATE TABLE IF NOT EXISTS boards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建帖子表
	postsTable := `
	CREATE TABLE IF NOT EXISTS posts (
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
	);`

	// 创建评论表
	commentsTable := `
	CREATE TABLE IF NOT EXISTS comments (
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
	);`

	// 创建收藏夹表
	favoriteFoldersTable := `
	CREATE TABLE IF NOT EXISTS favorite_folders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		is_public BOOLEAN DEFAULT 0,
		item_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	// 创建收藏夹项目表
	favoriteItemsTable := `
	CREATE TABLE IF NOT EXISTS favorite_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		folder_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(folder_id, post_id),
		FOREIGN KEY (folder_id) REFERENCES favorite_folders(id),
		FOREIGN KEY (post_id) REFERENCES posts(id)
	);`

	// 创建帖子点赞记录表
	postLikesTable := `
	CREATE TABLE IF NOT EXISTS post_likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, post_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (post_id) REFERENCES posts(id)
	);`

	// 创建评论点赞记录表
	commentLikesTable := `
	CREATE TABLE IF NOT EXISTS comment_likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		comment_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, comment_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (comment_id) REFERENCES comments(id)
	);`

	// 创建浏览历史表
	viewHistoriesTable := `
	CREATE TABLE IF NOT EXISTS view_histories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		viewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (post_id) REFERENCES posts(id)
	);`

	// 创建索引
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
	}

	// 执行建表语句
	if _, err := DB.Exec(usersTable); err != nil {
		return err
	}
	if _, err := DB.Exec(followsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(checkInsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(tokensTable); err != nil {
		return err
	}
	if _, err := DB.Exec(userTagsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(boardsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(postsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(commentsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(favoriteFoldersTable); err != nil {
		return err
	}
	if _, err := DB.Exec(favoriteItemsTable); err != nil {
		return err
	}
	if _, err := DB.Exec(postLikesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(commentLikesTable); err != nil {
		return err
	}
	if _, err := DB.Exec(viewHistoriesTable); err != nil {
		return err
	}

	// 创建索引
	for _, index := range indexes {
		if _, err := DB.Exec(index); err != nil {
			return err
		}
	}

	log.Println("数据表创建成功")

	// 创建默认主板块
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM boards WHERE id = 1").Scan(&count)
	if count == 0 {
		_, err := DB.Exec(
			"INSERT INTO boards (id, name, description) VALUES (1, '综合讨论', '默认主板块，所有话题都可以在这里讨论')",
		)
		if err != nil {
			log.Printf("创建默认主板块失败: %v", err)
		} else {
			log.Println("默认主板块创建成功")
		}
	}

	return nil
}
