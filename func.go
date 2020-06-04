package filelogger

import (
	"log"
	"os"
)

// Initialize Loggerを初期化する。
func Initialize(conf *Config) {
	file := LogFile{
		Perm: conf.FilePerm,
		flag: conf.FileFlags,
		fm:   newFileNameManager(conf.FilePath),
	}
	Logger = &fileLogger{
		LogFile: &file,
		Logger:  log.New(os.Stdout, conf.Prefix, conf.LoggerFlags),
		Conf:    conf,
	}
}

// Rprintln ログを出力する
// 最初にロックをかけ、ローテーションが必要なら現在のファイルの名前にローテーション時の日時を付与し、次のファイルに移る。
// この関数が呼び出されたファイル名と行数を取得し、ログのタイプ、ログと一緒に出力する。
// ローテーションした場合はロック解除後にファイルの圧縮を行う
func Rprintln(logLevel logLevel, output string) {
	if Logger.Conf.Mode.isNoOutput(logLevel) {
		return
	}

	var err error
	Logger.Mutex.Lock()
	err = Logger.setOutput()
	handleError(err)

	prevFileName, rotation, err := Logger.rotation()
	handleError(err)

	Logger.Logger.Printf("[%s] %s\n", logLevel, output)

	Logger.FileClose()
	Logger.Mutex.Unlock()

	if rotation {
		err = CompressFile(prevFileName)
		handleError(err)
	}
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
