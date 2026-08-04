package routes

import (
	"github.com/chiraggarg879/url-shortener/database"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
)

func ResolveURL(c *fiber.Ctx) error {

	url := c.Params("url")

	r := database.CreateClient(0)
	defer r.Close()

	//fetch value from redis
	value, err := r.Get(url).Result()
	if err != redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "shortened url not found in database",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot connect to db",
		})
	}
	//why new connection to db
	rInr := database.CreateClient(1)
	defer rInr.Close()

	//why we are incrementing this
	_ = rInr.Incr("counter")

	return c.Redirect(value, 301)
}
