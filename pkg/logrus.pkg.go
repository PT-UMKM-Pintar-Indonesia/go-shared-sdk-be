package sdk_pkg

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"

	sdk_cons "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/constants"
	sdk_dto "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/dtos"
	sdk_helper "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/helpers"
)

func Logrus(Type string, Msg any, Args ...any) {
	format := false
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: time.RFC3339,
	})

	if Args != nil {
		format = true
	}

	switch Type {

	case sdk_cons.INFO:
		if format {
			logrus.Infof(Msg.(string), Args...)
		} else {
			logrus.Info(Msg)
		}

	case sdk_cons.ERROR:
		if format {
			logrus.Errorf(Msg.(string), Args...)
		} else {
			logrus.Error(Msg)
		}

	case sdk_cons.PRINT:
		if format {
			logrus.Printf(Msg.(string), Args...)
		} else {
			logrus.Print(Msg)
		}

	case sdk_cons.FATAL:
		if format {
			logrus.Fatalf(Msg.(string), Args...)
		} else {
			logrus.Fatal(Msg)
		}

	case sdk_cons.DEBUG:
		if format {
			logrus.Debugf(Msg.(string), Args...)
		} else {
			logrus.Debug(Msg)
		}

	case sdk_cons.PANIC:
		if format {
			logrus.Panicf(Msg.(string), Args...)
		} else {
			logrus.Panic(Msg)
		}

	case sdk_cons.TRACE:
		if format {
			logrus.Tracef(Msg.(string), Args...)
		} else {
			logrus.Trace(Msg)
		}

	default:
		logrus.Println(Msg)

	}
}

func LogrusCustomLogger(req sdk_dto.Request[sdk_dto.LogrusCustomLogger]) ([]byte, error) {
	logger := logrus.New()

	if req.Payload.FileName != sdk_cons.EMPTY {
		file, err := os.OpenFile(req.Payload.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		if req.Payload.FileFormatter == nil {
			req.Payload.FileFormatter = &logrus.JSONFormatter{
				TimestampFormat: sdk_cons.DATE_TIME_FORMAT,
			}
		}

		logger.SetOutput(file)
		logger.SetFormatter(req.Payload.FileFormatter)
	}

	lg := logger.WithFields(req.Payload.Fields)

	if req.Payload.CustomFields != nil {
		fields := logrus.Fields{}

		if err := sdk_helper.NewTransform().SrcToDest(req.Payload.CustomFields, &fields); err != nil {
			return nil, err
		}

		lg = logger.WithFields(fields)
	}

	switch req.Payload.Type {

	case sdk_cons.WARN:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Warnf(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Warn(req.Payload.Args)
		}

	case sdk_cons.TRACE:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Tracef(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Trace(req.Payload.Args)
		}

	case sdk_cons.DEBUG:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Debugf(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Debug(req.Payload.Args)
		}

	case sdk_cons.ERROR:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Errorf(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Error(req.Payload.Args)
		}

	case sdk_cons.INFO:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Infof(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Info(req.Payload.Args)
		}

	case sdk_cons.PANIC:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Panicf(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Panic(req.Payload.Args)
		}

	case sdk_cons.FATAL:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Fatalf(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Fatal(req.Payload.Args)
		}

	default:
		if req.Payload.CustomMessage != sdk_cons.EMPTY && req.Payload.Args != nil {
			lg.Printf(req.Payload.CustomMessage, req.Payload.Args)
		} else {
			lg.Println(req.Payload.Args)
		}
	}

	if req.Payload.Entry == nil {
		req.Payload.Entry = lg.WithTime(time.Now())
	}

	if req.Payload.TextFormatter != nil {
		return req.Payload.TextFormatter.Format(req.Payload.Entry)
	} else if req.Payload.JSONFormatter != nil {
		return req.Payload.JSONFormatter.Format(req.Payload.Entry)
	}

	return nil, nil
}
