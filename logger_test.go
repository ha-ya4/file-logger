package filelogger

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	dirPath     = "./logtest"
	fileName    = "/test.log"
	filePath    = filepath.Join(dirPath, fileName)
	msg         = "test err"
	maxLine     = 100
	maxRotation = 5
	testCount   = 10000 // 回数が少ないとテスト失敗の可能性あり
)

// テスト用ディレクトリを作成しテスト終了後に削除する。ローテーションの設定もここで行う
func TestMain(m *testing.M) {
	os.Mkdir(dirPath, 0777)
	Logger.SetFilePath(filePath)
	Logger.SetRotate(RotateConfig{maxLine: maxLine, maxRotation: maxRotation})
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
				if e.Error() != "runtime error: invalid memory address or nil pointer dereference" {
					t.Errorf("予期せぬエラーです：%s", err)
				}
			}
		}
	}()
	_, e := u(fileName)

	if e == nil {
		t.Errorf("解凍に成功しました：　%v", e)
	}
}

// ファイルにログが出力されているかのテスト
func TestOutput(t *testing.T) {
	f, err := os.Open(filePath)
	tfatal(t, err)
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		ex1 := strings.Contains(s.Text(), "[ERROR] test err")
		ex2 := strings.Contains(s.Text(), "logger_test.go")
		if !ex1 && !ex2 {
			t.Errorf("期待される出力が得られませんでした\n t=%s", s.Text())
		}
		break
	}
}

// 最新のファイル以外が圧縮されているか
func TestCompress(t *testing.T) {}

// 指定した最大行数で次のファイルに移行しているか
func TestMaxLine(t *testing.T) {}

// 指定した最大ファイル数でローテーションしているか
func TestRotation(t *testing.T) {
	fi, err := ioutil.ReadDir(dirPath)
	tfatal(t, err)
	result := len(fi)
	expected := maxRotation
	if result != expected {
		t.Errorf("設定されたファイル数でローテーションされていません\n r=%v e=%v", result, expected)
	}
}

// 指定した回数並行処理でログを出力する
func forPrintln(c int) {
	wg := &sync.WaitGroup{}
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func() {
			Logger.Rprintln(ERROR, msg)
			wg.Done()
		}()
	}
	wg.Wait()
}

// gzipを解凍する。調査用。
func u(n string) (bytes.Buffer, error) {
	fn := filepath.Join(dirPath, n)
	f, _ := os.Open(fn)
	b, err := Unfreeze(f)
	return b, err
}

func tfatal(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err.Error())
	}
}
