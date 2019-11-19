package filelogger

import (
	"fmt"
	"testing"
	"path/filepath"
)

// 実際の動作テスト用
func TestLogger(t *testing.T) {
	Logger.SetFilePath("./test.log")
	//Logger.SetMaxLine(2)
	//Logger.SetMaxRotation(1)
	dir := filepath.Dir(Logger.filePath)
	name := filepath.Base(Logger.filePath)
	fileList := containsSTRFileList(dir, name)
	for _, s := range fileList {
		fmt.Println(s.Name())
	}
	Logger.Println(ERROR, "a")
}
