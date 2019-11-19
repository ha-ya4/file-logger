package filelogger

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	//"fmt"
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
	*logFile
	logger      *log.Logger
	maxLine     int
	maxRotation int
	callPlace   bool
}

func newfileLogger() *fileLogger {
	file := NewLogFileDefault("")
	args := NewLoggerArgsDefault()

	return &fileLogger{
		logFile:   file,
		logger:    log.New(os.Stdout, args.prefix, args.flags),
		callPlace: true,
	}
}

type fileLoggerArgs struct {
	prefix string
	flags  int
	custom bool
}

func NewLoggerArgsDefault() *fileLoggerArgs {
	return &fileLoggerArgs{
		prefix: "",
		flags:  log.Ldate | log.Ltime | log.LstdFlags,
	}
}

func NewLoggerArgsCustom(prefix string, flags int) *fileLoggerArgs {
	return &fileLoggerArgs{
		prefix: prefix,
		flags:  flags,
		custom: true,
	}
}

type logFile struct {
	Perm     os.FileMode
	flag     int
	filePath string
	file     *os.File
	custom   bool
}

func NewLogFileCustom(filePath string, flag int, perm os.FileMode) *logFile {
	return &logFile{
		Perm:     perm,
		flag:     flag,
		filePath: filePath,
		custom:   true,
	}
}

func NewLogFileDefault(filePath string) *logFile {
	return &logFile{
		Perm:     0666,
		flag:     os.O_APPEND | os.O_CREATE | os.O_RDWR,
		filePath: filePath,
	}
}

func (l *logFile) SetFilePath(path string) {
	l.filePath = path
}

func (l *logFile) FileClose() error {
	return l.file.Close()
}

func (l *logFile) openFile() error {
	var err error

	if l.filePath == "" {
		err = errors.New(ErrFilePath)
	}

	l.file, err = os.OpenFile(l.filePath, l.flag, l.Perm)
	return err
}

func (l *fileLogger) Custom(lFile *logFile, lArgs *fileLoggerArgs) {
	var file *logFile
	var args *fileLoggerArgs

	if lFile.custom {
		file = lFile
	} else {
		file = NewLogFileDefault(lFile.filePath)
	}

	if lArgs.custom {
		args = lArgs
	} else {
		args = NewLoggerArgsDefault()
	}

	l.logFile = file
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

func (l *fileLogger) SetMaxLine(ml int) {
	l.maxLine = ml
}

func (l *fileLogger) SetMaxRotation(mr int) {
	l.maxRotation = mr
}

func (l *fileLogger) Println(logLevel level, outputLog string) {
	l.Mutex.Lock()
	l.setOutput()

	var rotation bool
	var prevFileName string
	if l.isOverLine() {
		l.FileClose()
		prevFileName, rotation = l.rotation()
	}

	callPlace := l.findCallPlace()
	l.logger.Printf("%s%s %s\n", callPlace, logLevel, outputLog)

	l.FileClose()
	l.Mutex.Unlock()

	if rotation {
		l.postProcessing(prevFileName)
	}
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

func (l *fileLogger) rotation() (string, bool) {
	complete := true
	fileName := createFileName(l.filePath)
	err := os.Rename(l.filePath, fileName)
	if err != nil {
		complete = false
	}
	l.setOutput()

	return fileName, complete
}

func (l *fileLogger) isOverLine() bool {
	if l.maxLine <= 0 {
		return false
	}
	lineCount, _ := lineCounter(l.file)
	return lineCount > l.maxLine
}

func (l *fileLogger) isOverFile(fileList []os.FileInfo) bool {
	if l.maxRotation <= 0 {
		return false
	}

	len := len(fileList)
	return len > l.maxRotation
}

func (l *fileLogger) postProcessing(prevFileName string) {
	dir := filepath.Dir(l.filePath)
	name := filepath.Base(l.filePath)
	fileList := containsSTRFileList(dir, name)
	if l.isOverFile(fileList) {
		oldFileName := oldFileName(fileList)
	}
}
