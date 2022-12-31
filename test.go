package main

import (
	"fmt"
	"mysql/mysql"
	"time"
)

func main() {
	// 数据库配置
	dbconf := map[string]string{
		"password": "123456",
		"username": "root",
		"network":  "tcp",
		"server":   "127.0.0.1",
		"port":     "33065",
		"database": "tbox",
		"prefix":   "box_",
	}
	Db := new(mysql.Db)
	Db.DbConn = new(mysql.DbConnection)
	Db.DbConn.Connt(&dbconf)
	// 连接池配置
	// 设置与数据库建立连接的最大数目
	Db.DbConn.DB.SetMaxOpenConns(500)
	// 设置连接空闲的最大时间
	Db.DbConn.DB.SetConnMaxIdleTime(16)
	// 设置连接池中的最大闲置连接数
	Db.DbConn.DB.SetMaxIdleConns(10)
	// 设置连接可重用的最大时间
	Db.DbConn.DB.SetConnMaxLifetime(30 * time.Second)
	// 查询例子
	ret, _ := Db.Db("example").Field("count(id) tot,id").
		Where("id = ?", 3).
		Or("id = ?", 1).
		Or("id = ?", 2).
		OrIn("id", []interface{}{1, 2}).
		Group("carNo,id").
		Having("tot = ?", 1).
		Select()
	fmt.Println(ret)
	ret, _ = Db.Db("example").Field("id").
		Where("id = ?", 3).
		Or("id = ?", 1).
		Or("id = ?", 2).
		OrIn("id", []interface{}{1, 2}).
		Select()
	fmt.Println(ret)
	// 更新例子
	Db.Db("example").Where("id = ?", 1).
		OrIn("id", []interface{}{1, 2}).
		Update(map[string]interface{}{
			"carNoColor": "黄色",
		})
	// 写入例子
	/*
		Db.Db("example").Save(map[string]interface{}{
			"carNoColor":  "黄色",
		})*/
	// 删除例子
	// Db.Db("example").Where("id=?", 9614).Del()
}
