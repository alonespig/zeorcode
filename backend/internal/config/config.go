package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func Init() {
	configureEnvironment()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
}

func configureEnvironment() {
	// Keep config.yaml as the local-development default while allowing containers
	// and deployment systems to override any scalar key with ZEORCODE_* variables.
	// Example: db.pool.max_open_conns -> ZEORCODE_DB_POOL_MAX_OPEN_CONNS.
	viper.SetEnvPrefix("ZEORCODE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
