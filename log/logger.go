package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Yamiyo/common/slack"
	"github.com/Yamiyo/common/timeutils"

	"github.com/sirupsen/logrus"
)

/*************************************************
Debug Level Setting
- debug
- info
- warning
- error
- fatal
- panic
*************************************************/

const (
	fileTag = "file"
	lineTag = "line"
	funcTag = "func"
)

type logConf struct { //執行時期的 log 功能配置
	showFileInfo bool //是否顯示 file name, func name, line number
	today        *time.Time
	logpath      string
	file         *os.File
}

var rtLogConf logConf
var logExit = make(chan error)
var wg sync.WaitGroup

// Init ...
func Init(env, level, logpath, duration, url, channel, hookLevel string, forceColor, fullTimestamp bool) {
	lv, _ := logrus.ParseLevel(level)
	hookLv, _ := logrus.ParseLevel(hookLevel)

	format := &logrus.TextFormatter{ForceColors: forceColor, FullTimestamp: fullTimestamp}

	if env == "dev" {
		InitLog(format, lv, hookLv, env, logpath, duration, url, channel, true, false)
	} else {
		InitLog(format, lv, hookLv, env, logpath, duration, url, channel, false, false)
	}
}

//InitLog config the log
func InitLog(format logrus.Formatter, level, hookLevel logrus.Level, env, logpath, duration, url, channel string, multiWriter, showFileInfo bool) {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(fmt.Sprintf("InitLog %v", err))
	}

	logrus.SetFormatter(format)
	logrus.SetLevel(level)

	logrus.AddHook(&slack.Hook{
		HookURL:        url,
		AcceptedLevels: slack.LevelThreshold(hookLevel),
		Channel:        channel,
		IconEmoji:      ":ghost:",
		Username:       "footbot",
		Env:            env,
	})

	now := time.Now().UTC()
	t := now.Truncate(d)

	fullpath := logpath + "/" + timeutils.Time2String(&t, "_") + "." + "log"

	if err := os.MkdirAll(filepath.Dir(fullpath), 0744); err != nil {
		Error(context.Background(), "error folder create : ", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(fullpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		Error(context.Background(), "error opening file: ", err)
		os.Exit(1)
	}

	if multiWriter {
		logrus.SetOutput(io.MultiWriter(f, os.Stdout))
	} else {
		logrus.SetOutput(os.Stdout)
	}

	rtLogConf.showFileInfo = showFileInfo
	rtLogConf.today = &t
	rtLogConf.logpath = logpath
	rtLogConf.file = f

	wg.Add(1)
	go func() {
		Debug(context.Background(), "Logger start fetching filename with datetime")
	loop:
		for {
			now := time.Now().UTC()
			next := now.Truncate(d)
			nt := next.Unix()
			t := rtLogConf.today.UTC().Unix()

			if nt > t {
				path := rtLogConf.logpath + "/" + timeutils.Time2String(&next, "_") + "." + "log"
				f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					panic(err)
				}

				if multiWriter {
					logrus.SetOutput(io.MultiWriter(f, os.Stdout))
				} else {
					logrus.SetOutput(f)
				}

				if err := rtLogConf.file.Close(); err != nil {
					panic(err)
				}

				rtLogConf.today = &next
				rtLogConf.file = f
			}

			time.Sleep(500 * time.Millisecond)

			select {
			case <-logExit:
				break loop
			default:
			}
		}
		wg.Done()
	}()
}

// Stop ...
func Stop() {
	logExit <- nil
	wg.Wait()
	Debug(context.Background(), "Logger stop fetching filename with datetime")
}

func getBaseName(fileName string, funcName string) (string, string) {
	return filepath.Base(fileName), filepath.Base(funcName)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(ctx context.Context, args ...interface{}) {
	message := message(ctx, "Debug", args...)
	logrus.Debug(message)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(ctx context.Context, msg string, args ...interface{}) {
	message := messagef(ctx, "Debug", msg, args...)
	logrus.Debug(message)
}

// Info logs a message at level Info on the standard logger.
func Info(ctx context.Context, args ...interface{}) {
	message := message(ctx, "Info", args...)
	logrus.Info(message)
}

// Infof logs a message at level Info on the standard logger.
func Infof(ctx context.Context, msg string, args ...interface{}) {
	message := messagef(ctx, "Info", msg, args...)
	logrus.Info(message)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(ctx context.Context, args ...interface{}) {
	message := message(ctx, "Warn", args...)
	logrus.Warn(message)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(ctx context.Context, msg string, args ...interface{}) {
	message := messagef(ctx, "Warn", msg, args...)
	logrus.Warn(message)
}

// Error logs a message at level Error on the standard logger.
func Error(ctx context.Context, args ...interface{}) {
	message := message(ctx, "Error", args...)
	logrus.Error(message)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(ctx context.Context, msg string, args ...interface{}) {
	message := messagef(ctx, "Error", msg, args...)
	logrus.Error(message)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(ctx context.Context, args ...interface{}) {
	message := message(ctx, "Panic", args...)
	logrus.Panic(message)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(ctx context.Context, msg string, args ...interface{}) {
	message := messagef(ctx, "Panic", msg, args...)
	logrus.Panic(message)
}

func message(ctx context.Context, level string, msg ...interface{}) string {
	message := Message{
		ChainID:     "",
		Level:       level,
		Version:     "v1.0.0",
		ServiceCode: "100",
		Time:        time.Now().UTC().Format(time.RFC3339),
		Msg:         fmt.Sprint(msg...),
	}
	if chainID := ctx.Value("ChainID"); chainID != nil {
		message.ChainID = chainID.(string)
	}
	m, _ := json.Marshal(message)

	return string(m)
}

func messagef(ctx context.Context, level string, msg string, args ...interface{}) string {
	message := Message{
		ChainID:     "",
		Level:       level,
		Version:     "v1.0.0",
		ServiceCode: "100",
		Time:        time.Now().UTC().Format(time.RFC3339),
		Msg:         fmt.Sprintf(msg, args...),
	}
	if chainID := ctx.Value("ChainID"); chainID != nil {
		message.ChainID = chainID.(string)
	}
	m, _ := json.Marshal(message)

	return string(m)
}
