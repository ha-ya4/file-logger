package filelogger

import (
	"fmt"
	"log"
	"os"
)

// Initialize Loggerを初期化する。
func Initialize(conf *Config) {
	Logger = newFileLogger(conf)
}

func newFileLogger(conf *Config) *fileLogger {
	file := LogFile{
		Perm: conf.FilePerm,
		flag: conf.FileFlags,
		fm:   newFileNameManager(conf.FilePath),
	}
	return &fileLogger{
		LogFile: &file,
		Logger:  log.New(os.Stdout, conf.Prefix, conf.LoggerFlags),
		Conf:    conf,
	}
}

// Rprintlf ローテーションとログレベルによる出力の有無を加えたlog.LoggerのPrintf
func Rprintf(logLevel string, format string, v ...interface{}) {
	Logger.logOutput(logLevel, func() {
		s := fmt.Sprintf("[%s] %v", logLevel, format)
		Logger.Logger.Printf(s, v...)
	})
}

// Rprintln ローテーションとログレベルによる出力の有無を加えたlog.LoggerのPrintln
func Rprintln(logLevel string, v ...interface{}) {
	Logger.logOutput(logLevel, func() {
		v[0] = fmt.Sprintf("[%s] %v", logLevel, v[0])
		Logger.Logger.Println(v...)
	})
}

// Rprint ローテーションとログレベルによる出力の有無を加えたlog.LoggerのPrint
func Rprint(logLevel string, v ...interface{}) {
	Logger.logOutput(logLevel, func() {
		v[0] = fmt.Sprintf("[%s] %v", logLevel, v[0])
		Logger.Logger.Print(v...)
	})
}

// LogPrintln パッケージlogのPrintln関数に引数のmsgを渡すだけの関数
// パッケージ使用側でパッケージlogをインポートしなくて済むように
func LogPrintln(msg string) {
	log.Println(msg)
}

// SetConfig
func SetRotate(conf RotateConfig) {
	Logger.Conf.Rotate = conf
}

func SetMode(mode logMode) {
	Logger.Conf.Mode = mode
}

// SetFilePath 受け取ったログファイルのpathをnewFileNameManagerに渡し、それをfmフィールドにセットする
func SetFilePath(path string) {
	Logger.fm = newFileNameManager(path)
}

func SetPrefix(prefix string) {
	Logger.Logger.SetPrefix(prefix)
}

func SetFlags(flags int) {
	Logger.Logger.SetFlags(flags)
}
