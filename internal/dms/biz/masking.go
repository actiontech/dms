package biz

import (
	maskingBiz "github.com/actiontech/dms/internal/data_masking/biz"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

type DataMaskingUsecase struct {
	log         *utilLog.Helper
	DataMasking *maskingBiz.DataMaskingUseCase
}

func NewMaskingUsecase(log utilLog.Logger, dataMaskingUsecase *maskingBiz.DataMaskingUseCase) *DataMaskingUsecase {
	return &DataMaskingUsecase{
		log:         utilLog.NewHelper(log, utilLog.WithMessageKey("biz.masking")),
		DataMasking: dataMaskingUsecase,
	}
}

type ListMaskingRule struct {
	MaskingType     string   `json:"masking_type"`
	Description     string   `json:"description"`
	ReferenceFields []string `json:"reference_fields"`
	Effect          string   `json:"effect"`
}
