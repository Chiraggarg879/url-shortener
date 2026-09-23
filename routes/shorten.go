package routes

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/chiraggarg879/url-shortener/database"
	"github.com/chiraggarg879/url-shortener/helpers"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx, db *sql.DB) error {

	//creating a request url
	body := new(request)

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	//implement rate limiting
	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		// val,_ = r2.Get(c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "Rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	//check if the input is an actual URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	//check for domain error

	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Url can't be corrected"})
	}

	//enforce https,SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	//check for customShort

	var id string
	if body.CustomShort != "" {
		id = body.CustomShort
	} else {
		id = uuid.New().String()[:6]
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ = r.Get(id).Result()

	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "custom short url already in use",
		})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = database.CreateURL(db, id, body.URL, time.Now().Add(body.Expiry*time.Hour))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to save url in db",
		})
	}

	err = r.Set(id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to connect to db",
		})
	}

	r2.Decr(c.IP())

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	val, _ = r2.Get(c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
