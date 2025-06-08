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
)
