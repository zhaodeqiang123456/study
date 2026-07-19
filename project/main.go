package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"simple_service/pkg"
	"syscall"
	"time"
)

func main() {

	srv := pkg.NewService() //构建服务

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("收到关闭信号，开始优雅退出...")

	//  给现有请求一定时间完成（如 5 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("强制关闭: %v", err)
	}

	// 6. 关闭后，Service() 返回，main 继续执行，此时 defer inst.detach() 自动执行
	log.Println("服务已安全退出")

}
