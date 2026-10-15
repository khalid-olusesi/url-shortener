package controllers

import (
	"context"
	"errors"
	"log"
	"math/rand"
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

	for {
		shortCode := generateShortCode()

		genCode := models.URL{
			OriginalURL: body.OriginalURL,
			ShortCode:   shortCode,
		}

		results := initializers.DB.Save(&genCode)

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

		c.JSON(http.StatusCreated, gin.H{
			"short_code": genCode.ShortCode,
		})

		break

	} //the for loop added here gives me the chance to run indefinitely if there is a collision(same random code generated), till a unique one is formed again

}

func generateShortCode() string {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	shortCode := ""

	for i := 0; i < 6; i++ { //loops through for 6 indexes, terminates when it gets to index 6
		index := rand.Intn(len(characters)) // generates 6 random indexes from the characters

		shortCode = shortCode + string(characters[index]) //brings out the index charcaters of the random generated and shortcode is added for the string
	}

	return shortCode
}

func GetURL(c *gin.Context) {

	shortcode := c.Param("shortcode") //retrieves the short code from the route(the one the user pastes)

	ctx := context.Background()

	if initializers.RedisClient != nil {
		originalUrl, err := initializers.RedisClient.Get(ctx, "url:"+shortcode).Result()

		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("redis lookup failed:%v", err)
		}

		if err == nil {
			c.Redirect(http.StatusFound, originalUrl)
			return
		}
	}

	var url models.URL

	results := initializers.DB.Where("short_code = ?", shortcode).First(&url)

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

	if initializers.RedisClient != nil {
		err := initializers.RedisClient.Set(ctx, "url:"+shortcode, url.OriginalURL, time.Hour).Err()

		if err != nil {
			log.Printf("failed to save url in redis: %v", err)
		}
	}

	c.Redirect(http.StatusFound, url.OriginalURL)

}

//two important redis error, cache miss(the shortcode is not stored in redis) and redis failing
