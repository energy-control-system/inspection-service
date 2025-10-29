package api

import (
	"context"
	"fmt"
	"inspection-service/api/handler"
	"inspection-service/config"
	"inspection-service/service/inspection"

	"github.com/sunshineOfficial/golib/gohttp/gorouter"
	"github.com/sunshineOfficial/golib/gohttp/gorouter/middleware"
	"github.com/sunshineOfficial/golib/gohttp/gorouter/plugin"
	"github.com/sunshineOfficial/golib/gohttp/goserver"
	"github.com/sunshineOfficial/golib/golog"
)

type ServerBuilder struct {
	server goserver.Server
	router *gorouter.Router
}

func NewServerBuilder(ctx context.Context, log golog.Logger, settings config.Settings) *ServerBuilder {
	return &ServerBuilder{
		server: goserver.NewHTTPServer(ctx, log, fmt.Sprintf(":%d", settings.Port)),
		router: gorouter.NewRouter(log).Use(
			middleware.Metrics(),
			middleware.Recover,
			middleware.LogError,
		),
	}
}

func (s *ServerBuilder) AddDebug() {
	s.router.Install(plugin.NewPProf(), plugin.NewMetrics())
}

func (s *ServerBuilder) AddInspections(service *inspection.Service) {
	r := s.router.SubRouter("/inspections")
	r.HandleGet("", handler.GetAllInspections(service))
	r.HandleGet("/task/{taskID}", handler.GetInspectionByTaskID(service))
	r.HandlePost("/{id}/photo", handler.AttachPhotoToInspection(service))
	r.HandlePatch("/{id}/finish", handler.FinishInspection(service))
}

func (s *ServerBuilder) Build() goserver.Server {
	s.server.UseHandler(s.router)

	return s.server
}
