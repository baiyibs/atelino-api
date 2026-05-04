package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

type DbConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	DBName   string
}

func (cfg DbConfig) BuildUrl() string {
	port := cfg.Port
	if port == 0 {
		port = 5432
	}
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password, cfg.Host, port, cfg.DBName)
}

func Init() error {
	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	port := 5432
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		} else {
			log.Printf("端口号 %q 无效,使用默认端口号 5432", portStr)
		}
	}

	dbConfig := DbConfig{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		DBName:   dbName,
	}
	url := dbConfig.BuildUrl()
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return err
	}

	config.MaxConns = 10 // 最大连接数
	config.MinConns = 2  // 最小空闲连接数
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	// 测试连通性
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("数据库初始化成功")
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
