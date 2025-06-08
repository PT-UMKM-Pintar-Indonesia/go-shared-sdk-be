package pkg

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

	if req.Option.FileName != sdk_cons.EMPTY {
		file, err := os.OpenFile(req.Option.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		if req.Option.FileFormatter == nil {
			req.Option.FileFormatter = &logrus.JSONFormatter{
				TimestampFormat: sdk_cons.DATE_TIME_FORMAT,
			}
		}

		logger.SetOutput(file)
		logger.SetFormatter(req.Option.FileFormatter)
	}

	lg := logger.WithFields(req.Option.Fields)

	if req.Option.CustomFields != nil {
		fields := logrus.Fields{}

		if err := sdk_helper.NewTransform().SrcToDest(req.Option.CustomFields, &fields); err != nil {
			return nil, err
		}

		lg = logger.WithFields(fields)
	}

	switch req.Option.Type {

	case sdk_cons.WARN:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Warnf(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Warn(req.Option.Args)
		}

	case sdk_cons.TRACE:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Tracef(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Trace(req.Option.Args)
		}

	case sdk_cons.DEBUG:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Debugf(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Debug(req.Option.Args)
		}

	case sdk_cons.ERROR:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Errorf(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Error(req.Option.Args)
		}

	case sdk_cons.INFO:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Infof(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Info(req.Option.Args)
		}

	case sdk_cons.PANIC:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Panicf(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Panic(req.Option.Args)
		}

	case sdk_cons.FATAL:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Fatalf(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Fatal(req.Option.Args)
		}

	default:
		if req.Option.CustomMessage != sdk_cons.EMPTY && req.Option.Args != nil {
			lg.Printf(req.Body.CustomMessage, req.Option.Args)
		} else {
			lg.Println(req.Option.Args)
		}
	}

	if req.Option.Entry == nil {
		req.Option.Entry = lg.WithTime(time.Now())
	}

	if req.Option.TextFormatter != nil {
		return req.Option.TextFormatter.Format(req.Option.Entry)
	} else if req.Option.JSONFormatter != nil {
		return req.Option.JSONFormatter.Format(req.Option.Entry)
	}

	return nil, nil
}
