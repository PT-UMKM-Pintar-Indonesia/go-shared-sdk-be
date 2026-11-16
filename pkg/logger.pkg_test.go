package sdk_pkg

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
)

// TestLabelHookPromotesReservedKeys guards the contract the GCP Ops Agent
// relies on: the reserved label and severity keys must be present on every
// entry, so Cloud Logging populates LogEntry.labels and LogEntry.severity
// instead of burying them in jsonPayload.
func TestLabelHookPromotesReservedKeys(t *testing.T) {
	labels := map[string]string{
		"service": "worker-external-beyond-disburse",
		"host":    "worker-vm-02",
		"env":     "release",
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{logrus.FieldKeyMsg: "message"},
	})
	logger.AddHook(labelHook{labels: labels, gcp: true})

	entry := logger.WithField("component", "outbound_http")
	entry.Level = logrus.ErrorLevel
	entry.Message = "LOG_WORKER_OUTBOUND_ERROR"

	if err := logger.Hooks.Fire(entry.Level, entry); err != nil {
		t.Fatalf("hook: %v", err)
	}

	raw, err := logger.Formatter.Format(entry)
	if err != nil {
		t.Fatalf("format: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("emitted line is not JSON: %v\n%s", err, raw)
	}

	if got[LogSeverityKey] != "ERROR" {
		t.Errorf("%s = %v, want ERROR", LogSeverityKey, got[LogSeverityKey])
	}

	promoted, ok := got[LogLabelsKey].(map[string]any)
	if !ok {
		t.Fatalf("%s = %T, want object", LogLabelsKey, got[LogLabelsKey])
	}

	for key, want := range labels {
		if promoted[key] != want {
			t.Errorf("labels.%s = %v, want %v", key, promoted[key], want)
		}
	}

	if got["component"] != "outbound_http" {
		t.Errorf("caller field dropped: component = %v", got["component"])
	}
}

func TestLabelHookSkipsReservedKeysInTextMode(t *testing.T) {
	logger := logrus.New()
	logger.AddHook(labelHook{labels: map[string]string{"service": "x"}, gcp: false})

	entry := logger.WithField("component", "outbound_http")
	entry.Level = logrus.ErrorLevel

	if err := logger.Hooks.Fire(entry.Level, entry); err != nil {
		t.Fatalf("hook: %v", err)
	}

	if _, exists := entry.Data[LogLabelsKey]; exists {
		t.Errorf("%s must not be emitted when gcp is false", LogLabelsKey)
	}

	if _, exists := entry.Data[LogSeverityKey]; exists {
		t.Errorf("%s must not be emitted when gcp is false", LogSeverityKey)
	}
}

// TestLabelHookKeepsCallerSuppliedLabels lets one call site override the
// shared label set without the hook clobbering it.
func TestLabelHookKeepsCallerSuppliedLabels(t *testing.T) {
	logger := logrus.New()
	logger.AddHook(labelHook{labels: map[string]string{"service": "shared"}, gcp: true})

	entry := logger.WithField(LogLabelsKey, map[string]string{"service": "explicit"})
	entry.Level = logrus.InfoLevel

	if err := logger.Hooks.Fire(entry.Level, entry); err != nil {
		t.Fatalf("hook: %v", err)
	}

	got, ok := entry.Data[LogLabelsKey].(map[string]string)
	if !ok {
		t.Fatalf("%s = %T, want map[string]string", LogLabelsKey, entry.Data[LogLabelsKey])
	}

	if got["service"] != "explicit" {
		t.Errorf("labels.service = %q, want explicit", got["service"])
	}
}

// TestGcpSeverityUsesCloudLoggingNames pins the upper-case LogSeverity names.
// Cloud Logging rejects the lower-case logrus level names and would silently
// downgrade every entry to DEFAULT.
func TestGcpSeverityUsesCloudLoggingNames(t *testing.T) {
	for _, tc := range []struct {
		level logrus.Level
		want  string
	}{
		{logrus.TraceLevel, "DEBUG"},
		{logrus.DebugLevel, "DEBUG"},
		{logrus.InfoLevel, "INFO"},
		{logrus.WarnLevel, "WARNING"},
		{logrus.ErrorLevel, "ERROR"},
		{logrus.FatalLevel, "CRITICAL"},
		{logrus.PanicLevel, "ALERT"},
	} {
		if got := GcpSeverity(tc.level); got != tc.want {
			t.Errorf("GcpSeverity(%v) = %q, want %q", tc.level, got, tc.want)
		}
	}
}

// TestResolveServiceNameFallsBackToBinary pins the behaviour that lets systemd
// units stay untouched: the service label equals the binary name, which each
// unit keeps identical to its SyslogIdentifier.
func TestResolveServiceNameFallsBackToBinary(t *testing.T) {
	t.Setenv(EnvLogService, "")

	if got := resolveServiceName("worker-external-saebo-qris-static"); got != "worker-external-saebo-qris-static" {
		t.Errorf("explicit name = %q", got)
	}

	if got := resolveServiceName(""); got == "" || got == FallbackServiceName {
		t.Errorf("binary fallback = %q, want the test binary name", got)
	}

	t.Setenv(EnvLogService, "worker-internal-mobile-callback-store")

	if got := resolveServiceName("ignored"); got != "worker-internal-mobile-callback-store" {
		t.Errorf("%s override = %q", EnvLogService, got)
	}
}

// TestResolveHostIdentifiesTheVM is what makes one service running on two VMs
// distinguishable in Log Explorer.
func TestResolveHostIdentifiesTheVM(t *testing.T) {
	t.Setenv(EnvLogHost, "")

	hostname, err := os.Hostname()
	if err != nil {
		t.Skipf("hostname unavailable: %v", err)
	}

	if got := resolveHost(""); got != hostname {
		t.Errorf("default host = %q, want %q", got, hostname)
	}

	if got := resolveHost("worker-vm-02"); got != "worker-vm-02" {
		t.Errorf("explicit host = %q", got)
	}

	t.Setenv(EnvLogHost, "worker-vm-07")

	if got := resolveHost("ignored"); got != "worker-vm-07" {
		t.Errorf("%s override = %q", EnvLogHost, got)
	}
}

func TestResolveEnvironmentPrefersAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("ENV", "")
	t.Setenv("GIN_MODE", "release")

	if got := resolveEnvironment(""); got != "release" {
		t.Errorf("GIN_MODE fallback = %q, want release", got)
	}

	t.Setenv("APP_ENV", "production")

	if got := resolveEnvironment(""); got != "production" {
		t.Errorf("APP_ENV = %q, want production", got)
	}

	if got := resolveEnvironment("staging"); got != "staging" {
		t.Errorf("explicit environment = %q, want staging", got)
	}
}

func TestResolveLevelDefaultsToInfo(t *testing.T) {
	t.Setenv(EnvLogLevel, "")

	if got := resolveLevel(""); got != logrus.InfoLevel {
		t.Errorf("empty = %v, want info", got)
	}

	if got := resolveLevel("not-a-level"); got != logrus.InfoLevel {
		t.Errorf("invalid = %v, want info", got)
	}

	if got := resolveLevel("debug"); got != logrus.DebugLevel {
		t.Errorf("explicit debug = %v", got)
	}

	t.Setenv(EnvLogLevel, "warn")

	if got := resolveLevel(""); got != logrus.WarnLevel {
		t.Errorf("%s = %v, want warn", EnvLogLevel, got)
	}
}

// TestResolveFormatterDefaultsToJson covers the InitLogger path, which inverts
// the legacy opt-in default because an explicit bootstrap implies a deployment
// that collects structured logs.
func TestResolveFormatterDefaultsToJson(t *testing.T) {
	t.Setenv(EnvLogFormat, "")

	formatter, gcp := resolveFormatter("")
	if _, ok := formatter.(*logrus.JSONFormatter); !ok {
		t.Errorf("default formatter = %T, want *logrus.JSONFormatter", formatter)
	}
	if !gcp {
		t.Error("default must emit the reserved GCP keys")
	}

	formatter, gcp = resolveFormatter("TEXT")
	if _, ok := formatter.(*logrus.TextFormatter); !ok {
		t.Errorf("text formatter = %T, want *logrus.TextFormatter", formatter)
	}
	if gcp {
		t.Error("text mode must not emit the reserved GCP keys")
	}

	t.Setenv(EnvLogFormat, "text")

	if _, gcp = resolveFormatter(""); gcp {
		t.Errorf("%s=text must not emit the reserved GCP keys", EnvLogFormat)
	}
}

// TestApplyDefaultFormatterKeepsLegacyTextDefault protects every SDK consumer
// that has not adopted InitLogger: the lazy fallback must stay opt-in JSON.
func TestApplyDefaultFormatterKeepsLegacyTextDefault(t *testing.T) {
	if LoggerInitialised() {
		t.Skip("InitLogger already owns the global logger in this process")
	}

	t.Setenv(EnvLogFormat, "")

	restore := logrus.StandardLogger().Formatter
	t.Cleanup(func() { logrus.SetFormatter(restore) })

	logrusFormatterOnce = sync.Once{}
	logrus.SetFormatter(&logrus.JSONFormatter{})

	applyDefaultFormatter()

	if _, ok := logrus.StandardLogger().Formatter.(*logrus.TextFormatter); !ok {
		t.Errorf("fallback formatter = %T, want *logrus.TextFormatter", logrus.StandardLogger().Formatter)
	}
}

// TestApplyDefaultFormatterYieldsToInitLogger is the regression that removes
// the need for callers to burn the sync.Once themselves: once InitLogger has
// installed a formatter, the lazy path must not replace it.
func TestApplyDefaultFormatterYieldsToInitLogger(t *testing.T) {
	restoreFormatter := logrus.StandardLogger().Formatter
	t.Cleanup(func() { logrus.SetFormatter(restoreFormatter) })

	loggerMu.Lock()
	restoreInitialised := loggerInitialised
	loggerInitialised = true
	loggerMu.Unlock()

	t.Cleanup(func() {
		loggerMu.Lock()
		loggerInitialised = restoreInitialised
		loggerMu.Unlock()
	})

	logrusFormatterOnce = sync.Once{}

	owned := &logrus.JSONFormatter{TimestampFormat: "owned-by-initlogger"}
	logrus.SetFormatter(owned)

	applyDefaultFormatter()
	Logrus(sdk_cons.INFO, "a log line through the legacy wrapper")

	got, ok := logrus.StandardLogger().Formatter.(*logrus.JSONFormatter)
	if !ok {
		t.Fatalf("formatter = %T, want the one InitLogger installed", logrus.StandardLogger().Formatter)
	}

	if got.TimestampFormat != "owned-by-initlogger" {
		t.Errorf("formatter was replaced: TimestampFormat = %q", got.TimestampFormat)
	}
}

// TestInitLoggerBuildsLabelSet exercises the real bootstrap. It runs last in
// this file by name ordering is not guaranteed, so it tolerates a logger that
// another test already initialised.
func TestInitLoggerBuildsLabelSet(t *testing.T) {
	t.Setenv(EnvLogService, "worker-external-beyond-disburse")
	t.Setenv(EnvLogHost, "worker-vm-02")
	t.Setenv(EnvLogInstance, "2")
	t.Setenv(EnvLogZone, "asia-southeast2-b")
	t.Setenv("APP_ENV", "production")

	InitLogger(&sdk_dto.LoggerOptions{Labels: map[string]string{"team": "payments"}})

	if !LoggerInitialised() {
		t.Fatal("LoggerInitialised() = false after InitLogger")
	}

	labels := LogServiceLabels()

	for key, want := range map[string]string{
		"service":  "worker-external-beyond-disburse",
		"host":     "worker-vm-02",
		"instance": "2",
		"zone":     "asia-southeast2-b",
		"env":      "production",
		"team":     "payments",
	} {
		if labels[key] != want {
			t.Errorf("labels[%q] = %q, want %q", key, labels[key], want)
		}
	}

	if got := LogServiceName(); got != "worker-external-beyond-disburse" {
		t.Errorf("LogServiceName() = %q", got)
	}

	labels["service"] = "mutated"

	if LogServiceLabels()["service"] == "mutated" {
		t.Error("LogServiceLabels() must return a copy")
	}
}
