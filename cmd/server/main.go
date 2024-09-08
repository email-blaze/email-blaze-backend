package main

import (
	"email-blaze/internals/auth"
	"email-blaze/internals/config"
	"email-blaze/internals/email"
	"email-blaze/internals/logger"
	"email-blaze/internals/ratelimit"
	"email-blaze/internals/smtp"
	"email-blaze/pkg/domainVerifier"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Fatal("Failed to load config", logger.Err(err))
	}

	sender := email.NewSender(cfg)
	rateLimiter := ratelimit.NewRateLimiter(cfg.RateLimit, cfg.RateLimit)

	go func() {
		if err := smtp.StartSMTPServer(cfg, sender); err != nil {
			logger.Error("Failed to start SMTP server", logger.Err(err))
			logger.Fatal("Exiting due to SMTP server failure")
		} else {
			logger.Info("SMTP server started successfully")
		}
	}()

	r := gin.Default()

	r.POST("/api/send", rateLimitMiddleware(rateLimiter), authMiddleware(cfg), func(c *gin.Context) {
		var req struct {
			From    string `json:"from"`
			To      string `json:"to"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := sender.Send(req.From, req.To, req.Subject, req.Body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
	})

	r.POST("/api/verify", rateLimitMiddleware(rateLimiter), authMiddleware(cfg), func(c *gin.Context) {
		var req struct {
			Domain   string `json:"domain"`
			Selector string `json:"selector"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		results, err := domainVerifier.VerifyDomain(req.Domain)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !results["MX"] || !results["SPF"] || !results["DKIM"] || !results["DMARC"] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Domain not verified"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Domain verified successfully"})
	})

	

	r.POST("/api/auth/login", rateLimitMiddleware(rateLimiter), func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Implement user authentication logic
		user := &auth.User{
			ID:     1,
			Email:  req.Email,
			Domain: "example.com",
		}

		token, err := auth.GenerateToken(user, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	logger.Info("Starting API server", logger.Field("port", cfg.APIPort))
	if err := r.Run(fmt.Sprintf(":%d", cfg.APIPort)); err != nil {
		logger.Fatal("Failed to start API server", logger.Err(err))
	}
}

func rateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func authMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		claims, err := auth.VerifyToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}