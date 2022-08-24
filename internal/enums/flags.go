package enums

import (
	"github.com/thediveo/enumflag/v2"
)

// LogLevelMode ① Define new enum flag type for log level.
type LogLevelMode enumflag.Flag

// ② Define the enumeration values for LogLevelMode.
const (
	Debug LogLevelMode = iota
	Trace
	Info
	Warn
	Error
	Fatal
	Panic
)

// LogLevelModeIds ③ Map enumeration values to their textual representations (value
// identifiers).
var LogLevelModeIds = map[LogLevelMode][]string{
	Trace: {"trace"},
	Debug: {"debug"},
	Info:  {"info"},
	Warn:  {"warn"},
	Error: {"error"},
	Fatal: {"fatal"},
	Panic: {"panic"},
}
