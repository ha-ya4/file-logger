package filelogger

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Logger ファイルへログ出力、ログローテーションなどをする
const (
	LoggerFlags = log.Ldate | log.Ltime | log.LstdFlags

	FileFlags = os.O_APPEND | os.O_CREATE | os.O_RDWR
)

var Logger *fileLogger

type fileLogger struct {
	sync.Mutex
	*LogFile
	Logger *log.Logger
	Conf   *Config
}

// Config loggerの設定を持つ構造体
type Config struct {
	Rotate      RotateConfig
	Mode        logMode // ログレベルによる出力の有無を切り替えるためのモード
	LoggerFlags int
	FilePath    string
	FilePerm    os.FileMode
	FileFlags   int
	Compress    bool
	Prefix      string
}

// RotateConfig ローテーションの設定をする構造体
type RotateConfig struct {
	MaxLine     int // 何行で次のファイルに移るか
	MaxRotation int // ファイル何枚ででローテーションするか
}

// LogFile ログファイルの設定、pathファイル自体を保持する構造体
type LogFile struct {
	Perm os.FileMode
	flag int
	fm   *fileNameManager
	file *os.File
}

// FileClose ファイルをクローズする
func (l *LogFile) FileClose() error {
	return l.file.Close()
}

func (l *LogFile) openFile() error {
	var err error
	l.file, err = os.OpenFile(l.fm.path, l.flag, l.Perm)
	return err
}

func (l *fileLogger) setOutput() error {
	var err error
	err = l.openFile()
	if err == nil {
		l.Logger.SetOutput(l.file)
	}
	return err
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
	fileName = filepath.Join(l.fm.dir, l.fm.getNameAddTimeNow())
	err = os.Rename(l.fm.path, fileName)
	if err != nil {
		rotation = false
		return fileName, rotation, err
	}

	l.FileClose()
	l.setOutput()
	fileList := containsSTRFileList(l.fm.dir, l.fm.name)
	err = l.deleteOldFile(fileList)

	return fileName, rotation, err
}

// isOverLine ローテーションが必要かチェックする。
func (l *fileLogger) isOverLine() bool {
	if l.Conf.Rotate.MaxLine <= 1 {
		return false
	}
	lineCount, _ := lineCounter(l.file)
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
	var err error
	if l.isOverFile(fileList) {
		oldFileName := oldFileName(fileList)
		err = os.Remove(filepath.Join(l.fm.dir, oldFileName))
	}

	return err
}

func handleError(err error) {
	if err != nil {
		logPrintln(err.Error())
	}
}
