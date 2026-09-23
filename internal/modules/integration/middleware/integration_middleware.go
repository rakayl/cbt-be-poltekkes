package middleware

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/integration/entity"
	"poltekkes-cat-backend/internal/modules/integration/repository"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func APIKeyAuthMiddleware(repo repository.IntegrationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		method := c.Request.Method
		endpoint := c.Request.URL.Path

		apiKey := strings.TrimSpace(c.GetHeader("X-API-Key"))
		if apiKey == "" {
			authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.Query("api_key"))
		}

		if apiKey == "" {
			errMsg := "Header 'X-API-Key' atau 'Authorization: Bearer <token>' wajib disertakan"
			go logAccess(repo, nil, "Unknown / Unauthenticated", clientIP, method, endpoint, 401, int(time.Since(startTime).Milliseconds()), userAgent, &errMsg)
			response.Unauthorized(c, "Otentikasi gagal: "+errMsg)
			c.Abort()
			return
		}

		keyRecord, err := repo.GetAPIKeyBySecret(c.Request.Context(), apiKey)
		if err != nil || keyRecord == nil {
			errMsg := "API Key tidak valid atau telah dinonaktifkan"
			go logAccess(repo, nil, "Invalid API Key", clientIP, method, endpoint, 401, int(time.Since(startTime).Milliseconds()), userAgent, &errMsg)
			response.Unauthorized(c, "Otentikasi gagal: "+errMsg)
			c.Abort()
			return
		}

		// Check IP Whitelist (Optional)
		if keyRecord.IPWhitelist != nil && strings.TrimSpace(*keyRecord.IPWhitelist) != "" {
			if !isIPAllowed(clientIP, *keyRecord.IPWhitelist) {
				errMsg := fmt.Sprintf("Akses ditolak: IP Anda (%s) tidak terdaftar dalam whitelist API Key '%s'", clientIP, keyRecord.Name)
				go logAccess(repo, &keyRecord.ID, keyRecord.Name, clientIP, method, endpoint, 403, int(time.Since(startTime).Milliseconds()), userAgent, &errMsg)
				response.Forbidden(c, errMsg)
				c.Abort()
				return
			}
		}

		// Async update last_used_at
		go func(id int) {
			_ = repo.UpdateLastUsed(context.Background(), id)
		}(keyRecord.ID)

		c.Set("api_key_id", keyRecord.ID)
		c.Set("api_key_name", keyRecord.Name)

		c.Next()

		// Record response log asynchronously
		durationMS := int(time.Since(startTime).Milliseconds())
		statusCode := c.Writer.Status()
		var errMsg *string
		if statusCode >= 400 {
			msg := fmt.Sprintf("HTTP %d returned", statusCode)
			errMsg = &msg
		}
		go logAccess(repo, &keyRecord.ID, keyRecord.Name, clientIP, method, endpoint, statusCode, durationMS, userAgent, errMsg)
	}
}

func logAccess(repo repository.IntegrationRepository, apiKeyID *int, clientName, ip, method, endpoint string, statusCode, durationMS int, userAgent string, errMsg *string) {
	_ = repo.CreateAccessLog(context.Background(), &entity.APIAccessLog{
		APIKeyID:       apiKeyID,
		ClientName:     clientName,
		IPAddress:      ip,
		Method:         method,
		Endpoint:       endpoint,
		StatusCode:     statusCode,
		ResponseTimeMS: durationMS,
		UserAgent:      &userAgent,
		ErrorMessage:   errMsg,
	})
}

func isIPAllowed(clientIP string, whitelist string) bool {
	parsedClientIP := net.ParseIP(strings.TrimSpace(clientIP))
	if parsedClientIP == nil {
		return false
	}

	parts := strings.FieldsFunc(whitelist, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == ' '
	})

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Direct IP match
		if p == clientIP {
			return true
		}

		// Handle localhost aliases (::1 vs 127.0.0.1)
		if (p == "127.0.0.1" || p == "localhost" || p == "::1") &&
			(clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "localhost") {
			return true
		}

		// CIDR range match
		if strings.Contains(p, "/") {
			_, ipNet, err := net.ParseCIDR(p)
			if err == nil && ipNet.Contains(parsedClientIP) {
				return true
			}
		} else {
			targetIP := net.ParseIP(p)
			if targetIP != nil && targetIP.Equal(parsedClientIP) {
				return true
			}
		}
	}

	return false
}
