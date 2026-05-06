package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.codycody31.dev/squad-aegis/internal/core"
	"go.codycody31.dev/squad-aegis/internal/models"
	"go.codycody31.dev/squad-aegis/internal/server/responses"
)

// AuthSession checks if the user is authenticated, if not it returns a 401 Unauthorized response
func (s *Server) AuthSession(c *gin.Context) {
	s.authSession(c, true)
}

// OptionalAuthSession checks if the user is authenticated, but does not require it
func (s *Server) OptionalAuthSession(c *gin.Context) {
	s.authSession(c, false)
}

// authSession checks if the user is authenticated and updates the session's last seen time and IP address
func (s *Server) authSession(c *gin.Context, required bool) {
	sessionToken := getSessionToken(c)
	session := &models.Session{}

	if sessionToken == "" {
		if required {
			responses.Unauthorized(c, "Unauthorized", nil)
			return
		}

		c.Next()
		return
	}

	dests := []any{&session.Id, &session.UserId, &session.Token, &session.CreatedAt, &session.ExpiresAt, &session.LastSeen, &session.LastSeenIp}

	// Check if the session token provided is valid
	tokenHash := core.HashSessionToken(sessionToken)
	row := s.Dependencies.DB.QueryRow("SELECT * FROM sessions WHERE token = $1 AND (expires_at IS NULL OR expires_at > NOW())", tokenHash)
	if err := row.Scan(dests...); err != nil {
		if required {
			responses.Unauthorized(c, "Unauthorized", nil)
			return
		} else {
			c.Next()
			return
		}
	}

	// Update the last seen time of the session and the IP address
	_, err := s.Dependencies.DB.Exec("UPDATE sessions SET last_seen = NOW(), last_seen_ip = $1 WHERE id = $2", c.ClientIP(), session.Id)
	if err != nil {
		responses.InternalServerError(c, err, nil)
		return
	}

	c.Set("session", session)
}

func getSessionToken(c *gin.Context) string {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := cleanSessionToken(strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer ")))
		if token != "" {
			return token
		}
	}

	cookieNames := []string{"session", "squad_aegis_session", "session_token"}
	for _, cookieName := range cookieNames {
		cookie, err := c.Cookie(cookieName)
		if err == nil {
			if token := cleanSessionToken(cookie); token != "" {
				return token
			}
		}
	}

	return ""
}

func cleanSessionToken(token string) string {
	token = strings.TrimSpace(token)
	switch strings.ToLower(token) {
	case "", "undefined", "null":
		return ""
	default:
		return token
	}
}

// AuthIsSuperAdmin checks if the user is a super admin, if not it returns a 403 Forbidden response
func (s *Server) AuthIsSuperAdmin() func(c *gin.Context) {
	return func(c *gin.Context) {
		sess, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}
		session := sess.(*models.Session)

		userIsSuperAdmin := s.Dependencies.DB.QueryRow("SELECT FROM users WHERE id = $1 AND super_admin = true", session.UserId)
		if err := userIsSuperAdmin.Scan(); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Forbidden",
				"code":    http.StatusForbidden,
			})
			c.Abort()
			return
		}
	}
}

func IsLoggedIn(c *gin.Context) bool {
	_, exists := c.Get("session")
	return exists
}
