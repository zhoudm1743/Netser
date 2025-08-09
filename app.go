package main

import (
	"context"
	"log"

	"github.com/zhoudm1743/Netser/core"
	"github.com/zhoudm1743/Netser/router"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化消息数据库管理器
	err := core.InitMessageDBManager()
	if err != nil {
		log.Printf("消息数据库管理器初始化失败: %v", err)
	}

	// 初始化WebSocket管理器
	err = core.InitWebSocketManager()
	if err != nil {
		log.Printf("WebSocket管理器初始化失败: %v", err)
	} else {
		// 启动WebSocket服务器
		err = core.GlobalWebSocketManager.StartServer()
		if err != nil {
			log.Printf("WebSocket服务器启动失败: %v", err)
		} else {
			log.Printf("WebSocket服务器启动成功，端口: %d", core.GlobalWebSocketManager.GetPort())
		}
	}
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	// 停止WebSocket服务器
	if core.GlobalWebSocketManager != nil {
		err := core.GlobalWebSocketManager.StopServer()
		if err != nil {
			log.Printf("WebSocket服务器停止失败: %v", err)
		} else {
			log.Printf("WebSocket服务器已停止")
		}
	}

	// 清理所有消息数据库
	if core.GlobalMessageDBManager != nil {
		err := core.GlobalMessageDBManager.CloseAllDatabases()
		if err != nil {
			log.Printf("清理消息数据库失败: %v", err)
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	resp, err := router.Handle(a.ctx, name)
	if err != nil {
		return err.Error()
	}
	return resp
}
