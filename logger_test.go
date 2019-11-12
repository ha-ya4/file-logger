package filelogger

import (
	"testing"
)

// 実際の動作テスト用
func TestLogger(t *testing.T) {
	fileLogger := NewfileLogger("./test.log")
	fileLogger.SetOutput()
	defer fileLogger.FileClose()
	fileLogger.SetPrefix("[test]")
	fileLogger.Println(INFO, "aaabbb")
	fileLogger.Println(WARN, "aaabbb")
	//fileLogger.Println(DEBUG, "aaabbb")
	//fileLogger.Println(ERROR, "aaabbb")
}