package filelogger

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Logger ファイルへログ出力、ログローテーションなどをする
var Logger *fileLogger

func init() {
	Logger = newfileLogger()
}

type fileLogger struct {
	sync.Mutex
	*LogFile
	logger     *log.Logger
	mode       logMode // ログレベルによる出力の有無を切り替えるためのモード
	rotateConf RotateConfig
	callPlace  bool
	depth      int
}

// RotateConfig ローテーションの設定をする構造体
type RotateConfig struct {
	MaxLine     int // 何行で次のファイルに移るか
	MaxRotation int // ファイル何枚ででローテーションするか
}

func newfileLogger() *fileLogger {
	file := newLogFileDefault("")
	args := newLoggerArgsDefault()

	return &fileLogger{
		LogFile:   file,
		logger:    log.New(os.Stdout, args.prefix, args.flags),
		mode:      ModeDebug,
		callPlace: true,
		depth:     4,
	}
}

// LoggerArgs log.Loggerの設定値を保持する
type LoggerArgs struct {
	prefix string
	flags  int
	custom bool
}

// newLoggerArgsDefault log.Loggerのデフォルトの設定値を持つ*LoggerArgs返す
func newLoggerArgsDefault() *LoggerArgs {
	return &LoggerArgs{
		prefix: "",
		flags:  log.Ldate | log.Ltime | log.LstdFlags,
	}
}

// NewLoggerArgsCustom log.Loggerの設定値を引数として受け取り、*LoggerArgsを返す
func NewLoggerArgsCustom(prefix string, flags int) *LoggerArgs {
	return &LoggerArgs{
		prefix: prefix,
		flags:  flags,
		custom: true,
	}
}

// LogFile ログファイルの設定、pathファイル自体を保持する構造体
type LogFile struct {
	Perm   os.FileMode
	flag   int
	fm     *fileNameManager
	file   *os.File
	custom bool
}

// newLogFileDefault ログファイルのデフォルトの設定をセットして*LogFileを返す
func newLogFileDefault(filePath string) *LogFile {
	fm := newFileNameManager(filePath)
	return &LogFile{
		Perm: 0666,
		flag: os.O_APPEND | os.O_CREATE | os.O_RDWR,
		fm:   fm,
	}
}

// NewLogFileCustom ログファイルの設定を引数として受け取り*LogFileを返す
func NewLogFileCustom(filePath string, flag int, perm os.FileMode) *LogFile {
	fm := newFileNameManager(filePath)
	return &LogFile{
		Perm:   perm,
		flag:   flag,
		fm:     fm,
		custom: true,
	}
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

// Custom log.Loggerとファイルの設定をする
func (l *fileLogger) Custom(lFile *LogFile, lArgs *LoggerArgs) {
	var file *LogFile
	var args *LoggerArgs

	if lFile.custom {
		file = lFile
	} else {
		file = newLogFileDefault("")
	}

	if lArgs.custom {
		args = lArgs
	} else {
		args = newLoggerArgsDefault()
	}

	l.LogFile = file
	l.logger = log.New(os.Stdout, args.prefix, args.flags)
}

func (l *fileLogger) setOutput() error {
	var err error
	err = l.openFile()
	if err == nil {
		l.logger.SetOutput(l.file)
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
	if l.rotateConf.MaxLine <= 1 {
		return false
	}
	lineCount, _ := lineCounter(l.file)
	return lineCount > l.rotateConf.MaxLine
}

// isOverFile セットされているローテーションするファイル数に達しているかチェックする。
func (l *fileLogger) isOverFile(fileList []os.FileInfo) bool {
	if l.rotateConf.MaxRotation <= 1 {
		return false
	}

	return len(fileList) > l.rotateConf.MaxRotation
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
