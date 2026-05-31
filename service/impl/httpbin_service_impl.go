package impl

import (
	"context"
	"evasbr/mclamg/client"
	"evasbr/mclamg/common"
	"evasbr/mclamg/model"
	"evasbr/mclamg/service"
	"github.com/sirupsen/logrus"
)

func NewHttpBinServiceImpl(httpBinClient *client.HttpBinClient) service.HttpBinService {
	return &httpBinServiceImpl{
		HttpBinClient: *httpBinClient,
		log:           common.Log.WithField("scope", "HttpBinService"),
	}
}

type httpBinServiceImpl struct {
	client.HttpBinClient
	log *logrus.Entry
}

func (h *httpBinServiceImpl) PostMethod(ctx context.Context) {
	httpBin := model.HttpBin{
		Name: "rizki",
	}
	var response map[string]interface{}
	h.HttpBinClient.PostMethod(ctx, &httpBin, &response)
	h.log.WithContext(ctx).Info("log response service ", response)
}
