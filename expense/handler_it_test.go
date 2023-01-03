//go:build integration

package expense_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/brown-kaew/assessment/config"
	"github.com/brown-kaew/assessment/expense"
)

func setUp() (config.Config, func()) {
	fmt.Println("setUp")
	config := config.New()
	database, close := expense.InitDB(config)
	service := expense.NewService(database)
	handler := expense.NewHandler(service)
	e := echo.New()
	go func() {
		e.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, "OK")
		})

		handler.InitRoutes(e)
		e.Start(config.Port)
	}()

	for {
		conn, err := net.DialTimeout("tcp", config.Port, 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}

	return config, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		e.Shutdown(ctx)
		close()
	}
}

func TestHealth(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	url := fmt.Sprintf("http://localhost%s/health", config.Port)
	fmt.Println("Arrange", url)
	reqBody := ``
	req, err := http.NewRequest(http.MethodGet, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "\"OK\"\n", string(byteBody))
	}
}
