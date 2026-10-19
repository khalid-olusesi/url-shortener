package controllers

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"
	"urlShortener/initializers"
	"urlShortener/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func CreateURL(c *gin.Context) {
	var body struct {
		OriginalURL string `json:"original_url"`
	}

	err := c.ShouldBindJSON(&body)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read the url",
		})
		return
	}

	if body.OriginalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "a url is needed",
		})
		return
	}

	parsedURL, err := url.Parse(body.OriginalURL)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "enter a valid url",
		})
		return
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid url",
		})
		return
	} //the scheme describes the protocol used by the url

	if parsedURL.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "input the the right hostname",
		})
		return
	}

	const maxRetries = 10 //caps the collision retry loop to prevent infinite retries

	for i := 0; i < maxRetries; i++ {
		shortCode, err := generateShortCode()

		if err != nil {
			log.Printf("failed to generate shortcode:%v", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to gnerate shortcode",
			})
			return
		}

		genCode := models.URL{
			OriginalURL: body.OriginalURL,
			ShortCode:   shortCode,
			ExpiresAt:   time.Now().AddDate(5, 0, 0), //sets the expiry to 5 years from now
		}

		results := initializers.DB.WithContext(c.Request.Context()).Create(&genCode)

		if results.Error != nil {
			var pgErr *pgconn.PgError

			if errors.As(results.Error, &pgErr) && pgErr.Code == "23505" { //the code is an error code(meaning:unique constraint was violated)
				continue
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to save in the database",
			})
			return
		}

		//warm the redis cache so the first redirect is a cache hit
		if initializers.RedisClient != nil {
			cacheErr := initializers.RedisClient.Set(c.Request.Context(), "url:"+genCode.ShortCode, genCode.OriginalURL, time.Hour).Err()

			if cacheErr != nil {
				log.Printf("failed to warm cache: %v", cacheErr)
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"short_code": genCode.ShortCode,
		})

		return

	} //the for loop caps at maxRetries to prevent infinite retries on collisions

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "failed to generate a unique short code",
	})

}

func generateShortCode() (string, error) {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 8

	shortCode := make([]byte, codeLength)

	for i := 0; i < codeLength; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			return "", err
		}

		shortCode[i] = characters[index.Int64()]
	}

	return string(shortCode), nil
}

func GetURL(c *gin.Context) {

	shortcode := c.Param("shortcode") //retrieves the short code from the route(the one the user pastes)

	ctx := c.Request.Context()

	if initializers.RedisClient != nil {
		originalUrl, err := initializers.RedisClient.Get(ctx, "url:"+shortcode).Result()

		if err == nil {
			c.Redirect(http.StatusFound, originalUrl)
			return
		}

		if !errors.Is(err, redis.Nil) {
			log.Printf("redis lookup failed: %v", err) //only log actual redis errors, not cache misses
		}
	}

	var url models.URL

	results := initializers.DB.WithContext(ctx).Where("short_code = ?", shortcode).First(&url)

	if errors.Is(results.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "short url not found",
		})
		return
	}

	if results.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve the URL",
		})
		return
	}

	if !url.ExpiresAt.IsZero() && time.Now().After(url.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{
			"error": "this short url has expired",
		})
		return
	}

	if initializers.RedisClient != nil {
		err := initializers.RedisClient.Set(ctx, "url:"+shortcode, url.OriginalURL, time.Hour).Err()

		if err != nil {
			log.Printf("failed to save url in redis: %v", err)
		}
	}

	c.Redirect(http.StatusFound, url.OriginalURL)

}

//two important redis error, cache miss(the shortcode is not stored in redis) and redis failing
//used hey as a lightweight load testing tool
