package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/gomail.v2"
)

// メール送信機能
// 環境変数からSMTPサーバーの情報を取得してメールを送信します。
func sendErrorEmail(err error, c echo.Context) {
	// --- 環境変数の読み込み ---
	// .envファイルから読み込まれます。
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	toEmail := smtpUser // 通知先はSMTP_USERと同じアドレスに設定

	// 環境変数が一つでも設定されていない場合は、エラーをログに出力して終了
	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" {
		c.Logger().Error("SMTP configuration is incomplete. Email not sent.")
		return
	}

	smtpPort, convErr := strconv.Atoi(smtpPortStr)
	if convErr != nil {
		c.Logger().Errorf("Invalid SMTP_PORT: %v", convErr)
		return
	}

	// --- メール作成 ---
	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "[ALERT] API Server Error Detected")
	body := fmt.Sprintf(
		"<h1>API Server Error</h1>"+
			"<p>An error occurred in the API server.</p>"+
			"<ul>"+
			"<li><b>Error:</b> %v</li>"+
			"<li><b>Request URI:</b> %s</li>"+
			"<li><b>Request Method:</b> %s</li>"+
			"</ul>",
		err,
		c.Request().RequestURI,
		c.Request().Method,
	)
	m.SetBody("text/html", body)

	// --- SMTPサーバー経由でメール送信 ---
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	c.Logger().Infof("Sending error email to %s...", toEmail)
	if err := d.DialAndSend(m); err != nil {
		c.Logger().Errorf("Failed to send email: %v", err)
	} else {
		c.Logger().Info("Error email sent successfully.")
	}
}

// カスタムHTTPエラーハンドラ
// この関数がサーバーで発生したすべてのHTTPエラーを処理します。
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	// デフォルトのエラーハンドラを呼び出して、クライアントにエラーレスポンスを返す
	c.Echo().DefaultHTTPErrorHandler(err, c)

	// 500番台のエラー（サーバーエラー）の場合のみメールを送信
	if code >= 500 {
		// ゴルーチンでメールを送信し、クライアントへのレスポンスをブロックしない
		go sendErrorEmail(err, c)
	}
}

func main() {
	// .envファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		// .envファイルがなくてもエラーにせず、ログだけ出力して続行
		log.Println("No .env file found, using environment variables from OS")
	}

	e := echo.New()

	// --- ミドルウェアの設定 ---
	// ロガー: すべてのリクエストをログに出力
	e.Use(middleware.Logger())
	// リカバー: パニックから復旧し、エラーハンドラに処理を渡す
	e.Use(middleware.Recover())

	// --- カスタムエラーハンドラの設定 ---
	e.HTTPErrorHandler = customHTTPErrorHandler

	// --- ルート（エンドポイント）の定義 ---
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to the API Server!")
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// このエンドポイントにアクセスすると意図的にパニックが発生します
	e.GET("/error", func(c echo.Context) error {
		// 意図的にパニックを発生させる
		panic("This is a test panic!")
	})

	// --- サーバーの起動 ---
	e.Logger.Info("Starting server on :1323")
	// .envファイルに関するメッセージ
	fmt.Println("---------------------------------------------------------")
	fmt.Println(" Server is starting...")
	fmt.Println(" It will try to load SMTP settings from the .env file.")
	fmt.Println(" Make sure your .env file contains:")
	fmt.Println("   SMTP_HOST=<your_smtp_host>")
	fmt.Println("   SMTP_PORT=<your_smtp_port>")
	fmt.Println("   SMTP_USER=<your_email_address>")
	fmt.Println("   SMTP_PASS=<your_email_password_or_app_password>")
	fmt.Println("---------------------------------------------------------")

	if err := e.Start(":1323"); err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
