package filelogger

import (
	"log"
	"os"
	"io/ioutil"
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
	Perm   os.FileMode
	flag   int
	fm     *fileNameManager
	file   *os.File
	custom bool
}

func NewLogFileCustom(filePath string, flag int, perm os.FileMode) *logFile {
	fm := newFileNameManager(filePath)
	return &logFile{
		Perm:   perm,
		flag:   flag,
		fm:     fm,
		custom: true,
	}
}

func NewLogFileDefault(filePath string) *logFile {
	fm := newFileNameManager(filePath)
	return &logFile{
		Perm: 0666,
		flag: os.O_APPEND | os.O_CREATE | os.O_RDWR,
		fm:   fm,
	}
}

func (l *logFile) SetFilePath(path string) {
	l.fm = newFileNameManager(path)
}

func (l *logFile) FileClose() error {
	return l.file.Close()
}

func (l *logFile) openFile() error {
	var err error

	l.file, err = os.OpenFile(l.fm.path, l.flag, l.Perm)
	return err
}

func (l *fileLogger) Custom(lFile *logFile, lArgs *fileLoggerArgs) {
	var file *logFile
	var args *fileLoggerArgs

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

	prevFileName, rotation := l.rotation()
	if rotation {
		l.deleteOldFile()
	}

	callPlace := l.findCallPlace()
	l.logger.Printf("%s%s %s\n", callPlace, logLevel, outputLog)

	l.FileClose()
	l.Mutex.Unlock()

	if rotation {
		l.compressPrevFile(prevFileName)
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
	var rotation bool
	var fileName string
	if !l.isOverLine() {
		return fileName, rotation
	}

	rotation = true
	fileName = createFileName(l.fm.path)
	err := os.Rename(l.fm.path, fileName)
	if err != nil {
		rotation = false
		return fileName, rotation
	}

	l.FileClose()
	l.setOutput()
	return fileName, rotation
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

func (l *fileLogger) deleteOldFile() error {
	var err error
	fileList := containsSTRFileList(l.fm.dir, l.fm.name)

	if l.isOverFile(fileList) {
		oldFileName := oldFileName(fileList)
		err = os.Remove(filepath.Join(l.fm.dir, oldFileName))
	}

	return err
}

func (l *fileLogger) compressPrevFile(path string) error {
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
