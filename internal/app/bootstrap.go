package app

import (
	"UserManagement/internal/config"
	"UserManagement/internal/middleware"
	"UserManagement/internal/repository"
	"UserManagement/internal/router"
	"UserManagement/internal/service"
	"UserManagement/internal/view"
	"fmt"
	"log"
	"net/http"
)

func Start() {
	cfg := config.Load()

	db, err := config.OpenDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	renderer := view.NewRenderer("static")
	sessions := middleware.NewSessionStore()

	r := router.NewRouter(cfg, userService, sessions, renderer)
	handler := r.Handler()

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}
	fmt.Printf("服务器启动, 监听端口%s...\n", cfg.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
