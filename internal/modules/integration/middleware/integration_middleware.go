package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	"poltekkes-cat-backend/internal/modules/integration/entity"
	"poltekkes-cat-backend/internal/modules/integration/repository"
	"poltekkes-cat-backend/internal/shared/response"

	"github.com/gin-gonic/gin"
)

var passwordMaskRegex = regexp.MustCompile(`(?i)("password"\s*:\s*)"([^"]+)"`)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func maskSensitiveData(raw string) string {
	if raw == "" {
		return ""
	}
	return passwordMaskRegex.ReplaceAllString(raw, `$1"********"`)
}

func truncateBody(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "\n... [truncated]"
	}
	return s
}

func strPtrOrNil(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func APIKeyAuthMiddleware(repo repository.IntegrationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		method := c.Request.Method
		endpoint := c.Request.URL.Path

		// Wrap ResponseWriter to capture response payload
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Read and preserve Request Body
		var rawReqBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				rawReqBody = truncateBody(maskSensitiveData(string(bodyBytes)), 65536)
			}
		}

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
			response.Unauthorized(c, "Otentikasi gagal: "+errMsg)
			c.Abort()

			durationMS := int(time.Since(startTime).Milliseconds())
			respBody := truncateBody(blw.body.String(), 65536)
			go logAccess(repo, nil, "Unknown / Unauthenticated", clientIP, method, endpoint, 401, durationMS, userAgent, &errMsg, strPtrOrNil(rawReqBody), strPtrOrNil(respBody))
			return
		}

		keyRecord, err := repo.GetAPIKeyBySecret(c.Request.Context(), apiKey)
		if err != nil || keyRecord == nil {
			errMsg := "API Key tidak valid atau telah dinonaktifkan"
			response.Unauthorized(c, "Otentikasi gagal: "+errMsg)
			c.Abort()

			durationMS := int(time.Since(startTime).Milliseconds())
			respBody := truncateBody(blw.body.String(), 65536)
			go logAccess(repo, nil, "Invalid API Key", clientIP, method, endpoint, 401, durationMS, userAgent, &errMsg, strPtrOrNil(rawReqBody), strPtrOrNil(respBody))
			return
		}

		// Check IP Whitelist (Optional)
		if keyRecord.IPWhitelist != nil && strings.TrimSpace(*keyRecord.IPWhitelist) != "" {
			if !isIPAllowed(clientIP, *keyRecord.IPWhitelist) {
				errMsg := fmt.Sprintf("Akses ditolak: IP Anda (%s) tidak terdaftar dalam whitelist API Key '%s'", clientIP, keyRecord.Name)
				response.Forbidden(c, errMsg)
				c.Abort()

				durationMS := int(time.Since(startTime).Milliseconds())
				respBody := truncateBody(blw.body.String(), 65536)
				go logAccess(repo, &keyRecord.ID, keyRecord.Name, clientIP, method, endpoint, 403, durationMS, userAgent, &errMsg, strPtrOrNil(rawReqBody), strPtrOrNil(respBody))
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
		respBody := truncateBody(blw.body.String(), 65536)
		go logAccess(repo, &keyRecord.ID, keyRecord.Name, clientIP, method, endpoint, statusCode, durationMS, userAgent, errMsg, strPtrOrNil(rawReqBody), strPtrOrNil(respBody))
	}
}

func logAccess(repo repository.IntegrationRepository, apiKeyID *int, clientName, ip, method, endpoint string, statusCode, durationMS int, userAgent string, errMsg *string, reqBody, respBody *string) {
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
		RequestBody:    reqBody,
		ResponseBody:   respBody,
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
