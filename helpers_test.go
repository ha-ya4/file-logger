package filelogger

import (
	"os"
	"strings"
	"testing"
	//"fmt"
)

func TestCallFuncName(t *testing.T) {
	funcName := callFuncName()

	currentPath, _ := os.Getwd()
	expected := getPath(currentPath) + ".TestCallFuncName"

	if funcName != expected {
		t.Errorf("\n関数名が一致していません\nfuncName=%s\nexpected=%s", funcName, expected)
	}
}

func TestFindCallLineAndFile(t *testing.T) {
	line, file := callFunc()

	expectedLine := 22
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

func TestShortFileName(t *testing.T) {
	_, file := callFunc()
	name, _ := shortFileName(file)
	expectedName := "helpers_test.go"

	if name != expectedName {
		t.Errorf("\nショートファイル名が一致していません\nfile=%s\nexpected=%s", name, expectedName)
	}
}

func TestCreateCallPlaceSTR(t *testing.T) {
	cps := cps()
	expectedCPS := "helpers_test.go:48:"
	t.Log(cps)

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
