package local

import (
	"os"

	"github.com/rilldata/rill/cli/pkg/dotrill"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogFormat string

// Default log formats for logger
const (
	LogFormatConsole = "console"
	LogFormatJSON    = "json"
)

func ParseLogFormat(format string) (LogFormat, bool) {
	switch format {
	case "json":
		return LogFormatJSON, true
	case "console":
		return LogFormatConsole, true
	default:
		return "", false
	}
}

func initLogger(isVerbose bool, logFormat LogFormat) (logger *zap.Logger, cleanupFn func()) {
	logLevel := zapcore.InfoLevel
	if isVerbose {
		logLevel = zapcore.DebugLevel
	}

	logPath, err := dotrill.ResolveFilename("rill.log", true)
	if err != nil {
		panic(err)
	}
	// lumberjack.Logger is already safe for concurrent use, so we don't need to
	// lock it.
	luLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     30, // days
		Compress:   true,
	}
	cfg := zap.NewProductionEncoderConfig()
	// hide logger name like `console`
	cfg.NameKey = zapcore.OmitKey
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileCore := zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(luLogger), logLevel)

	var consoleEncoder zapcore.Encoder
	opts := make([]zap.Option, 0)
	switch logFormat {
	case LogFormatJSON:
		cfg := zap.NewProductionEncoderConfig()
		cfg.NameKey = zapcore.OmitKey
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
		// never
		opts = append(opts, zap.AddStacktrace(zapcore.InvalidLevel))
		consoleEncoder = zapcore.NewJSONEncoder(cfg)
	case LogFormatConsole:
		encCfg := zap.NewDevelopmentEncoderConfig()
		encCfg.NameKey = zapcore.OmitKey
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000")
		consoleEncoder = zapcore.NewConsoleEncoder(encCfg)
	}

	// If it's not verbose, skip the instance_id field.
	if !isVerbose {
		consoleEncoder = skipFieldZapEncoder{
			Encoder: consoleEncoder,
			fields:  []string{"instance_id"},
		}
	}

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), logLevel)

	// For the local console output, apply some filters that are usually not informative for the user.
	if !isVerbose {
		consoleCore = &filterZapCore{
			next:      consoleCore,
			predicate: consoleFilterPredicate,
		}
	}

	core := zapcore.NewTee(
		fileCore,
		consoleCore,
	)

	return zap.New(core, opts...), func() {
		_ = logger.Sync()
		luLogger.Close()
	}
}

// consoleFilterPredicate filters out log that we don't want to show in the local console output (not applied if --verbose is passed).
func consoleFilterPredicate(entry zapcore.Entry, fields []zapcore.Field) bool {
	// Always log warn and error logs
	if entry.Level > zapcore.InfoLevel {
		return true
	}

	switch entry.Message {
	// Filter out reconciling logs for internal resources
	case "Reconciling resource", "Reconciled resource":
		for _, field := range fields {
			if field.Key != "type" {
				continue
			}

			switch field.String {
			case "ProjectParser", "RefreshTrigger":
				return false
			}

			break
		}
	// Filter out model executing log (usually the "reconciling resource" log is enough)
	case "Executing model":
		return false
	}

	return true
}

// skipFieldZapEncoder skips fields with the given keys. only string fields are supported.
type skipFieldZapEncoder struct {
	zapcore.Encoder
	fields []string
}

func (s skipFieldZapEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	res := make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		skip := false
		for _, skipField := range s.fields {
			if field.Key == skipField {
				skip = true
				break
			}
		}
		if !skip {
			res = append(res, field)
		}
	}
	return s.Encoder.EncodeEntry(entry, res)
}

func (s skipFieldZapEncoder) Clone() zapcore.Encoder {
	return skipFieldZapEncoder{
		Encoder: s.Encoder.Clone(),
		fields:  s.fields,
	}
}

func (s skipFieldZapEncoder) AddString(key, val string) {
	skip := false
	for _, skipField := range s.fields {
		if key == skipField {
			skip = true
			break
		}
	}
	if !skip {
		s.Encoder.AddString(key, val)
	}
}

// filterZapCore filters out logs that the predicate function returns false for.
type filterZapCore struct {
	next      zapcore.Core
	predicate func(zapcore.Entry, []zapcore.Field) bool
}

func (c *filterZapCore) Enabled(level zapcore.Level) bool {
	return c.next.Enabled(level)
}

func (c *filterZapCore) With(fields []zapcore.Field) zapcore.Core {
	return &filterZapCore{
		next:      c.next.With(fields),
		predicate: c.predicate,
	}
}

func (c *filterZapCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.next.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func (c *filterZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if c.predicate(entry, fields) {
		return c.next.Write(entry, fields)
	}
	return nil
}

func (c *filterZapCore) Sync() error {
	return c.next.Sync()
}
