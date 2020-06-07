# file-logger


```
import (
  "github.com/ha-ya4/file-logger"
)

func main() {
  conf = &fileLogger.Config{
		LoggerFlags: log.Ldate | log.Ltime | log.LstdFlags,
		FilePath:    "test.log",
		FilePerm:    0666,
		FileFlags:   os.O_APPEND | os.O_CREATE | os.O_RDWR,
  }
  fileLogger.Initialize(conf)
  filelogger.Rprintln(filelogger.ERROR, "test")
}
```

```
// test.log
2019/12/02 23:25:15 [ERROR] test

```
### ローテーションする場合

```
import (
  "github.com/ha-ya4/file-logger"
)

func main() {
  conf = &fileLogger.Config{
    Rotate:      RotateConfig{MaxLine: 1000, MaxRotation: 5},
    LoggerFlags: log.Ldate | log.Ltime | log.LstdFlags,
    FilePath:    "test.log",
    FilePerm:    0666,
    FileFlags:   os.O_APPEND | os.O_CREATE | os.O_RDWR,
    Compress     true
  }
  fileLogger.Initialize(conf)
  filelogger.Rprintln(filelogger.ERROR, "test")

  // ログ出力...
}
```

```
// logファイル
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

### ログレベルによる出力の有無
```
import (
  "github.com/ha-ya4/file-logger"
)

func main() {
  conf = &fileLogger.Config{
    Mode:        "PROD"
    LoggerFlags: log.Ldate | log.Ltime | log.LstdFlags,
    FilePath:    "test.log",
    FilePerm:    0666,
    FileFlags:   os.O_APPEND | os.O_CREATE | os.O_RDWR,
    LogLevelConf: LogLevelConfig{
      LevelConfig{
        Mode:          "DEV",
        ExcludedLevel: []string{},
      },
      LevelConfig{
        Mode:          "PROD",
        ExcludedLevel: []string{"DEBUG", "INFO"},
      },
    },
  }
  fileLogger.Initialize(conf)


  filelogger.Rprintln(filelogger.ERROR, "test error")
  filelogger.Rprintln(filelogger.INFO, "test info")
  filelogger.Rprintln(filelogger.WARN, "test warn")

  // test.log
  // 2019/12/02 23:25:15 main.go:9:[ERROR] test error
  // 2019/12/02 23:25:16 main.go:9:[ERROR] test warn
}
```
