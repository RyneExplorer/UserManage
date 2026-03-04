package app

import (
	"UserManagement/internal/config"
	"UserManagement/internal/controller"
	"UserManagement/internal/middleware"
	"UserManagement/internal/repository/mysql"
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

	userRepo := mysql.NewUserRepositoryMySQL(db)
	userService := service.NewUserService(userRepo)

	renderer := view.NewRenderer("static")
	sessions := middleware.NewSessionStore()

	authController := &controller.AuthController{
		Users:      userService,
		Sessions:   sessions,
		CookieName: cfg.SessionCookieName,
		Renderer:   renderer,
	}
	userController := &controller.UserController{
		Users:    userService,
		Renderer: renderer,
	}

	handler := router.NewMux(cfg, authController, userController, sessions)

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}
	fmt.Printf("服务器启动, 监听端口%s...\n", cfg.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
