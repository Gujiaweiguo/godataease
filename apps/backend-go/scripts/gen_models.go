package main

import (
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"log"
	"strings"
)

var tables = flag.String("tables", "all", "tables to generate")
var dsn = flag.String("dsn", "root:Admin168@tcp(mysql8:3306)/dataease_dev?charset=utf8mb4&parseTime=True&loc=Local", "database dsn")
var outPath = flag.String("out", "./internal/domain/auto", "output directory")

func main() {
	flag.Parse()

	db, err := gorm.Open(mysql.Open(*dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var tableNames []string
	if *tables == "all" {
		if err := db.Raw("SELECT TABLE_NAME FROM information_schema.tables WHERE table_schema = DATABASE() AND TABLE_TYPE = 'BASE TABLE'").Scan(&tableNames).Error; err != nil {
			log.Fatalf("failed to get tables: %v", err)
		}
	} else {
		tableNames = strings.Split(*tables, ",")
	}

	fmt.Printf("Found %d tables to generate\n", len(tableNames))

	g := gen.NewGenerator(gen.Config{
		OutPath:      *outPath + "/query",
		ModelPkgPath: *outPath,
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db)

	for _, tableName := range tableNames {
		tableName = strings.TrimSpace(tableName)
		if tableName == "" {
			continue
		}
		fmt.Printf("Generating model for table: %s\n", tableName)
		g.GenerateModel(tableName)
	}

	g.Execute()

	fmt.Println("Done!")
}
