package main

import (
	"context"

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
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	resp, err := router.Handle(a.ctx, name)
	if err != nil {
		return err.Error()
	}
	return resp
}
