package main

import (
	"context"

	"SwiftN2N/internal/edge"
)

// App is the Wails-bound application facade.
type App struct {
	ctx  context.Context
	edge *edge.Manager
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{
		edge: edge.NewManager(),
	}
}

// startup is called when the app starts. The context is saved so runtime
// events can be emitted from process goroutines.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.edge.SetContext(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	_, _ = a.edge.Stop()
}

func (a *App) StartEdge(config edge.Config) (edge.Status, error) {
	return a.edge.Start(config)
}

func (a *App) StopEdge() (edge.Status, error) {
	return a.edge.Stop()
}

func (a *App) GetStatus() edge.Status {
	return a.edge.Status()
}

func (a *App) ValidateConfig(config edge.Config) error {
	return a.edge.Validate(config)
}

func (a *App) GetEdgeVersion(config edge.Config) (string, error) {
	return a.edge.Version(config)
}

func (a *App) CheckEnvironment(config edge.Config) edge.EnvironmentStatus {
	return a.edge.Environment(config)
}
