package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/segmentio/encoding/json"
	"github.com/spf13/viper"
	"github.com/zercle/ccxt-proxy/configs"
	"github.com/zercle/ccxt-proxy/internal/datasources"
	"github.com/zercle/ccxt-proxy/internal/routes"
	helpers "github.com/zercle/gofiber-helpers"
)

// PrdMode from GO_ENV
var (
	version string
	build   string
)

func main() {
	var err error

	// load config
	if err = configs.LoadConfig("default"); err != nil {
		log.Panicf("error while loading the env:\n %+v", err)
	}

	// Init datasources
	InitDataSources()

	// Init app
	log.Printf("init app")
	app := fiber.New(fiber.Config{
		ErrorHandler:   customErrorHandler,
		ReadTimeout:    60 * time.Second,
		ReadBufferSize: 8 * 1024,
		Prefork:        true,
		// speed up json with segmentio/encoding
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Logger middleware for Fiber that logs HTTP request/response details.
	app.Use(logger.New())

	// Recover middleware for Fiber that recovers from panics anywhere in the stack chain and handles the control to the centralized ErrorHandler.
	app.Use(recover.New())

	// CORS middleware for Fiber that that can be used to enable Cross-Origin Resource Sharing with various options.
	app.Use(cors.New())

	// set api router
	routes.SetupRoutes(app)

	// Log GO_ENV
	log.Printf("Runtime ENV: %s", viper.GetString("app.env"))
	log.Printf("Version: %s", version)
	log.Printf("Build: %s", build)

	// Listen from a different goroutine
	go func() {
		if err := app.Listen(":"+viper.GetString("app.port.http")); err != nil {
			log.Panic(err)
		}
	}()

	// Create channel to signify a signal being sent
	quit := make(chan os.Signal, 1)
	// When an interrupt or termination signal is sent, notify the channel
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// This blocks the main thread until an interrupt is received
	<-quit
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	fmt.Println("Running cleanup tasks...")
	// Your cleanup tasks go here
	// if datasources.RedisStorage != nil {
	// 	datasources.RedisStorage.Close()
	// }
	fmt.Println("Successful shutdown.")
}

var customErrorHandler = func(c *fiber.Ctx, err error) error {
	// Default 500 statuscode
	code := http.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		// Override status code if fiber.Error type
		code = e.Code
	}

	responseData := helpers.ResponseForm{
		Success: false,
		Errors: []*helpers.ResposeError{
			{
				Code:    code,
				Message: err.Error(),
			},
		},
	}

	// Return statuscode with error message
	err = c.Status(code).JSON(responseData)
	if err != nil {
		// In case the JSON fails
		log.Printf("customErrorHandler: %+v", err)
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Return from handler
	return nil
}

func InitDataSources() (err error) {
	datasources.FastHttpClient = datasources.InitFasthttpClient()
	datasources.JsonParserPool = datasources.InitJsonParserPool()
	return
}
