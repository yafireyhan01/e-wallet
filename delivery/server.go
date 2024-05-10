package delivery

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/yafireyhan01/e-wallet/config"
	"github.com/yafireyhan01/e-wallet/delivery/controller"
	"github.com/yafireyhan01/e-wallet/delivery/middleware"
	"github.com/yafireyhan01/e-wallet/manager"
)

type Server struct {
	uc     manager.UseCaseManager
	engine *gin.Engine
	host   string
}

func (s *Server) setupControllers() {
	rg := s.engine.Group("/api/v1")
	rg.Use(middleware.LogMiddleware())
	controller.NewTransferController(s.uc.TransferUseCase(), s.uc.UserUseCase(), rg).Route()
	controller.NewTopupController(s.uc.TopupUseCase(), s.uc.UserUseCase(), rg).Route()
	controller.NewUserController(s.uc.UserUseCase(), rg).Route()
	controller.NewAdminController(s.uc.AdminUseCase(), s.uc.UserUseCase(), rg).Route()
}

func (s *Server) Run() {
	s.setupControllers()
	if err := s.engine.Run(s.host); err != nil {
		log.Fatal("server can't run")
	}
}

func NewServer() *Server {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	infra, err := manager.NewInfraManager(cfg)
	if err != nil {
		log.Fatal(err)
	}
	repo := manager.NewRepoManager(infra)
	uc := manager.NewUseCaseManager(repo)
	engine := gin.Default()
	host := fmt.Sprintf(":%s", cfg.ApiPort)
	return &Server{
		uc:     uc,
		engine: engine,
		host:   host,
	}
}
