package filelogger

import (
	"io/ioutil"
	//"fmt"
	"strings"
	"testing"
	"os"
	//"path/filepath"
	//"time"
)

var dirPath = "./logtest"

func TestMain(m *testing.M) {
	os.Mkdir(dirPath, 0777)
	Logger.SetFilePath(dirPath + "/test.log")

	code := m.Run()

	os.RemoveAll("./logtest")

	os.Exit(code)
}

// ファイルにログが出力されているかのテスト
func TestLoggerOutput(t *testing.T) {
	path := dirPath + "/t.log"
	Logger.SetFilePath(path)
	Logger.Rprintln(ERROR, "test err")

	f, err := os.Open(path)
	tfatal(t, err)
	b, err := ioutil.ReadAll(f)
	tfatal(t, err)

	expect := "2019/11/27 23:31: logger_test.go:28:[ERROR] test err"

	if strings.Contains(string(b), expect) {
		t.Errorf("ERR TestLoggerOutput:\n b=%s\nexpect=%s\n", string(b), expect)
	}
}

func tfatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err.Error())
	}
}
