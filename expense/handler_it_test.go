//go:build integration

package expense_test

import (
	"context"
	"encoding/json"
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
	handler := expense.NewHandler(database)
	e := echo.New()
	go func() {
		e.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, "OK")
		})

		handler.InitRoutes(e)
		e.Start(config.Port)
	}()

	checkServerReadiness(config)

	return config, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		e.Shutdown(ctx)
		close()
	}
}

func setUpNoDB() (config.Config, func()) {
	fmt.Println("setUp")
	config := config.New()
	database, close := expense.InitDB(config)
	defer close() //close DB after every thing is set
	handler := expense.NewHandler(database)
	e := echo.New()
	go func() {
		e.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, "OK")
		})

		handler.InitRoutes(e)
		e.Start(config.Port)
	}()

	checkServerReadiness(config)

	return config, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		e.Shutdown(ctx)
		close()
	}
}

func checkServerReadiness(config config.Config) {
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost%s", config.Port), 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}
}

func seedExpenses(t *testing.T, config config.Config) string {
	reqBody := `{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath",
		"tags": [
		  "food",
		  "beverage"
		]
	  }`
	url := fmt.Sprintf("http://localhost%s/expenses", config.Port)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	var expense expense.Expense
	err = json.Unmarshal(byteBody, &expense)

	return expense.Id
}

func TestHealth(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	reqBody := ``
	url := fmt.Sprintf("http://localhost%s/health", config.Port)
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

func TestCreateNewExpense_Success(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	reqBody := `{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath",
		"tags": [
		  "food",
		  "beverage"
		]
	  }`
	url := fmt.Sprintf("http://localhost%s/expenses", config.Port)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	var expense expense.Expense
	err = json.Unmarshal(byteBody, &expense)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NotEmpty(t, expense.Id)
		assert.Equal(t, "strawberry smoothie", expense.Title)
		assert.Equal(t, float64(79), expense.Amount)
		assert.Equal(t, "night market promotion discount 10 bath", expense.Note)
		assert.Equal(t, []string{"food", "beverage"}, expense.Tags)
	}
}

func TestCreateNewExpense_InvalidJsonRequest_ShouldGetBadRequest(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	reqBody := `{
		"title": "strawberry smoothie",
		"amount": "79 bath",
		"note": "night market promotion discount 10 bath",
		"tags": [
		  "food",
		  "beverage"
		]
	  }`
	url := fmt.Sprintf("http://localhost%s/expenses", config.Port)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Contains(t, strings.TrimSpace(string(byteBody)), `"message":"Unmarshal type error`)
	}
}

func TestCreateNewExpense_NoDbConn_ShouldGetInternalServerError(t *testing.T) {
	config, teardown := setUpNoDB()
	defer teardown()

	// Arrange
	reqBody := `{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath",
		"tags": [
		  "food",
		  "beverage"
		]
	  }`
	url := fmt.Sprintf("http://localhost%s/expenses", config.Port)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, strings.TrimSpace(string(byteBody)), `"message":"Internal Server Error"`)
	}
}

func TestGetExpenseById_Success(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	id := seedExpenses(t, config)
	reqBody := ``
	url := fmt.Sprintf("http://localhost%s/expenses/%s", config.Port, id)
	req, err := http.NewRequest(http.MethodGet, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	var expense expense.Expense
	err = json.Unmarshal(byteBody, &expense)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, expense.Id)
		assert.Equal(t, "strawberry smoothie", expense.Title)
		assert.Equal(t, float64(79), expense.Amount)
		assert.Equal(t, "night market promotion discount 10 bath", expense.Note)
		assert.Equal(t, []string{"food", "beverage"}, expense.Tags)
	}
}

func TestGetExpenseById_NoIdIsFound_ShouldGetNotFound(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	id := "999"
	reqBody := ``
	url := fmt.Sprintf("http://localhost%s/expenses/%s", config.Port, id)
	req, err := http.NewRequest(http.MethodGet, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	var expense expense.Expense
	err = json.Unmarshal(byteBody, &expense)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Contains(t, strings.TrimSpace(string(byteBody)), `"message":"Expense not found"`)
		assert.Empty(t, expense.Id)
		assert.Empty(t, expense.Title)
		assert.Empty(t, expense.Amount)
		assert.Empty(t, expense.Note)
		assert.Empty(t, expense.Tags)
	}
}

func TestGetExpenseById_NoDbConn_ShouldGetInternalServerError(t *testing.T) {
	config, teardown := setUpNoDB()
	defer teardown()

	// Arrange
	reqBody := ``
	url := fmt.Sprintf("http://localhost%s/expenses/1", config.Port)
	req, err := http.NewRequest(http.MethodGet, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, strings.TrimSpace(string(byteBody)), `"message":"Cannot prepare statment"`)
	}
}

func TestUpdateExpenseById_Success(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	id := seedExpenses(t, config)
	reqBody := `{
		"title": "MaMa",
		"amount": 5,
		"note": "Yummy",
		"tags": [
		  "food"
		]
	  }`
	url := fmt.Sprintf("http://localhost%s/expenses/%s", config.Port, id)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	var expense expense.Expense
	err = json.Unmarshal(byteBody, &expense)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, id, expense.Id)
		assert.Equal(t, "MaMa", expense.Title)
		assert.Equal(t, float64(5), expense.Amount)
		assert.Equal(t, "Yummy", expense.Note)
		assert.Equal(t, []string{"food"}, expense.Tags)
	}
}

func TestUpdateExpenseById_NoIdIsFound_ShouldGetNotFound(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	id := "999"
	reqBody := `{
		"title": "MaMa",
		"amount": 5,
		"note": "Yummy",
		"tags": [
		  "food"
		]
	  }`
	url := fmt.Sprintf("http://localhost%s/expenses/%s", config.Port, id)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	var expense expense.Expense
	err = json.Unmarshal(byteBody, &expense)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Contains(t, strings.TrimSpace(string(byteBody)), `"message":"Expense not found"`)
		assert.Empty(t, expense.Id)
		assert.Empty(t, expense.Title)
		assert.Empty(t, expense.Amount)
		assert.Empty(t, expense.Note)
		assert.Empty(t, expense.Tags)
	}
}

func TestGetAllExpenses_Success(t *testing.T) {
	config, teardown := setUp()
	defer teardown()

	// Arrange
	seedExpenses(t, config)
	seedExpenses(t, config)
	reqBody := ``
	url := fmt.Sprintf("http://localhost%s/expenses", config.Port)
	req, err := http.NewRequest(http.MethodGet, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	var expenses []expense.Expense
	err = json.Unmarshal(byteBody, &expenses)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, len(expenses) >= 2)
	}
}

func TestGetAllExpenses_NoDbConn_ShouldGetInternalServerError(t *testing.T) {
	config, teardown := setUpNoDB()
	defer teardown()

	// Arrange
	seedExpenses(t, config)
	seedExpenses(t, config)
	reqBody := ``
	url := fmt.Sprintf("http://localhost%s/expenses", config.Port)
	req, err := http.NewRequest(http.MethodGet, url, strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{}

	//Act
	resp, err := client.Do(req)
	assert.NoError(t, err)
	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	//Assert
	//Assert
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, strings.TrimSpace(string(byteBody)), `"message":"Cannot prepare statment"`)
	}
}
