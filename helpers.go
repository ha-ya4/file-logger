package filelogger

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	//"fmt"
)

func logPrintln(msg string) {
	prefix := "[filelogger error] "
	l := log.New(os.Stdout, prefix, log.Ldate|log.Ltime|log.LstdFlags)
	l.Println(msg)
}

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

const timeFormat = "Jan 2 15:04:05.000000000 2006"

// getNameAddTimeNow ファイル名の先頭に現在の日時を付与した名前を返す
func (f *fileNameManager) getNameAddTimeNow() string {
	now := time.Now().Format(timeFormat)
	return now + "_" + f.name
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

// containsSTRFileList 指定したディレクトリにある、指定した文字列が含まれるファイル名のos.FileInfo配列を返す
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

// oldFileName 受け取った配列の中の日時が付与されたファイル名で一番古いファイルの名前を返す
func oldFileName(fileList []os.FileInfo) string {
	var (
		varTime time.Time
		t       time.Time
		name    string
	)
	flag := true

	// 一回目のループの日時をvarTime変数にセットし、次のループの日時tと比較する
	// tのほうが古い場合varTimeにセットする、を繰り返す
	for _, fi := range fileList {
		var err error
		timeSTR := strings.Split(fi.Name(), "_")[0]
		t, err = time.Parse(timeFormat, timeSTR)
		if err != nil {
			continue
		}

		if flag {
			varTime = t
			name = fi.Name()
			flag = false
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

// CompressFile 指定したファイルをgzip形式で圧縮する
func CompressFile(path string) error {
	var err error

	file, err := os.Open(path)
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	newFile, err := os.Create(path)
	if err != nil {
		return err
	}

	err = compress(newFile, b)
	return err
}

// Unfreeze gzipで圧縮されたものを解凍する。このパッケージには直接かかわらないが、補助用の関数として書いておく
func Unfreeze(r io.Reader) (*bytes.Buffer, error) {
	var err error
	reader, err := gzip.NewReader(r)
	b := &bytes.Buffer{}
	b.ReadFrom(reader)
	reader.Close()
	return b, err
}
