package sdk_dto

import "github.com/sirupsen/logrus"

type (
	LogrusCustomFields struct {
		Code  string `json:"log_code,omitempty"`
		Name  string `json:"log_name,omitempty"`
		Type  string `json:"log_type,omitempty"`
		Msg   string `json:"log_msg,omitempty"`
		Data  any    `json:"log_data,omitempty"`
		Stack error  `json:"log_stack,omitempty"`
		Time  string `json:"log_time,omitempty"`
	}

	LogrusCustomLogger struct {
		FileName         string
		Type             string
		Entry            *logrus.Entry
		Fields           logrus.Fields
		FileFormatter    *logrus.JSONFormatter
		LogFormatterType string
		JSONFormatter    *logrus.JSONFormatter
		TextFormatter    *logrus.TextFormatter
		Args             any
		CustomMessage    string
		CustomFields     *LogrusCustomFields
	}

	// LoggerOptions configures the global logger bootstrap. Every field is
	// optional: a blank field falls back to its environment variable, then to
	// a derived default. See sdk_pkg.InitLogger.
	LoggerOptions struct {
		// Service is the value of labels.service. Defaults to LOG_SERVICE_NAME,
		// then to the binary name.
		Service string

		// Instance distinguishes replicas of one service, such as the %i of a
		// systemd template unit. Defaults to LOG_SERVICE_INSTANCE.
		Instance string

		// Host is the value of labels.host. Defaults to LOG_SERVICE_HOST, then
		// to the kernel hostname, which on GCE is the instance name.
		Host string

		// Zone optionally records the GCP zone or region of the VM. Defaults
		// to LOG_SERVICE_ZONE.
		Zone string

		// Environment is the value of labels.env. Defaults to APP_ENV, then
		// ENV, then GIN_MODE.
		Environment string

		// Level is a logrus level name. Defaults to LOG_LEVEL, then info.
		Level string

		// Format is "json" or "text". Defaults to LOG_FORMAT, then json.
		Format string

		// Labels are merged into the shared label set, overriding the derived
		// keys on conflict.
		Labels map[string]string
	}
)
