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
	conf := config.New()
	database, closeDB := expense.InitDB(conf)
	defer closeDB()

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})

	g := e.Group("")
	g.Use(config.HardCodeAuth)
	expense.NewHandler(database, g)

	go func() {
		if err := e.Start(conf.Port); err != nil && err != http.ErrServerClosed { // Start server
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
