package judgepool

import (
	"time"

	judgeapi "github.com/alonespig/zeorcode/backend/pkg/judge"

	"github.com/spf13/viper"
)

// New creates the go-judge pool used by synchronous code runs in the HTTP service.
func New() (*judgeapi.Pool, func()) {
	urls := viper.GetStringSlice("judge.urls")
	if len(urls) == 0 {
		if url := viper.GetString("judge.url"); url != "" {
			urls = []string{url}
		}
	}

	pool := judgeapi.NewPool(urls)
	pool.StartHealthCheck(15 * time.Second)
	return pool, pool.Stop
}
