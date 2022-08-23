package types_convertation

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var _levelsMap = map[uint8]zerolog.Level{
	1: zerolog.DebugLevel,
	2: zerolog.InfoLevel,
	3: zerolog.WarnLevel,
	4: zerolog.ErrorLevel,
	5: zerolog.FatalLevel,
	6: zerolog.PanicLevel,
	7: zerolog.TraceLevel,
}

func ChangeLevel(level uint8, logger *zerolog.Logger) {
	if l, ok := _levelsMap[level]; ok {
		if l == zerolog.TraceLevel {
			zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		}

		*logger = logger.Level(l)
	}
}

func InitLogger() (log zerolog.Logger) {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}

	return zerolog.New(output).With().Timestamp().Caller().Logger()
}
