package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend-playground/notification" // 新しいnotificationパッケージをインポート

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// カスタムHTTPエラーハンドラ
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	c.Echo().DefaultHTTPErrorHandler(err, c)

	if code >= 500 {
		// DISCORD_WEBHOOK_URL があればDiscordに、なければメールに通知
		if os.Getenv("DISCORD_WEBHOOK_URL") != "" {
			go notification.SendDiscordNotification(err, c.Request().Method, c.Request().RequestURI)
		} else {
			go notification.SendErrorEmail(err, c.Request().Method, c.Request().RequestURI)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables from OS")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = customHTTPErrorHandler

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to the API Server!")
	})
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/error", func(c echo.Context) error {
		panic("This is a test panic!")
	})

	e.Logger.Info("Starting server on :1323")
	fmt.Println("---------------------------------------------------------")
	fmt.Println(" Server is starting...")
	fmt.Println(" Notification settings loaded from .env file:")
	fmt.Println(" - If DISCORD_WEBHOOK_URL is set, alerts will be sent to Discord.")
	fmt.Println(" - Otherwise, alerts will be sent to email as a fallback.")
	fmt.Println("   (Requires SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS)")
	fmt.Println("---------------------------------------------------------")

	if err := e.Start(":1323"); err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
