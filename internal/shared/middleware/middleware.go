package middleware

import (
	"strings"

	"poltekkes-cat-backend/internal/shared/jwt"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Cache-Control, X-Requested-With, X-Participant-Token, X-Trace-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Trace-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func AdminAuthMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization token is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization header format (Bearer <token>)")
			c.Abort()
			return
		}

		claims, err := jwt.ValidateAdminToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Invalid or expired admin access token")
			c.Abort()
			return
		}

		if len(requiredRoles) > 0 {
			hasRole := false
			for _, required := range requiredRoles {
				for _, userRole := range claims.Roles {
					if strings.EqualFold(required, userRole) {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				response.Forbidden(c, "Access forbidden: insufficient role permissions")
				c.Abort()
				return
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("full_name", claims.FullName)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

func ParticipantAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Participant-Token")
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			response.Unauthorized(c, "Participant session token is required (X-Participant-Token)")
			c.Abort()
			return
		}

		claims, err := jwt.ValidateParticipantToken(token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired participant session token")
			c.Abort()
			return
		}

		c.Set("participant_code", claims.ParticipantCode)
		c.Set("participant_name", claims.Name)
		c.Set("token_login", claims.TokenLogin)

		c.Next()
	}
}
