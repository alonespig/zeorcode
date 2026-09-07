package conn

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMysqlClient 按配置建立 GORM 连接并配置连接池，返回 *gorm.DB 和关闭函数。
// 不做 AutoMigrate（表结构由 database/schema.sql + 迁移脚本手动维护）；
// 不维护全局单例，由 DI 在组合根处创建并注入。
func NewMysqlClient() (*gorm.DB, func()) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("db.user"),
		viper.GetString("db.password"),
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.name"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database connection pool", err)
	}

	maxOpenConns := positiveOrDefault(viper.GetInt("db.pool.max_open_conns"), 50)
	maxIdleConns := positiveOrDefault(viper.GetInt("db.pool.max_idle_conns"), 10)
	if maxIdleConns > maxOpenConns {
		maxIdleConns = maxOpenConns
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(positiveOrDefault(
		viper.GetInt("db.pool.conn_max_lifetime_seconds"), 3600,
	)) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(positiveOrDefault(
		viper.GetInt("db.pool.conn_max_idle_time_seconds"), 600,
	)) * time.Second)

	return db, func() {
		_ = sqlDB.Close()
	}
}

func positiveOrDefault(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
