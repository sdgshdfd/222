package main

import (
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func tableToStructName(table string) string {
	// 去掉前缀"tb_"，首字母大写
	name := strings.TrimPrefix(table, "tb_")
	// 下划线转驼峰
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func main() {
	// 连接 SQLite 数据库
	db, err := gorm.Open(sqlite.Open("../../filesys.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 创建生成器实例
	g := gen.NewGenerator(gen.Config{
		OutPath:      "../../dao", // 输出目录
		ModelPkgPath: "model",     // model 包路径
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db) // 绑定数据库

	// 获取所有表名，排除sqlite系统表
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables)

	// 只生成业务表的 model
	var models []interface{}
	for _, table := range tables {
		structName := tableToStructName(table)
		models = append(models, g.GenerateModelAs(table, structName))
	}

	// 生成 dao 到 dao 目录
	g.ApplyBasic(models...)

	// 执行代码生成
	g.Execute()
}
