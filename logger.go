package filelogger

import (
	"os"
	"log"
)

type level string

// loglevel?
const (
	DEBUG = level("[DEBUG]")
	INFO = level("[INFO]")
	WARN = level("[WARN]")
	ERROR = level("[ERROR]")
)

type fileLogger struct {
	*logFile
	logger *log.Logger
	global bool
	callPlace bool
}

type logFile struct {
	Perm os.FileMode
	flag int
	filePath string
	file *os.File
	custom bool
}

func NewLogFileCustom(filePath string, flag int, perm os.FileMode) *logFile {
	return &logFile {
		Perm: perm,
		flag: flag,
		filePath: filePath,
		custom: true,
	}
}

func newLogFileDefault(filePath string) *logFile {
	return &logFile {
		Perm: 0666,
		flag:  os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		filePath: filePath,
	}
}

func (l *logFile) close() error {
	return l.file.Close()
}

func (l *logFile) openFile() error {
	var err error
	l.file, err = os.OpenFile(l.filePath, l.flag, l.Perm)
	return err
}

type fileLoggerArgs struct {
	prefix string
	flags int
	custom bool
}

func NewfileLoggerArgsDefault() *fileLoggerArgs {
	return &fileLoggerArgs {
		prefix: "",
		flags: log.Ldate|log.Ltime|log.Lshortfile|log.LstdFlags,
		custom: true,
	}
}

func NewfileLoggerArgsCustom(prefix string, flags int) *fileLoggerArgs {
	return &fileLoggerArgs {
		prefix: prefix,
		flags: flags,
		custom: true,
	}
}

func NewfileLogger(filePath string) *fileLogger {
	file := newLogFileDefault(filePath)

	 return &fileLogger{
		logFile: file,
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
		callPlace: true,
	}
}

func NewfileLoggerCustom(lFile *logFile, lArgs *fileLoggerArgs) *fileLogger {
	var file *logFile
	var args *fileLoggerArgs

	if lFile.custom {
		file = lFile
	} else {
		file = newLogFileDefault(lFile.filePath)
	}

	if lArgs.custom {
		args = lArgs
	} else {
		args =  NewfileLoggerArgsDefault()
	}

	l := &fileLogger{
		logFile: file,
		logger: log.New(os.Stdout, args.prefix, args.flags),
		callPlace: true,
	}
	return l
}

func (l *fileLogger) FileClose() error {
	l.global = false
	return l.logFile.close()
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

func (l *fileLogger) SetOutput() error {
	var err error
	err = l.openFile()
	if err == nil {
		l.logger.SetOutput(l.file)
		l.global = true
	}
	return err
}

func (l *fileLogger) Println(logLevel level, outputLog string) {
	var callPlace string
	if l.callPlace {
		funcName := callFuncName()
		callPlace = createCallPlaceSTR(funcName)
	}
	l.logger.Printf("%s%s %s\n", callPlace, logLevel, outputLog)
}