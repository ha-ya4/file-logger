package filelogger

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// エラーメッセージ
var (
	ErrFilePath = "missing file path"
)

// Logger ファイルへログ出力、ログローテーションなどをする
var Logger *fileLogger

func init() {
	Logger = newfileLogger()
}

type level string

// loglevel?
const (
	DEBUG = level("[DEBUG]")
	INFO  = level("[INFO]")
	WARN  = level("[WARN]")
	ERROR = level("[ERROR]")
)

type fileLogger struct {
	sync.Mutex
	*LogFile
	logger     *log.Logger
	rotateConf RotateConfig
	callPlace  bool
}

// RotateConfig ローテーションの設定をする構造体
type RotateConfig struct {
	maxLine     int // 何行で次のファイルに移るか
	maxRotation int // ファイル何枚ででローテーションするか
}

func newfileLogger() *fileLogger {
	file := NewLogFileDefault("")
	args := NewLoggerArgsDefault()

	return &fileLogger{
		LogFile:   file,
		logger:    log.New(os.Stdout, args.prefix, args.flags),
		callPlace: true,
	}
}

// LoggerArgs log.Loggerの設定値を保持する
type LoggerArgs struct {
	prefix string
	flags  int
	custom bool
}

// NewLoggerArgsDefault log.Loggerのデフォルトの設定値を持つ*LoggerArgs返す
func NewLoggerArgsDefault() *LoggerArgs {
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

// NewLogFileDefault ログファイルのデフォルトの設定をセットして*LogFileを返す
func NewLogFileDefault(filePath string) *LogFile {
	fm := newFileNameManager(filePath)
	return &LogFile{
		Perm: 0666,
		flag: os.O_APPEND | os.O_CREATE | os.O_RDWR,
		fm:   fm,
	}
}

// SetFilePath 受け取ったログファイルのpathをnewFileNameManagerに渡し、それをfmフィールドにセットする
func (l *LogFile) SetFilePath(path string) {
	l.fm = newFileNameManager(path)
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

func (l *fileLogger) Custom(lFile *LogFile, lArgs *LoggerArgs) {
	var file *LogFile
	var args *LoggerArgs

	if lFile.custom {
		file = lFile
	} else {
		file = NewLogFileDefault("")
	}

	if lArgs.custom {
		args = lArgs
	} else {
		args = NewLoggerArgsDefault()
	}

	l.LogFile = file
	l.logger = log.New(os.Stdout, args.prefix, args.flags)
}

func (l *fileLogger) SetCallPlace(flag bool) {
	l.callPlace = flag
}

func (l *fileLogger) SetPrefix(prefix string) {
	l.logger.SetPrefix(prefix)
}

func (l *fileLogger) SetFlags(flags int) {
	l.logger.SetFlags(flags)
}

func (l *fileLogger) SetRotate(conf RotateConfig) {
	l.rotateConf = conf
}

// Println ログを出力する
// 最初にロックをかけ、ローテーションが必要なら現在のファイルの名前にローテーション時の日時を付与し、次のファイルに移る。
// この関数が呼び出されたファイル名と行数を取得し、ログのタイプ、ログと一緒に出力する。
// ローテーションした場合はロック解除後にファイルの圧縮を行う
func (l *fileLogger) Println(logLevel level, outputLog string) error {
	var err error
	l.Mutex.Lock()
	l.setOutput()

	prevFileName, rotation := l.rotation()
	if rotation {
		l.deleteOldFile()
	}

	callPlace := l.findCallPlace()
	l.logger.Printf("%s%s %s\n", callPlace, logLevel, outputLog)

	l.FileClose()
	l.Mutex.Unlock()

	if rotation {
		err = l.compressFile(prevFileName)
	}

	return err
}

func (l *fileLogger) findCallPlace() string {
	var cp string
	if l.callPlace {
		funcName := callFuncName()
		cp = createCallPlaceSTR(funcName)
	}
	return cp
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
func (l *fileLogger) rotation() (string, bool) {
	var rotation bool
	var fileName string
	if !l.isOverLine() {
		return fileName, rotation
	}

	rotation = true
	fileName = l.fm.getNameAddTimeNow()
	err := os.Rename(l.fm.path, fileName)
	if err != nil {
		rotation = false
		return fileName, rotation
	}

	l.FileClose()
	l.setOutput()
	return fileName, rotation
}

// isOverLine ローテーションが必要かチェックする。
func (l *fileLogger) isOverLine() bool {
	if l.rotateConf.maxLine <= 0 {
		return false
	}
	lineCount, _ := lineCounter(l.file)
	return lineCount > l.rotateConf.maxLine
}

// isOverFile セットされているローテーションするファイル数に達しているかチェックする。
func (l *fileLogger) isOverFile(fileList []os.FileInfo) bool {
	if l.rotateConf.maxRotation <= 0 {
		return false
	}

	len := len(fileList)
	return len > l.rotateConf.maxRotation
}

// deleteOldFile 一番古いログファイルを削除する必要があるかチェックし、必要なら削除する
func (l *fileLogger) deleteOldFile() error {
	var err error
	fileList := containsSTRFileList(l.fm.dir, l.fm.name)

	if l.isOverFile(fileList) {
		oldFileName := oldFileName(fileList)
		err = os.Remove(filepath.Join(l.fm.dir, oldFileName))
	}

	return err
}

func (l *fileLogger) compressFile(path string) error {
	var err error

	file, err := os.Open(path)
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	newFile, err := os.Create(path)
	if err != nil {
		return err
	}

	err = compress(newFile, b)
	return err
}
