package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mygola/config"
)

var DB *gorm.DB
var MongoClient *mongo.Client

func InitDB() {
	cfg := config.AppConfig
	dbConn := cfg.Database.Default
	conn := cfg.Database.Connections[dbConn]

	var err error

	switch dbConn {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			conn["username"], conn["password"], conn["host"], conn["port"], conn["database"])
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "pgsql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			conn["host"], conn["username"], conn["password"], conn["database"], conn["port"])
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(conn["path"]), &gorm.Config{})
	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			conn["username"], conn["password"], conn["host"], conn["port"], conn["database"])
		DB, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	case "mongodb":
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		clientOptions := options.Client().ApplyURI(conn["uri"])
		MongoClient, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		err = MongoClient.Ping(ctx, nil)
		if err != nil {
			log.Fatal("MongoDB ping failed:", err)
		}
		fmt.Println("MongoDB connected successfully")
	default:
		log.Fatal("Invalid DB connection")
	}

	if dbConn != "mongodb" && err != nil {
		log.Fatal(err)
	}

	if dbConn != "mongodb" {
		fmt.Println(dbConn, "connected successfully")
	}
}
