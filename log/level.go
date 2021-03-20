package log

// Level represents the log level of the logger
type Level uint

// Supported log levels
const (
	DebugLevel Level = 1
	InfoLevel  Level = 2
	ErrorLevel Level = 3
)

// MarshalJSON returns a json representation of the Level
func (level Level) String() string {
	values := []string{"debug", "info", "error"}
	if int(level) > len(values)-1 {
		return ""
	}

	return values[level]
}
