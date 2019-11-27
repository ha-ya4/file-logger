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
	Logger.SetRotate(RotateConfig{maxLine: 2, maxRotation: 3})
	Logger.Rprintln(ERROR, "a")
}
