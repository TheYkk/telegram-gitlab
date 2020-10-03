package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gopkg.in/go-playground/webhooks.v5/gitlab"
)

func main() {
	app := fiber.New()
	gitlab.New(gitlab.Options)
	loggerConf := logger.Config{
		Next:       nil,
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02T15:04:05",
		TimeZone:   "Local",
		Output:     os.Stderr,
	}

	app.Use(logger.New(loggerConf))

	// GET /john
	app.Get("/jon", func(c *fiber.Ctx) error {
		msg := fmt.Sprintf("Hello AAA 👋!")
		return c.SendString(msg) // => Hello john 👋!
	})

	log.Fatal(app.Listen(":3000"))
}
