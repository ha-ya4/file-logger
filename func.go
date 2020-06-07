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
		perm: conf.FilePerm,
		flag: conf.FileFlags,
		fm:   newFileNameManager(conf.FilePath),
	}
	return &fileLogger{
		file:   &file,
		Logger: log.New(os.Stdout, conf.Prefix, conf.LoggerFlags),
		Conf:   conf,
	}
}

// LogPrintf ログレベルによる出力の有無を加えたlogパッケージのPrintf
func LogPrintf(logLevel string, format string, v ...interface{}) {
	if Logger.shouldNotOutput(logLevel) {
		return
	}
	log.Printf(format, v...)
}

// LogPrintln ログレベルによる出力の有無を加えたlogパッケージのPrintln
func LogPrintln(logLevel string, v ...interface{}) {
	if Logger.shouldNotOutput(logLevel) {
		return
	}
	log.Println(v...)
}

// LogPrint ログレベルによる出力の有無を加えたlogパッケージのPrint
func LogPrint(logLevel string, v ...interface{}) {
	if Logger.shouldNotOutput(logLevel) {
		return
	}
	log.Print(v...)
}

// Rprintf ローテーションとログレベルによる出力の有無を加えたlog.LoggerのPrintf
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

//******************************************************
// ショートカット系関数
//******************************************************

// SetPrefix prefixをセットする
func SetPrefix(prefix string) {
	Logger.Logger.SetPrefix(prefix)
}

// SetFlags loggerのフラグをセットする
func SetFlags(flags int) {
	Logger.Logger.SetFlags(flags)
}

/* 現時点で必要ないと思うのけど今後の変更でまた追加したくなる可能性があるのでコメントアウトしておく
// SetFilePath 受け取ったログファイルのpathをnewFileNameManagerに渡し、それをfmフィールドにセットする
func SetFilePath(path string) {
	Logger.file.fm = newFileNameManager(path)
}

// SetFileFlags
func SetFileFlags(flag int) {
	Logger.file.flag = flag
}

// SetFilePerm
func SetFilePerm(perm os.FileMode) {
	Logger.file.perm = perm
}
*/
