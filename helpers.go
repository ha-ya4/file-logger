package filelogger

import (
	"bytes"
	//"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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

func createFileName(filePath string) string {
	now := getNow()
	fileName := filepath.Base(filePath)
	return now + fileName
}

// 2019-11-18|13:20:18|というような形で現在時刻を返す
func getNow() string {
	n := time.Now().String()
	now := strings.Split(n, ".")[0]
	now = strings.Replace(now, " ", "|", 1)
	return now + "|"
}

func getFileList(path string) []string {
	files, _ := ioutil.ReadDir(path)
	var fileList []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileList = append(fileList, file.Name())
	}
	return fileList
}

func containsLen(list []string, str string) int {
	var count int
	for _, s := range list {
		if strings.Contains(s, str) {
			count++
		}
	}
	return count
}
