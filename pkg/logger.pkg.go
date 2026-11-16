package sdk_pkg

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

/**
#===============================================
# GCP CLOUD LOGGING BOOTSTRAP
#===============================================
# InitLogger owns the global logrus formatter. It emits one JSON object per
# line, which lets the GCP Ops Agent turn each line into a real LogEntry
# instead of an opaque string, and attaches the reserved keys Cloud Logging
# promotes out of the payload:
#
#   logging.googleapis.com/labels   -> LogEntry.labels
#   logging.googleapis.com/severity -> LogEntry.severity
#
# Resulting Log Explorer queries:
#
#   labels.service="worker-external-beyond-disburse"
#   labels.host="worker-vm-02"
#   labels.service="worker-x" AND labels.host="worker-vm-02"
*/

const (
	// LogLabelsKey is promoted to LogEntry.labels by the Ops Agent.
	LogLabelsKey = "logging.googleapis.com/labels"

	// LogSeverityKey is promoted to LogEntry.severity by the Ops Agent.
	LogSeverityKey = "logging.googleapis.com/severity"

	// EnvLogService overrides the service label; defaults to the binary name.
	EnvLogService = "LOG_SERVICE_NAME"

	// EnvLogInstance identifies one replica of a service, such as the %i of a
	// systemd template unit.
	EnvLogInstance = "LOG_SERVICE_INSTANCE"

	// EnvLogHost overrides the host label; defaults to os.Hostname.
	EnvLogHost = "LOG_SERVICE_HOST"

	// EnvLogZone optionally records the GCP zone or region of the VM.
	EnvLogZone = "LOG_SERVICE_ZONE"

	// EnvLogLevel sets the logrus level; defaults to info.
	EnvLogLevel = "LOG_LEVEL"

	// EnvLogFormat selects "json" (default) or "text" for local development.
	EnvLogFormat = "LOG_FORMAT"

	// LogFormatText disables JSON output and the reserved GCP keys.
	LogFormatText = "text"

	// FallbackServiceName is used when no binary name can be determined.
	FallbackServiceName = "go-service"
)

var (
	// loggerOnce guards the explicit bootstrap.
	loggerOnce sync.Once

	// loggerInitialised reports whether InitLogger has run, so that the lazy
	// fallback in Logrus never fights an explicit bootstrap.
	loggerInitialised bool

	// loggerMu guards loggerInitialised and loggerLabels.
	loggerMu sync.RWMutex

	// loggerLabels is immutable after InitLogger.
	loggerLabels map[string]string
)

// labelHook attaches the shared label set and the Cloud Logging severity to
// every entry the global logger emits.
type labelHook struct {
	labels map[string]string

	// gcp is false under LOG_FORMAT=text, where the reserved keys are noise
	// rather than metadata.
	gcp bool
}

// Levels reports that the hook decorates every emitted entry.
func (labelHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire injects the shared labels and mirrors the logrus level into the
// reserved severity key. Without it Cloud Logging derives severity from the
// journald PRIORITY of the whole unit, which makes every line look like INFO.
func (h labelHook) Fire(entry *logrus.Entry) error {
	if !h.gcp {
		return nil
	}

	if entry.Data == nil {
		entry.Data = logrus.Fields{}
	}

	if _, exists := entry.Data[LogLabelsKey]; !exists {
		entry.Data[LogLabelsKey] = h.labels
	}

	entry.Data[LogSeverityKey] = GcpSeverity(entry.Level)

	return nil
}

// InitLogger configures the global logrus logger exactly once per process and
// takes ownership of the formatter away from the lazy fallback used by Logrus.
// Call it as the first statement of main, before any other log site runs.
//
// Pass nil to derive every label from the environment. A service name is read
// from LOG_SERVICE_NAME, then from opts.Service, then from the binary name.
func InitLogger(opts *sdk_dto.LoggerOptions) {
	loggerOnce.Do(func() {
		if opts == nil {
			opts = &sdk_dto.LoggerOptions{}
		}

		labels := map[string]string{
			"service": resolveServiceName(opts.Service),
			"env":     resolveEnvironment(opts.Environment),
		}

		if host := resolveHost(opts.Host); host != sdk_cons.EMPTY {
			labels["host"] = host
		}

		if instance := firstNonEmpty(opts.Instance, os.Getenv(EnvLogInstance)); instance != sdk_cons.EMPTY {
			labels["instance"] = instance
		}

		if zone := firstNonEmpty(opts.Zone, os.Getenv(EnvLogZone)); zone != sdk_cons.EMPTY {
			labels["zone"] = zone
		}

		for key, value := range opts.Labels {
			if key = strings.TrimSpace(key); key != sdk_cons.EMPTY {
				labels[key] = value
			}
		}

		formatter, gcp := resolveFormatter(opts.Format)

		logrus.SetOutput(os.Stdout)
		logrus.SetFormatter(formatter)
		logrus.SetLevel(resolveLevel(opts.Level))
		logrus.AddHook(labelHook{labels: labels, gcp: gcp})

		loggerMu.Lock()
		loggerLabels = labels
		loggerInitialised = true
		loggerMu.Unlock()
	})
}

// LoggerInitialised reports whether InitLogger has taken ownership of the
// global logger.
func LoggerInitialised() bool {
	loggerMu.RLock()
	defer loggerMu.RUnlock()

	return loggerInitialised
}

// LogServiceName returns the resolved service label, falling back to the
// binary name when InitLogger has not run yet.
func LogServiceName() string {
	loggerMu.RLock()
	labels := loggerLabels
	loggerMu.RUnlock()

	if labels != nil {
		return labels["service"]
	}

	return resolveServiceName(sdk_cons.EMPTY)
}

// LogServiceLabels returns a copy of the shared label set for callers that
// build their own log payloads.
func LogServiceLabels() map[string]string {
	loggerMu.RLock()
	labels := loggerLabels
	loggerMu.RUnlock()

	out := make(map[string]string, len(labels))
	for key, value := range labels {
		out[key] = value
	}

	return out
}

// GcpSeverity maps a logrus level onto a Cloud Logging LogSeverity string.
// Cloud Logging only accepts these upper-case names; the lower-case logrus
// level names are rejected and fall back to DEFAULT.
func GcpSeverity(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel, logrus.DebugLevel:
		return "DEBUG"
	case logrus.InfoLevel:
		return "INFO"
	case logrus.WarnLevel:
		return "WARNING"
	case logrus.ErrorLevel:
		return "ERROR"
	case logrus.FatalLevel:
		return "CRITICAL"
	case logrus.PanicLevel:
		return "ALERT"
	default:
		return "DEFAULT"
	}
}

// resolveServiceName prefers LOG_SERVICE_NAME, then the explicit option, then
// the binary name. Deployments that keep the systemd SyslogIdentifier equal to
// the binary name therefore need no per-unit configuration.
func resolveServiceName(service string) string {
	if name := strings.TrimSpace(os.Getenv(EnvLogService)); name != sdk_cons.EMPTY {
		return name
	}

	if service = strings.TrimSpace(service); service != sdk_cons.EMPTY {
		return service
	}

	if len(os.Args) > 0 {
		base := filepath.Base(os.Args[0])
		if base != sdk_cons.EMPTY && base != "." && base != string(filepath.Separator) {
			return base
		}
	}

	return FallbackServiceName
}

// resolveHost prefers LOG_SERVICE_HOST, then the explicit option, then the
// kernel hostname, which on GCE is the instance name.
func resolveHost(host string) string {
	if value := strings.TrimSpace(os.Getenv(EnvLogHost)); value != sdk_cons.EMPTY {
		return value
	}

	if host = strings.TrimSpace(host); host != sdk_cons.EMPTY {
		return host
	}

	if name, err := os.Hostname(); err == nil {
		return strings.TrimSpace(name)
	}

	return sdk_cons.EMPTY
}

// resolveEnvironment reports the deployment stage used as the "env" label.
func resolveEnvironment(environment string) string {
	if environment = strings.TrimSpace(environment); environment != sdk_cons.EMPTY {
		return environment
	}

	for _, key := range []string{"APP_ENV", "ENV", "GIN_MODE"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != sdk_cons.EMPTY {
			return value
		}
	}

	return "unknown"
}

// resolveLevel parses the explicit option, then LOG_LEVEL, defaulting to info
// on an empty or invalid value.
func resolveLevel(level string) logrus.Level {
	raw := firstNonEmpty(level, os.Getenv(EnvLogLevel))
	if raw == sdk_cons.EMPTY {
		return logrus.InfoLevel
	}

	parsed, err := logrus.ParseLevel(raw)
	if err != nil {
		return logrus.InfoLevel
	}

	return parsed
}

// resolveFormatter returns the formatter to install and whether the reserved
// GCP keys should be emitted alongside it.
func resolveFormatter(format string) (logrus.Formatter, bool) {
	if strings.EqualFold(firstNonEmpty(format, os.Getenv(EnvLogFormat)), LogFormatText) {
		return &logrus.TextFormatter{
			ForceColors:     true,
			TimestampFormat: time.RFC3339,
		}, false
	}

	return &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "timestamp",
			logrus.FieldKeyMsg:  "message",
			// The Cloud Logging severity lives under LogSeverityKey; "level"
			// stays a plain payload field so raw journals remain greppable.
			logrus.FieldKeyLevel: "level",
		},
	}, true
}

// firstNonEmpty returns the first argument that is not blank.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != sdk_cons.EMPTY {
			return value
		}
	}

	return sdk_cons.EMPTY
}
