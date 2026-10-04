package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger  *zap.Logger
	sugared *zap.SugaredLogger
)

// Mode 日志模式
type Mode string

const (
	ModeDev  Mode = "dev"
	ModeProd Mode = "prod"
)

// 日志文件配置
const (
	logDir      = "runtime"
	logBaseName = "oj"
)

// Init 在程序入口调用一次。重复调用以最后一次为准
//
// 输出策略：
//   - stdout：按 mode 选 console / JSON encoder
//   - runtime/oj-YYYY-MM-DD.log：始终 JSON encoder（方便后续 grep/jq 分析）
//     按天切分，不做自动清理
func Init(mode Mode) {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var (
		stdoutEncoder zapcore.Encoder
		level         zapcore.Level
	)
	switch mode {
	case ModeProd:
		level = zapcore.InfoLevel
		stdoutEncoder = zapcore.NewJSONEncoder(encCfg)
	default:
		level = zapcore.DebugLevel
		consoleCfg := encCfg
		consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		stdoutEncoder = zapcore.NewConsoleEncoder(consoleCfg)
	}

	cores := []zapcore.Core{
		zapcore.NewCore(stdoutEncoder, zapcore.AddSync(os.Stdout), level),
	}

	if w, err := newDailyWriter(); err != nil {
		fmt.Fprintf(os.Stderr, "logger: file output disabled (%v)\n", err)
	} else {
		fileEncoder := zapcore.NewJSONEncoder(encCfg)
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(w), level))
	}

	core := zapcore.NewTee(cores...)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugared = logger.Sugar()

	zap.RedirectStdLog(logger.WithOptions(zap.AddCallerSkip(-1)))
}

// dailyWriter 按天切分的文件写入器：每次 Write 时检查日期，跨天就重开文件
// 不做后台 goroutine，没并发开销
type dailyWriter struct {
	mu          sync.Mutex
	file        *os.File
	currentDate string // "2006-01-02"
}

func newDailyWriter() (*dailyWriter, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir log dir: %w", err)
	}
	w := &dailyWriter{}
	if err := w.openForDate(time.Now()); err != nil {
		return nil, err
	}
	return w, nil
}

// Write 实现 io.Writer；跨天时换文件
func (w *dailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != w.currentDate {
		if err := w.openForDate(time.Now()); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

// Sync 实现 zapcore.WriteSyncer
func (w *dailyWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

// openForDate 关旧文件、开新文件；调用方持锁（或在 init 时单线程）
func (w *dailyWriter) openForDate(t time.Time) error {
	date := t.Format("2006-01-02")
	path := filepath.Join(logDir, logBaseName+"-"+date+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	w.file = f
	w.currentDate = date
	return nil
}

// L 返回结构化 logger
func L() *zap.Logger {
	if logger == nil {
		Init(ModeDev)
	}
	return logger
}

// S 返回 printf 风格 logger
func S() *zap.SugaredLogger {
	if sugared == nil {
		Init(ModeDev)
	}
	return sugared
}

// Sync 退出前 flush
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

func Debugw(msg string, kv ...any) { S().Debugw(msg, kv...) }
func Infow(msg string, kv ...any)  { S().Infow(msg, kv...) }
func Warnw(msg string, kv ...any)  { S().Warnw(msg, kv...) }
func Errorw(msg string, kv ...any) { S().Errorw(msg, kv...) }
func Fatalw(msg string, kv ...any) { S().Fatalw(msg, kv...) }
