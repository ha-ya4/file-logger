package filelogger

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var printTestLogger *fileLogger
var printTestFilePath = "./print_test.txt"

func init() {
	conf := &Config{
		LoggerFlags: LoggerFlags,
		FilePath:    printTestFilePath,
		FilePerm:    0666,
		FileFlags:   FileFlags,
	}
	printTestLogger = newFileLogger(conf)
}

/*
** print系はテストの都合上その関数ではなく、その関数と全く同じ内容の処理をやらせる
** 関数内で直接グローバル変数を見てるからこうなってしまう。
** package内グローバルじゃなくてpackage使用側にloggerを渡した方がいいか？
 */

func TestRprintln(t *testing.T) {
	testRprintln := func(logLevel string, v ...interface{}) {
		printTestLogger.logOutput(logLevel, func() {
			v[0] = fmt.Sprintf("[%s] %v", logLevel, v[0])
			printTestLogger.Logger.Println(v...)
		})
	}
	testRprintln("INFO", "test", "log")

	file, err := os.Open(printTestFilePath)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(file)
	assert.NoError(t, err)

	expect := "test log"
	assert.True(t, strings.Contains(string(b), expect))

	os.Remove(printTestFilePath)
	assert.NoError(t, err)
}

func TestRprintf(t *testing.T) {
	testRprintf := func(logLevel string, format string, v ...interface{}) {
		printTestLogger.logOutput(logLevel, func() {
			s := fmt.Sprintf("[%s] %v", logLevel, format)
			printTestLogger.Logger.Printf(s, v...)
		})
	}
	testRprintf("INFO", "%s %s", "test", "log")

	file, err := os.Open(printTestFilePath)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(file)
	assert.NoError(t, err)

	expect := "test log"
	assert.True(t, strings.Contains(string(b), expect))

	os.Remove(printTestFilePath)
	assert.NoError(t, err)
}

func TestRprint(t *testing.T) {
	testRprint := func(logLevel string, v ...interface{}) {
		printTestLogger.logOutput(logLevel, func() {
			v[0] = fmt.Sprintf("[%s] %v", logLevel, v[0])
			printTestLogger.Logger.Print(v...)
		})
	}
	testRprint("INFO", "test", "log")

	file, err := os.Open(printTestFilePath)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(file)
	assert.NoError(t, err)

	expect := "testlog"
	assert.True(t, strings.Contains(string(b), expect))

	os.Remove(printTestFilePath)
	assert.NoError(t, err)
}
