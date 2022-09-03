package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zercle/ccxt-proxy/internal/handlers"
)

// SetupRoutes is the Router for GoFiber App
func SetupRoutes(app *fiber.App) {

	app.Get("/", handlers.Index())

	// binance
	groupBinance := app.Group("/binance")
	{
		groupBinance.Get("/", handlers.Index())
	}

	// kucoin
	groupKucoin := app.Group("/kucoin")
	{
		groupKucoin.Get("/", handlers.Index())
	}
}
