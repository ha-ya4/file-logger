package filelogger

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

// loggerとファイルのフラグ
const (
	LoggerFlags = log.Ldate | log.Ltime | log.LstdFlags
	FileFlags   = os.O_APPEND | os.O_CREATE | os.O_RDWR
)

// 基本的なログレベルをpackage側で定義
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)

// 基本的なログモードをpackage側で定義
const (
	ModeDebug      = "DebugMode"
	ModeProduction = "ProductionMode"
)

// Logger ファイルへログ出力、ログローテーションなどをする
var Logger *fileLogger

type fileLogger struct {
	sync.Mutex
	file   *LogFile
	Logger *log.Logger
	Conf   *Config
}

// Config loggerの設定を持つ構造体
type Config struct {
	Rotate       RotateConfig
	Mode         string // ログレベルによる出力の有無を切り替えるためのモード
	LoggerFlags  int
	FilePath     string
	FilePerm     os.FileMode
	FileFlags    int
	Compress     bool
	Prefix       string
	LogLevelConf LogLevelConfig
}

// LogFile ログファイルの設定、pathファイル自体を保持する構造体
type LogFile struct {
	perm os.FileMode
	flag int
	fm   *fileNameManager
	file *os.File
}

// RotateConfig ローテーションの設定をする構造体
type RotateConfig struct {
	MaxLine     int // 何行で次のファイルに移るか
	MaxRotation int // ファイル何枚ででローテーションするか
}

// LogLevelConfig LogLevelConfのスライス
type LogLevelConfig []LevelConfig

// LevelConfig 指定したモードのときに指定したレベルのログを出力しないための設定
type LevelConfig struct {
	Mode          string
	ExcludedLevel []string
}

func (llc LogLevelConfig) findMode(mode string) (int, bool) {
	for idx, l := range llc {
		if l.Mode == mode {
			return idx, true
		}
	}
	return 0, false
}

func (lc LevelConfig) isExcluded(level string) bool {
	for _, el := range lc.ExcludedLevel {
		if el == level {
			return true
		}
	}
	return false
}

func (l *fileLogger) setOutput() error {
	var err error
	l.file.file, err = os.OpenFile(l.file.fm.path, l.file.flag, l.file.perm)
	if err == nil {
		l.Logger.SetOutput(l.file.file)
	}
	return err
}

// 最初にロックをかけ、ローテーションが必要なら現在のファイルの名前にローテーション時の日時を付与し、次のファイルに移る。
// この関数が呼び出されたファイル名と行数を取得し、ログのタイプ、ログと一緒に出力する。
// ローテーションした場合はロック解除後にファイルの圧縮を行う
func (l *fileLogger) logOutput(logLevel string, printFunc func()) {
	// loglevelの設定を見て出力の必要がなければリターン
	if l.shouldNotOutput(logLevel) {
		return
	}

	var err error
	l.Mutex.Lock()
	if err = l.setOutput(); err != nil {
		logPrintln(err.Error())
	}

	prevFileName, rotation, err := l.rotation()
	if err != nil {
		logPrintln(err.Error())
	}

	printFunc()

	if err = l.file.file.Close(); err != nil {
		logPrintln(err.Error())
	}
	l.Mutex.Unlock()

	if rotation && l.Conf.Compress {
		if err = CompressFile(prevFileName); err != nil {
			logPrintln(err.Error())
		}
	}
}

// rotation セットされているファイルに書き込む最大行数に達しているかチェックし、必要なら次のファイルを作成しアウトプット先としてセットする。
// 前のファイルにはローテーション時の日時を付与した名前に変更する。
// 名前の変更に成功してから前のファイルのクローズをしている。
func (l *fileLogger) rotation() (string, bool, error) {
	var err error
	var rotation bool
	var fileName string

	if !l.isOverLine() {
		return fileName, rotation, err
	}

	rotation = true
	fileName = filepath.Join(l.file.fm.dir, l.file.fm.getNameAddTimeNow())
	err = os.Rename(l.file.fm.path, fileName)
	if err != nil {
		rotation = false
		return fileName, rotation, err
	}

	if err = l.file.file.Close(); err != nil {
		return fileName, rotation, err
	}
	if l.setOutput(); err != nil {
		return fileName, rotation, err
	}

	fileList := containsSTRFileList(l.file.fm.dir, l.file.fm.name)
	if l.isOverFile(fileList) {
		err = l.deleteOldFile(fileList)
	}

	return fileName, rotation, err
}

// isOverLine ローテーションが必要かチェックする。
func (l *fileLogger) isOverLine() bool {
	if l.Conf.Rotate.MaxLine <= 1 {
		return false
	}
	lineCount, _ := lineCounter(l.file.file)
	return lineCount > l.Conf.Rotate.MaxLine
}

// isOverFile セットされているローテーションするファイル数に達しているかチェックする。
func (l *fileLogger) isOverFile(fileList []os.FileInfo) bool {
	if l.Conf.Rotate.MaxRotation <= 1 {
		return false
	}

	return len(fileList) > l.Conf.Rotate.MaxRotation
}

// deleteOldFile 一番古いログファイルを削除する必要があるかチェックし、必要なら削除する
func (l *fileLogger) deleteOldFile(fileList []os.FileInfo) error {
	oldFileName := oldFileName(fileList)
	return os.Remove(filepath.Join(l.file.fm.dir, oldFileName))
}

func (l *fileLogger) shouldNotOutput(level string) bool {
	idx, exist := l.Conf.LogLevelConf.findMode(l.Conf.Mode)
	if !exist {
		return false
	}
	if l.Conf.LogLevelConf[idx].isExcluded(level) {
		return true
	}
	return false
}
