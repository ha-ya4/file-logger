package filelogger

import (
	"errors"
	"log"
	"os"
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

func (l *fileLogger) Println(logLevel level, outputLog string) {
	l.Mutex.Lock()
	l.setOutput()

	if l.isOverLine() {
		l.FileClose()
		l.rotation()
	}

	callPlace := l.findCallPlace()
	l.logger.Printf("%s%s %s\n", callPlace, logLevel, outputLog)

	l.FileClose()
	l.Mutex.Unlock()
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

func (l *fileLogger) rotation() error {
	fileName := createFileName(l.filePath)
	if err := os.Rename(l.filePath, fileName); err != nil {
		return err
	}
	l.setOutput()
	return nil
}

func (l *fileLogger) isOverLine() bool {
	if l.maxLine <= 0 {
		return false
	}
	lineCount, _ := lineCounter(l.file)
	return lineCount > l.maxLine
}

func (l *fileLogger) isOverFile() bool {
	if l.maxLine <= 0 {
		return false
	}
	lineCount, _ := lineCounter(l.file)
	return lineCount > l.maxLine
}
