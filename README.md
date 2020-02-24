# file-logger


```
import (
  "github.com/ha-ya4/file-logger"
)

func main() {
  filelogger.SetFilePath("./test.log")
  filelogger.Rprintln(filelogger.ERROR, "test")
}
```

```
// test.log
2019/12/02 23:25:15 main.go:9:[ERROR] test

```
ローテーションする場合

```
import (
  "github.com/ha-ya4/file-logger"
)

func main() {
  filelogger.SetFilePath("./test.log")
  filelogger.SetRotate(
    filelogger.RotateConfig{MaxLine: 1000, MaxRotation: 5},
  )
  filelogger.Rprintln(filelogger.ERROR, "test")

  // ログ出力...
}
```

```
Dec 2 23:41:09.954114180 2019_test.log
Dec 2 23:41:09.957540479 2019_test.log
Dec 2 23:41:09.960394743 2019_test.log
Dec 2 23:41:09.963279050 2019_test.log
test.log
```

```
// test.log
1 2019/12/02 23:25:15 main.go:9:[ERROR] test
...
...
...
...

1000 2019/12/02 23:26:14 main.go:9:[ERROR] test

```

モード
```
filelogger.SetMode(ModeProduction)
filelogger.Rprintln(filelogger.ERROR, "test error")
filelogger.Rprintln(filelogger.DEBUG, "test debug")
```
```
// test.log
1 2019/12/02 23:25:15 main.go:9:[ERROR] test error
```

|         |DEBUG|INFO |WARN |ERROR|
|---      |:---:  |:---:  |:---:  |:---:  |
|debug    |  ○  |  ○  |  ○  |  ○  |
|roduction|     |     |     |  ○  |
