package filelogger

import (
	"bytes"
	"compress/gzip"
	//"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type fileNameManager struct {
	path string
	name string
	dir  string
}

func newFileNameManager(path string) *fileNameManager {
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	return &fileNameManager{
		path: path,
		name: name,
		dir:  dir,
	}
}

// ファイル名と行数を main.go:145: のような形に結合する
// shortFileNameはos.Getwdのエラーを返すが、ここでエラーがでても（おそらく）短くできなかった元のカレントディレクトリ名が返ってくると思うのでerrorは無視する
// その場合は長いままのディレクトリ名で出力する
func createCallPlaceSTR(callFuncName string) string {
	line, file := findCallLineAndFile(callFuncName)
	name := filepath.Base(file)
	return name + ":" + strconv.Itoa(line) + ":"
}

// iをインクリメントしていってスタックトレースを遡り,FileLoggerのprint系メソッドが呼び出された次のトレースのラインとファイル名を返す
func findCallLineAndFile(callFuncName string) (int, string) {
	var (
		pc        uintptr
		file      string
		line      int
		ok        bool
		breakFlag bool
		i         int
	)

	// 呼び出した関数の次の情報がほしいので、ブレークするべきかは先に確認し、最後にフラグの操作をする
	for {
		pc, file, line, ok = runtime.Caller(i)
		funcName := runtime.FuncForPC(pc).Name()

		if breakFlag {
			break
		}

		if !ok {
			break
		}

		breakFlag = strings.Contains(funcName, callFuncName)
		i++
	}

	return line, file
}

func callFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

const bufSize = 8 * 1024

// ファイルの行数を取得する
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, bufSize)
	delimiter := []byte{'\n'}
	count := 0

	// ファイル内の\nの数をカウントする。実際の行数より１少なくなってしまっているのでEOFを確認したあとに+1する
	for {
		b, err := r.Read(buf)
		count += bytes.Count(buf[:b], delimiter)

		switch {
		case err == io.EOF:
			count++
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

const timeFormat = "Jan 2 15:04:05 2006"

func createFileName(filePath string) string {
	now := time.Now().Format(timeFormat)
	fileName := filepath.Base(filePath)
	return now + "_" + fileName
}

func containsSTRFileList(path, str string) []os.FileInfo {
	files, _ := ioutil.ReadDir(path)
	var fileList []os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.Contains(file.Name(), str) {
			fileList = append(fileList, file)
		}
	}
	return fileList
}

func oldFileName(fileList []os.FileInfo) string {
	var (
		varTime time.Time
		t       time.Time
		name    string
	)

	for i, fi := range fileList {
		var err error
		timeSTR := strings.Split(fi.Name(), "_")[0]
		t, err = time.Parse(timeFormat, timeSTR)
		if err != nil {
			continue
		}

		if i == 0 {
			varTime = t
			name = fi.Name()
			continue
		}

		if t.Before(varTime) {
			varTime = t
			name = fi.Name()
		}
	}

	return name
}

func compress(w io.Writer, content []byte) error {
	writer := gzip.NewWriter(w)
	_, err := writer.Write(content)
	writer.Close()

	return err
}

// Unfreeze gzipで圧縮されたものを解答する
func Unfreeze(r io.Reader) (bytes.Buffer, error) {
	var err error
	reader, err := gzip.NewReader(r)
	b := bytes.Buffer{}
	b.ReadFrom(reader)
	return b, err
}

