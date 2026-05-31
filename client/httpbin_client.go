package client

import (
	"context"
	"evasbr/mclamg/model"
)

type HttpBinClient interface {
	PostMethod(ctx context.Context, requestBody *model.HttpBin, response *map[string]interface{})
}
