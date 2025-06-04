package sdk_helper

import (
	sdk_inf "github.com/PT-UMKM-Pintar-Indonesia/shared-sdk/interfaces"
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
