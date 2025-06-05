package sdk_helper

import (
	"os"

	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
	"github.com/caarlos0/env"
	"github.com/spf13/viper"
)

type transform struct{}

func NewTransform() sdk_inf.ITransform {
	return transform{}
}

func (h transform) SrcToDest(src, dest any) error {
	helper := NewParser()

	srcByte, err := helper.Marshal(src)
	if err != nil {
		return err
	}

	if err = helper.Unmarshal(srcByte, dest); err != nil {
		return err
	}

	return nil
}

func (h transform) ReqToRes(src, dest any) error {
	helper := NewParser()

	srcByte, err := helper.Marshal(src)
	if err != nil {
		return err
	}

	if err = helper.Unmarshal(srcByte, dest); err != nil {
		return err
	}

	return nil
}

func (h transform) ResToReq(src, dest any) error {
	helper := NewParser()

	srcByte, err := helper.Marshal(src)
	if err != nil {
		return err
	}

	if err = helper.Unmarshal(srcByte, dest); err != nil {
		return err
	}

	return nil
}

func (h transform) EnvToStruct(name, path, ext string, dest any) error {
	if _, ok := os.LookupEnv("GO_ENV"); !ok {
		viper.SetConfigName(name)
		viper.SetConfigType(ext)
		viper.AddConfigPath(path)
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		if err := viper.Unmarshal(&dest); err != nil {
			return err
		}
	} else {
		if err := env.Parse(&dest); err != nil {
			return err
		}
	}

	return nil
}
