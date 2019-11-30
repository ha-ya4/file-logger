package filelogger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestCallFuncName(t *testing.T) {
	funcName := func() string {
		return func() string {
			return callFuncName()
		}()
	}()

	currentPath, _ := os.Getwd()
	expected := getPath(currentPath) + ".TestCallFuncName"

	if funcName != expected {
		t.Errorf("\n関数名が一致していません\nfuncName=%s\nexpected=%s", funcName, expected)
	}
}

func TestFindCallLineAndFile(t *testing.T) {
	line, file := func() (int, string) {
		return func() (int, string) {
			return callFunc()
		}()
	}()

	expectedLine := 28
	if line != expectedLine {
		t.Errorf("\n行数が一致しません\nline=%d\nexpected=%d", line, expectedLine)
	}

	path := getPath(file)
	currentPath, _ := os.Getwd()
	expectedPath := getPath(currentPath) + "/helpers_test.go"
	if path != expectedPath {
		t.Errorf("\nファイル名が一致していません\nfile=%s\nexpected=%s", path, expectedPath)
	}
}

func TestCreateCallPlaceSTR(t *testing.T) {
	funcName := func() string {
		return func() string {
			return callFuncName()
		}()
	}()

	cps := createCallPlaceSTR(funcName)
	expectedCPS := "helpers_test.go:52:"

	if cps != expectedCPS {
		t.Errorf("\n呼び出し位置が一致しません\ncps=%s\nexpected=%s", cps, expectedCPS)
	}
}

func TestLineCounter(t *testing.T) {
	file, _ := os.Open("./helpers_test.txt")
	defer file.Close()
	count, _ := lineCounter(file)
	expectedCount := 13

	if count != expectedCount {
		t.Errorf("\nファイル行数が一致していません\ncount=%d\nexpected=%d", count, expectedCount)
	}
}

func TestContainsSTRFileList(t *testing.T) {
	path := "./"
	fileName := ".go"
	fileList := containsSTRFileList(path, fileName)
	len := len(fileList)
	expectedLen := 4

	if len != expectedLen {
		t.Errorf("\n期待されるファイル数ではありません\ncount=%d\nexpected=%d", len, expectedLen)
	}
}

func TestCompressAndUnfreeze(t *testing.T) {
	var err error
	c := []byte("Hello World!")

	b := &bytes.Buffer{}
	err = compress(b, c)
	if err != nil {
		t.Errorf("TestCompressAndUnfreeze: 圧縮に失敗しました")
	}
	if b.String() == string(c) {
		t.Errorf("TestCompressAndUnfreeze: 圧縮に失敗しました\nunfreezw=%s\nexpected=%s", b.String(), string(c))
	}

	bb, err := Unfreeze(b)
	if err != nil {
		t.Errorf("TestCompressAndUnfreeze: 解凍に失敗しました")
	}
	if bb.String() != string(c) {
		t.Errorf("TestCompressAndUnfreeze: 期待される結果が得られませんでした\nunfreezw=%s\nexpected=%s", bb.String(), string(c))
	}
}

// /home/user/golang/src/github.com/ha-ya4/my-package/file-logger/file-logger.TestCallFuncName
// github.com/ha-ya4/my-package/file-logger.TestCallFuncNameの形にする
func getPath(fullPath string) string {
	pathArray := strings.Split(fullPath, "/")
	var flag bool
	var path string

	for _, s := range pathArray {
		if s == "github.com" {
			flag = true
		}

		if flag {
			path += s
			path += "/"
		}
	}

	return strings.TrimSuffix(path, "/")
}

func callFunc() (int, string) {
	funcName := callFuncName()
	return findCallLineAndFile(funcName)
}

func cps() string {
	funcName := callFuncName()
	return createCallPlaceSTR(funcName)
}
