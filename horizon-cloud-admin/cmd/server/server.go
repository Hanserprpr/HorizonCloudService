package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"horizon-cloud-admin/config"
	"horizon-cloud-admin/internal/global/database"
	"horizon-cloud-admin/internal/global/httpclient"
	"horizon-cloud-admin/internal/global/logger"
	"horizon-cloud-admin/internal/global/middleware"
	"horizon-cloud-admin/internal/module"
	"horizon-cloud-admin/tools"
	"log/slog"
)

var log *slog.Logger

func Init() {
	config.Init()
	log = logger.New("Server")

	database.Init()

	httpclient.Init()

	for _, m := range module.Modules {
		log.Info(fmt.Sprintf("Init Module: %s", m.GetName()))
		m.Init()
	}
}

func Run() {
	gin.SetMode(string(config.Get().Mode))
	r := gin.New()

	switch config.Get().Mode {
	case config.ModeRelease:
		r.Use(middleware.Logger(logger.Get()))
	case config.ModeDebug:
		r.Use(gin.Logger())
	}
	r.Use(middleware.Cors())
	r.Use(middleware.Recovery())

	for _, m := range module.Modules {
		log.Info(fmt.Sprintf("Init Router: %s", m.GetName()))
		m.InitRouter(r.Group("/" + config.Get().Prefix))
	}
	err := r.Run(config.Get().Host + ":" + config.Get().Port)
	tools.PanicOnErr(err)
}
