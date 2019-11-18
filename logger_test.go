package filelogger

import (
	//"fmt"
	"testing"
)

// 実際の動作テスト用
func TestLogger(t *testing.T) {
	Logger.SetFilePath("./test.log")
	Logger.SetMaxLine(10)
	//Logger.Println(ERROR, "test")
}
