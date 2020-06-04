package filelogger

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineCounter(t *testing.T) {
	file, _ := os.Open("./linecounter_test.txt")
	defer file.Close()

	count, _ := lineCounter(file)
	expectedCount := 18
	assert.True(t, expectedCount == count)
}

func TestContainsSTRFileList(t *testing.T) {
	//test用ディレクトリとファイル作成
	dirName := "ttttest"
	err := os.Mkdir(dirName, 0777)
	assert.NoError(t, err)
	fName := []string{"a", "b", "c", "d", "e", "f"}
	for _, n := range fName {
		fn := filepath.Join(dirName, n+".txt")
		_, err = os.Create(fn)
		assert.NoError(t, err)
	}
	// 一つだけ拡張子が違うファイルを作り、containsSTRFileListで取得できる配列にふくまれていないかをチェックする
	fn := filepath.Join(dirName, "hello.js")
	_, err = os.Create(fn)
	assert.NoError(t, err)

	fileList := containsSTRFileList(dirName, ".txt")
	for i := 0; i < len(fileList); i++ {
		equal := fName[i]+".txt" == fileList[i].Name()
		assert.True(t, equal)
	}

	assert.NoError(t, os.RemoveAll(dirName))
}

func TestCompressAndUnfreeze(t *testing.T) {
	var err error
	target := []byte("Hello World!")

	byt := &bytes.Buffer{}
	err = compress(byt, target)
	assert.NoError(t, err)
	assert.False(t, byt.String() == string(target))

	byt, err = Unfreeze(byt)
	assert.NoError(t, err)
	assert.True(t, byt.String() == string(target))
}
