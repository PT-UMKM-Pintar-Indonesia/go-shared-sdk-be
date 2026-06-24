package sdk_helper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	sdk_opt "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/outputs"
)

type sizeWriter struct {
	http.ResponseWriter
	size int64
}

func (w *sizeWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += int64(n)
	return n, err
}

func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fms", float64(d.Microseconds())/1000.0)
}

var errorCodeMapping = map[int]string{
	http.StatusBadGateway:          "SERVICE_ERROR",
	http.StatusServiceUnavailable:  "SERVICE_UNAVAILABLE",
	http.StatusGatewayTimeout:      "SERVICE_TIMEOUT",
	http.StatusRequestTimeout:      "SERVICE_TIMEOUT",
	http.StatusConflict:            "DUPLICATE_RESOURCE",
	http.StatusBadRequest:          "INVALID_REQUEST",
	http.StatusUnprocessableEntity: "INVALID_REQUEST",
	http.StatusPreconditionFailed:  "REQUEST_COULD_NOT_BE_PROCESSED",
	http.StatusForbidden:           "ACCESS_DENIED",
	http.StatusUnauthorized:        "UNAUTHORIZED_TOKEN",
	http.StatusNotFound:            "UNKNOWN_RESOURCE",
	http.StatusInternalServerError: "GENERAL_ERROR",
}

func Version(path string) string {
	return fmt.Sprintf("%s/%s", sdk_cons.API, path)
}

func Api(rw http.ResponseWriter, r *http.Request, options sdk_opt.Response) {
	start := time.Now()

	sw := &sizeWriter{ResponseWriter: rw}

	response := buildResponse(options, r, start)
	writeResponse(sw, NewParser(), response)
}

func getErrorCode(statusCode int) string {
	if code, exists := errorCodeMapping[statusCode]; exists {
		return code
	}

	return errorCodeMapping[http.StatusInternalServerError]
}

func getProtocol(r *http.Request) string {
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		return "https"
	}
	return "http"
}

func getIPAddress(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	return r.RemoteAddr
}

func buildResponse(options sdk_opt.Response, r *http.Request, start time.Time) sdk_opt.Response {
	statusCode := options.StatCode
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	errCode := options.ErrCode
	if statusCode >= http.StatusBadRequest && errCode == nil {
		code := getErrorCode(int(statusCode))
		errCode = &code
	}

	errMsg := options.ErrMsg
	if errMsg == "" {
		errMsg = sdk_cons.DEFAULT_ERR_MSG
	}

	return sdk_opt.Response{
		StatCode:   statusCode,
		Message:    options.Message,
		ErrCode:    errCode,
		ErrMsg:     errMsg,
		Data:       options.Data,
		Errors:     options.Errors,
		Pagination: options.Pagination,
		Info: sdk_opt.Info{
			Host:         r.Host,
			Protocol:     getProtocol(r),
			Path:         r.URL.Path,
			Method:       r.Method,
			Timestamp:    time.Now().Format(time.RFC3339),
			ResponseTime: formatDuration(time.Since(start)),
			UserAgent:    r.UserAgent(),
			IPAddress:    getIPAddress(r),
		},
	}
}

func writeResponse(sw *sizeWriter, parser sdk_inf.IParser, response sdk_opt.Response) {
	sw.Header().Set("Content-Type", "application/json")
	sw.WriteHeader(int(response.StatCode))

	if err := parser.Encode(sw, response); err != nil {
		return
	}

	response.Info.ResponseSize = formatBytes(sw.size)
}
