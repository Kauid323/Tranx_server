//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	// 打开数据库
	db, err := sql.Open("sqlite", "./taruapp.db")
	if err != nil {
		log.Fatal("打开数据库失败:", err)
	}
	defer db.Close()

	fmt.Println("开始数据库迁移...")

	// 检查并添加 coins 字段
	_, err = db.Exec("ALTER TABLE users ADD COLUMN coins INTEGER DEFAULT 0")
	if err != nil {
		fmt.Println("coins 字段可能已存在或添加失败:", err)
	} else {
		fmt.Println("✓ 添加 coins 字段成功")
	}

	// 检查并添加 exp 字段
	_, err = db.Exec("ALTER TABLE users ADD COLUMN exp INTEGER DEFAULT 0")
	if err != nil {
		fmt.Println("exp 字段可能已存在或添加失败:", err)
	} else {
		fmt.Println("✓ 添加 exp 字段成功")
	}

	// 检查并添加 user_level 字段
	_, err = db.Exec("ALTER TABLE users ADD COLUMN user_level INTEGER DEFAULT 1")
	if err != nil {
		fmt.Println("user_level 字段可能已存在或添加失败:", err)
	} else {
		fmt.Println("✓ 添加 user_level 字段成功")
	}

	fmt.Println("\n数据库迁移完成！")
	fmt.Println("现在可以重新启动服务器。")
}
