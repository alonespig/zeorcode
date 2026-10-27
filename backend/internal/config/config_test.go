package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestEnvironmentOverridesNestedConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("ZEORCODE_DB_HOST", "mysql")
	t.Setenv("ZEORCODE_DB_POOL_MAX_OPEN_CONNS", "25")

	configureEnvironment()
	viper.SetDefault("db.host", "127.0.0.1")
	viper.SetDefault("db.pool.max_open_conns", 50)

	if got := viper.GetString("db.host"); got != "mysql" {
		t.Fatalf("db.host = %q, want mysql", got)
	}
	if got := viper.GetInt("db.pool.max_open_conns"); got != 25 {
		t.Fatalf("db.pool.max_open_conns = %d, want 25", got)
	}
}
