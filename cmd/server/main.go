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
	if err := logger.Init("debug", "development", "console"); err != nil {
		panic(err)
	}
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
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		api.POST("/send", rateLimitMiddleware(rateLimiter), authMiddleware(cfg), sendEmailHandler(sender))
		api.POST("/verify", rateLimitMiddleware(rateLimiter), authMiddleware(cfg), verifyDomainHandler())
		api.POST("/verify-sender", rateLimitMiddleware(rateLimiter), authMiddleware(cfg), verifySenderHandler())
		api.POST("/send-verified", rateLimitMiddleware(rateLimiter), authMiddleware(cfg), sendVerifiedEmailHandler(sender))
	}

	auth := r.Group("/auth")
	{
		auth.POST("/login", rateLimitMiddleware(rateLimiter), loginHandler(cfg))
		auth.POST("/refresh", rateLimitMiddleware(rateLimiter), refreshTokenHandler(cfg))
	}

	logger.Info("Starting API server", logger.Field("port", cfg.APIPort))
	if err := r.Run(fmt.Sprintf(":%d", cfg.APIPort)); err != nil {
		logger.Fatal("Failed to start API server", logger.Err(err))
	}
}

func sendEmailHandler(sender *email.Sender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req email.SendRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	

		if err := sender.Send(req.From, req.To, req.Subject, req.Body, req.HTML); err != nil {
			logger.Error("Failed to send email", logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
	}
}

func verifyDomainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Domain string `json:"domain" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		results := make(map[string]string)

		// Check MX record (essential)
		if err := domainVerifier.VerifyMXRecord(req.Domain); err == nil {
			results["MX"] = "Valid"
		} else {
			results["MX"] = fmt.Sprintf("Invalid: %v", err)
		}

		// Check SPF record (recommended)
		if err := domainVerifier.VerifySPFRecord(req.Domain); err == nil {
			results["SPF"] = "Valid"
		} else {
			results["SPF"] = fmt.Sprintf("Invalid: %v", err)
		}

		// Determine overall status
		if results["MX"] == "Valid" {
			status := "Domain verified for sending"
			if results["SPF"] == "Valid" {
				status += " with SPF"
			} else {
				status += " (SPF recommended)"
			}
			c.JSON(http.StatusOK, gin.H{
				"message": status,
				"results": results,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Domain not verified for sending",
				"results": results,
			})
		}
	}
}

func verifySenderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		isValid, err := auth.VerifyEmail(req.Email)
		if err != nil {
			logger.Error("Failed to verify email", logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
			return
		}

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or domain not properly configured"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
	}
}

func sendVerifiedEmailHandler(sender *email.Sender) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req email.SendRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if err := req.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := sender.SendWithVerifiedSender(req.From, req.To, req.Subject, req.Body, req.ReplyTo); err != nil {
			logger.Error("Failed to send email", logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
	}
}

func loginHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
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
			logger.Error("Failed to generate token", logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func refreshTokenHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		newToken, err := auth.RefreshToken(req.Token, cfg.JWTSecret)
		if err != nil {
			logger.Error("Failed to refresh token", logger.Err(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": newToken})
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