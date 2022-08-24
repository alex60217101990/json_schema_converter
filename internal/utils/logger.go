package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func ChangeLevel(level string, logger *zerolog.Logger) (err error) {
	var l zerolog.Level
	l, err = zerolog.ParseLevel(level)
	if err != nil {
		return err
	}
	if l == zerolog.TraceLevel {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	}

	*logger = logger.Level(l)
	return err
}

func InitLogger(useColor bool) (log zerolog.Logger) {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		if level, ok := i.(string); ok && useColor {
			switch level {
			case zerolog.LevelTraceValue:
				return color.HiRedString(strings.ToUpper(fmt.Sprintf("| %-6s|", level)))
			case zerolog.LevelDebugValue:
				return color.MagentaString(strings.ToUpper(fmt.Sprintf("| %-6s|", level)))
			case zerolog.LevelInfoValue:
				return color.BlueString(strings.ToUpper(fmt.Sprintf("| %-6s|", level)))
			case zerolog.LevelWarnValue:
				return color.YellowString(strings.ToUpper(fmt.Sprintf("| %-6s|", level)))
			case zerolog.LevelErrorValue:
				return color.RedString(strings.ToUpper(fmt.Sprintf("| %-6s|", level)))
			case zerolog.LevelFatalValue:
				return color.RedString(strings.ToUpper(fmt.Sprintf("| %-6s|", level)))
			case zerolog.LevelPanicValue:
				return color.RedString(strings.ToUpper(fmt.Sprintf("| %-6s|", level)))
			}
		}

		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}

	return zerolog.New(output).With().Timestamp().Caller().Logger()
}
