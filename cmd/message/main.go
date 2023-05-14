package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"notifyGo/internal/api/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := ":8080"
	srv := server.NewServer(addr)
	go func() {
		// 后面一个err是为啥
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动失败{%v}", err)
		}
		log.Fatalf("服务启动, 端口：%s...", addr)
	}()

	// 服务优雅退出
	gracefulShutdown(srv)
}

func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Println("服务准备停止（5s后强制停止）。等待业务处理完毕 ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println(fmt.Sprintf("服务停止失败: %v", err))
	}
	log.Println("服务退出")
}
