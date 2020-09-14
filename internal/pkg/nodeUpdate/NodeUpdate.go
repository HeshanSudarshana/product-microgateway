package nodeUpdate

import (
	"fmt"
	cachev2 "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	"github.com/sirupsen/logrus"
	"github.com/wso2/micro-gw/internal/pkg/models"
	oasParser "github.com/wso2/micro-gw/internal/pkg/oasparser"
	"github.com/wso2/micro-gw/internal/pkg/oasparser/cache"
	"sync/atomic"
)

var (
	version int32
)

type logger struct {
	*logrus.Logger
}

var log = &logger{
	Logger: logrus.StandardLogger(),
}

func UpdateEnvoyMgw(swaggerFiles []models.SwaggerFile) {
	var nodeId string
	if len(cache.Cache.GetStatusKeys()) > 0 {
		nodeId = cache.Cache.GetStatusKeys()[0]
	}

	listeners, clusters, routes, endpoints := oasParser.GetProductionSourcesMgw(swaggerFiles)

	atomic.AddInt32(&version, 1)
	log.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot Version " + fmt.Sprint(version))
	snap := cachev2.NewSnapshot(fmt.Sprint(version), endpoints, clusters, routes, listeners, nil)
	snap.Consistent()

	err := cache.Cache.SetSnapshot(nodeId, snap)
	if err != nil {
		logrus.Error(err)
	}
}
