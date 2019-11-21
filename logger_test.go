package filelogger

import (
	//"fmt"
	"testing"
	//"path/filepath"
	//"time"
)

// 実際の動作テスト用
func TestLogger(t *testing.T) {
	Logger.SetFilePath("./test.log")
	Logger.SetMaxLine(2)
	Logger.SetMaxRotation(3)
	Logger.Println(ERROR, "a")
}
