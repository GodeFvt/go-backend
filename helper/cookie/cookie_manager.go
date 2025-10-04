package cookie

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type CookieManager struct {
	domain      string
	secure      bool
	sameSite    http.SameSite
	httpOnly    bool
	accessPath  string
	refreshPath string
}

func NewCookieManager(domain string, secure, crossSite bool) *CookieManager {
	sameSite := http.SameSiteLaxMode
	if crossSite && secure {
		sameSite = http.SameSiteNoneMode
	}
	return &CookieManager{
		domain:      domain,   // เว้นว่าง "" ใน dev
		secure:      secure,   // dev=false, prod=true
		sameSite:    sameSite, // dev=Lax, prod None/Lax ตาม topology
		httpOnly:    true,
		accessPath:  "/",
		refreshPath: "/",
	}
}

func (cm *CookieManager) SetCookie(c echo.Context, name, value string, expiresAt time.Time, path string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   cm.domain,
		Expires:  expiresAt,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(cookie)
}

func (cm *CookieManager) ClearCookie(c echo.Context, name, path string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   cm.domain,
		MaxAge:   -1,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(cookie)
}

func (cm *CookieManager) SetAccessTokenCookie(c echo.Context, token string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     cm.accessPath,
		Domain:   cm.domain,
		Expires:  expiresAt,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(cookie)
}

func (cm *CookieManager) SetRefreshTokenCookie(c echo.Context, token string, expiresAt time.Time) {
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     cm.refreshPath,
		Domain:   cm.domain,
		Expires:  expiresAt,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(refreshCookie)
}

func (cm *CookieManager) SetPKCEVerifierCookie(c echo.Context, verifier string, expiresAt time.Time) {
	verifierCookie := &http.Cookie{
		Name:     "pkce_verifier",
		Value:    verifier,
		Path:     "/",
		Domain:   cm.domain,
		Expires:  expiresAt,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(verifierCookie)
}

func (cm *CookieManager) SetSessionCookie(c echo.Context, sessionID string, expiresAt time.Time) {
	sessionCookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Domain:   cm.domain,
		Expires:  expiresAt,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(sessionCookie)
}

func (cm *CookieManager) GetAccessToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (cm *CookieManager) GetRefreshToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (cm *CookieManager) GetPKCEVerifier(c echo.Context) (string, error) {
	cookie, err := c.Cookie("pkce_verifier")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (cm *CookieManager) GetSessionID(c echo.Context) (string, error) {
	cookie, err := c.Cookie("session_id")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (cm *CookieManager) ClearAllJwtCookies(c echo.Context) {
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     cm.accessPath,
		Domain:   cm.domain,
		MaxAge:   -1,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(accessCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     cm.refreshPath,
		Domain:   cm.domain,
		MaxAge:   -1,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(refreshCookie)

	verifierCookie := &http.Cookie{
		Name:     "pkce_verifier",
		Value:    "",
		Path:     "/",
		Domain:   cm.domain,
		MaxAge:   -1,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(verifierCookie)

	sessionCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Domain:   cm.domain,
		MaxAge:   -1,
		Secure:   cm.secure,
		HttpOnly: cm.httpOnly,
		SameSite: cm.sameSite,
	}
	c.SetCookie(sessionCookie)
}
