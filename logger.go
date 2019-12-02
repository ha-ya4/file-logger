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

// Rprintln ログを出力する
// 最初にロックをかけ、ローテーションが必要なら現在のファイルの名前にローテーション時の日時を付与し、次のファイルに移る。
// この関数が呼び出されたファイル名と行数を取得し、ログのタイプ、ログと一緒に出力する。
// ローテーションした場合はロック解除後にファイルの圧縮を行う
func Rprintln(logLevel level, output string) {
	var err error
	Logger.Mutex.Lock()
	err = Logger.setOutput()
	handleError(err)

	prevFileName, rotation, err := Logger.rotation()
	handleError(err)

	//callPlace := l.createCallPlace()
	//l.logger.Printf("%s%s %s\n", callPlace, logLevel, output)
	depth := 4
	Logger.println(logLevel, output, depth)

	Logger.FileClose()
	Logger.Mutex.Unlock()

	if rotation {
		err = CompressFile(prevFileName)
		handleError(err)
	}
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
	file := newLogFileDefault("")
	args := newLoggerArgsDefault()

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

// SetCallPlace 呼び出し位置と行数を出力するかのフラグをセットする
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

func (l *fileLogger) println(logLevel level, output string, depth int) {
	callPlace := l.createCallPlace(depth)
	l.logger.Printf("%s%s %s\n", callPlace, logLevel, output)
}

func (l *fileLogger) createCallPlace(depth int) string {
	var cp string
	if l.callPlace {
		funcName := callFuncName(depth)
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
	if l.rotateConf.maxLine <= 1 {
		return false
	}
	lineCount, _ := lineCounter(l.file)
	return lineCount > l.rotateConf.maxLine
}

// isOverFile セットされているローテーションするファイル数に達しているかチェックする。
func (l *fileLogger) isOverFile(fileList []os.FileInfo) bool {
	if l.rotateConf.maxRotation <= 1 {
		return false
	}

	return len(fileList) > l.rotateConf.maxRotation
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
