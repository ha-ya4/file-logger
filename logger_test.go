package filelogger

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dirPath     = "./logtest"
	fileName    = "test.log"
	filePath    = filepath.Join(dirPath, fileName)
	msg         = "test err"
	maxLine     = 100
	maxRotation = 5
	testCount   = 10000 // 回数が少ないとテスト失敗の可能性あり
)

// 指定した回数並行処理でログを出力する
func forPrintln(c int) {
	wg := &sync.WaitGroup{}
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func() {
			Rprintln(ERROR, msg)
			wg.Done()
		}()
	}
	wg.Wait()
}

// gzipを解凍する。調査用。
func u(n string) (*bytes.Buffer, error) {
	fn := filepath.Join(dirPath, n)
	f, _ := os.Open(fn)
	b, err := Unfreeze(f)
	return b, err
}

//******************************************************
// ここからテスト
//******************************************************

// テスト用ディレクトリを作成しテスト終了後に削除する。ローテーションの設定もここで行う
// testCountが大きすぎるとgoroutineがおかしく？なってpanicしてしまうのでifでチェックして必要ならpanicする
// (おそらくgoroutineの立ち上げすぎが原因だと思う)
// 特定の値以下だとテストに失敗するのでifでチェックし必要ならpanicする
func TestMain(m *testing.M) {
	if testCount > 100000 {
		panic("logger_test.go: testCountは100000までの数値にしてください")
	}

	c := maxLine*maxRotation - maxLine + 1
	if testCount < c {
		msg := fmt.Sprintf("logger_test.go: testCountは%vより大きい数値にしてください", c)
		panic(msg)
	}

	conf := &Config{
		Rotate:      RotateConfig{MaxLine: maxLine, MaxRotation: maxRotation},
		Mode:        ModeProduction,
		LoggerFlags: LoggerFlags,
		FilePath:    filePath,
		FilePerm:    0666,
		FileFlags:   FileFlags,
		Compress:    true,
		LogLevelConf: LogLevelConfig{
			LevelConfig{
				Mode:          ModeProduction,
				ExcludedLevel: []string{DEBUG},
			},
		},
	}
	Initialize(conf)

	os.Mkdir(dirPath, 0777)
	SetFilePath(filePath)
	SetRotate(RotateConfig{MaxLine: maxLine, MaxRotation: maxRotation})
	forPrintln(testCount)

	code := m.Run()

	os.RemoveAll(dirPath)

	os.Exit(code)
}

// 圧縮されていないことが期待されるファイルを解凍し、nil pointer dereferenceが起こるか確認する
// 解凍に成功したり、nil pointer dereference以外のエラーならテスト失敗となる
func TestNoCompress(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			if e, ok := err.(error); ok {
				assert.True(t, e.Error() == "runtime error: invalid memory address or nil pointer dereference")
			}
		}
	}()
	if _, e := u(fileName); e == nil {
		t.Errorf("解凍に成功しました：　%v", e)
	}
}

// ファイルにログが出力されているかのテスト
func TestOutput(t *testing.T) {
	f, err := os.Open(filePath)
	assert.NoError(t, err)
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		assert.True(t, strings.Contains(s.Text(), "[ERROR] test err"))
		break
	}
}

// 最新のファイル以外が圧縮されているか。Unfreezeがエラーを返さなければ圧縮されているとみなす
func TestCompress(t *testing.T) {
	fi, err := ioutil.ReadDir(dirPath)
	assert.NoError(t, err)
	for _, f := range fi {
		// 圧縮されていないファイルを飛ばす
		if f.Name() == fileName {
			continue
		}

		path := filepath.Join(dirPath, f.Name())
		f, err := os.Open(path)
		assert.NoError(t, err)
		defer f.Close()
		_, err = Unfreeze(f)
		assert.NoError(t, err)
	}
}

// 指定した最大行数で次のファイルに移行しているか
func TestMaxLine(t *testing.T) {
	fi, err := ioutil.ReadDir(dirPath)
	assert.NoError(t, err)

	ofn := oldFileName(fi)
	path := filepath.Join(dirPath, ofn)
	f, err := os.Open(path)
	defer f.Close()
	assert.NoError(t, err)

	b, err := Unfreeze(f)
	assert.NoError(t, err)

	result, err := lineCounter(b)
	assert.NoError(t, err)
	expected := maxLine + 1 // 改行が入るので指定した値に+1する
	assert.True(t, expected == result)
}

// 指定した最大ファイル数でローテーションしているか
func TestRotation(t *testing.T) {
	fi, err := ioutil.ReadDir(dirPath)
	assert.NoError(t, err)
	result := len(fi)
	expected := maxRotation
	assert.True(t, expected == result)
}

func TestLogLevelConfigFindMode(t *testing.T) {
	lc := LogLevelConfig{
		LevelConfig{
			Mode: "go",
		},
		LevelConfig{
			Mode: "js",
		},
		LevelConfig{
			Mode: "ts",
		},
		LevelConfig{
			Mode: "py",
		},
		LevelConfig{
			Mode: "rust",
		},
	}
	i, exist := lc.findMode("py")
	assert.True(t, exist)
	assert.True(t, i == 3)

	i, exist = lc.findMode("php")
	assert.False(t, exist)
}

func TestLevelConfigIsExcluded(t *testing.T) {
	lc := LevelConfig{
		Mode: "go",
		ExcludedLevel: []string{
			"INFO", "DEBUG",
		},
	}
	assert.True(t, lc.isExcluded("INFO"))
	assert.False(t, lc.isExcluded("ERROR"))
}

func TestShouldNotOutput(t *testing.T) {
	lc := LogLevelConfig{
		LevelConfig{
			Mode: "go",
			ExcludedLevel: []string{
				"INFO", "DEBUG",
			},
		},
		LevelConfig{
			Mode: "js",
			ExcludedLevel: []string{
				"INFO",
			},
		},
		LevelConfig{
			Mode: "ts",
			ExcludedLevel: []string{
				"DEBUG",
			},
		},
	}
	logger := fileLogger{
		Conf: &Config{
			Mode:         "go",
			LogLevelConf: lc,
		},
	}

	assert.True(t, logger.shouldNotOutput("INFO"))
	assert.False(t, logger.shouldNotOutput("ERROR"))

	logger.Conf.Mode = "ts"
	assert.True(t, logger.shouldNotOutput("DEBUG"))
	assert.False(t, logger.shouldNotOutput("INFO"))
}
