package routes

import (
	"database/sql"
	"errors"

	"github.com/chiraggarg879/url-shortener/database"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx, db *sql.DB) error {

	url := c.Params("url")

	// r := database.CreateClient(0)
	// defer r.Close()

	// //fetch value from redis
	// value, err := r.Get(url).Result()

	// Fetch the destination URL from MySQL.
	value, err := database.ResolveURL(db, url)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "shortened url not found in database",
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot connect to db",
		})
	}
	//why new connection to db
	rInr := database.CreateClient(1)
	defer rInr.Close()

	//why we are incrementing this
	_ = rInr.Incr("counter")

	return c.Redirect(value, fiber.StatusMovedPermanently)
}
