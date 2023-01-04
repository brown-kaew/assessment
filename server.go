package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brown-kaew/assessment/config"
	"github.com/brown-kaew/assessment/expense"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	banner()
	config := config.New()
	database, close := expense.InitDB(config)
	defer close()
	handler := expense.NewHandler(database)

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	handler.InitRoutes(e)

	go func() {
		if err := e.Start(config.Port); err != nil && err != http.ErrServerClosed { // Start server
			e.Logger.Fatal("shutting down the server")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Info("Server stopped")
}

func banner() {
	fmt.Println(`
     ____                              
    / __/_ __ ___  ___ ___  ___ ___ ___
   / _/ \ \ // _ \/ -_) _ \(_-</ -_|_-<
  /___//_\_\/ .__/\__/_//_/___/\__/___/
           /_/                         
   Go : software-engineering/assessment
	`)
}
