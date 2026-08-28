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

func LogrusCustomLogger(req *sdk_dto.LogrusCustomLogger) ([]byte, error) {
	logger := logrus.New()

	if req.FileName != sdk_cons.EMPTY {
		file, err := os.OpenFile(req.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		if req.FileFormatter == nil {
			req.FileFormatter = &logrus.JSONFormatter{
				TimestampFormat: sdk_cons.DATE_TIME_FORMAT,
			}
		}

		logger.SetOutput(file)
		logger.SetFormatter(req.FileFormatter)
	}

	lg := logger.WithFields(req.Fields)

	if req.CustomFields != nil {
		fields := logrus.Fields{}

		if err := sdk_helper.NewTransform().SrcToDest(req.CustomFields, &fields); err != nil {
			return nil, err
		}

		lg = logger.WithFields(fields)
	}

	switch req.Type {

	case sdk_cons.WARN:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Warnf(req.CustomMessage, req.Args)
		} else {
			lg.Warn(req.Args)
		}

	case sdk_cons.TRACE:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Tracef(req.CustomMessage, req.Args)
		} else {
			lg.Trace(req.Args)
		}

	case sdk_cons.DEBUG:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Debugf(req.CustomMessage, req.Args)
		} else {
			lg.Debug(req.Args)
		}

	case sdk_cons.ERROR:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Errorf(req.CustomMessage, req.Args)
		} else {
			lg.Error(req.Args)
		}

	case sdk_cons.INFO:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Infof(req.CustomMessage, req.Args)
		} else {
			lg.Info(req.Args)
		}

	case sdk_cons.PANIC:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Panicf(req.CustomMessage, req.Args)
		} else {
			lg.Panic(req.Args)
		}

	case sdk_cons.FATAL:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Fatalf(req.CustomMessage, req.Args)
		} else {
			lg.Fatal(req.Args)
		}

	default:
		if req.CustomMessage != sdk_cons.EMPTY && req.Args != nil {
			lg.Printf(req.CustomMessage, req.Args)
		} else {
			lg.Println(req.Args)
		}
	}

	if req.Entry == nil {
		req.Entry = lg.WithTime(time.Now())
	}

	if req.TextFormatter != nil {
		return req.TextFormatter.Format(req.Entry)
	} else if req.JSONFormatter != nil {
		return req.JSONFormatter.Format(req.Entry)
	}

	return nil, nil
}
