package filelogger

import (
	"bytes"
	"os"
	"testing"
)

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
	expectedLen := 7

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
