package filelogger

type logLevel string
type logMode string

// logLevel?
const (
	DEBUG logLevel = "DEBUG"
	INFO  logLevel = "INFO"
	WARN  logLevel = "WARN"
	ERROR logLevel = "ERROR"

	ModeDebug      logMode = "DebugMode"
	ModeProduction logMode = "ProductionMode"
)

func (lm logMode) isNoOutput(level logLevel) bool {
	if lm == ModeDebug {
		return lm.noOutputDebug(level)
	}

	if lm == ModeProduction {
		return lm.noOutputProd(level)
	}

	return false
}

func (lm logMode) noOutputDebug(level logLevel) bool {
	return false
}

func (lm logMode) noOutputProd(level logLevel) bool {
	no := false

	if level != ERROR {
		no = true
	}

	return no
}
