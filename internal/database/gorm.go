package database

import (
	model2 "backend/internal/model"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GormDB *gorm.DB

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
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?timezone=UTC",
		cfg.Username, cfg.Password, cfg.Host, port, cfg.DBName)
}

func InitGorm() error {
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

	var err error
	url := dbConfig.BuildUrl()
	GormDB, err = gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, _ := GormDB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动迁移
	if err := GormDB.AutoMigrate(&model2.Hitokoto{}, &model2.User{}, &model2.RefreshToken{}); err != nil {
		return err
	}
	sql := "SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 99999) FROM users));"
	if err := GormDB.Exec(sql).Error; err != nil {
		return fmt.Errorf("设置序列起始值失败: %v", err)
	}
	log.Println("Gorm 初始化成功")
	return nil
}

func CloseGorm() {
	if GormDB != nil {
		sqlDB, _ := GormDB.DB()
		err := sqlDB.Close()
		if err != nil {
			log.Fatalf("Gorm 关闭失败: %v", err)
		}
	}
}
