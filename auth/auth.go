package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/uh-kay/glimpze/env"
	"github.com/uh-kay/glimpze/store/cache"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")
)

type Tokens struct {
	Access   string
	Refresh  string
	JTIAcc   string
	JTIRef   string
	ExpAcc   time.Time
	ExpRef   time.Time
	UserID   string
	Issuer   string
	Audience string
}

type Claims struct {
	Email  string `json:"email"`
	UserID int64  `json:"user_id"`
	jwt.RegisteredClaims
}

func IssueTokens(userID string) (*Tokens, error) {
	now := time.Now().UTC()
	t := &Tokens{
		UserID:   userID,
		JTIAcc:   uuid.NewString(),
		JTIRef:   uuid.NewString(),
		ExpAcc:   now.Add(15 * time.Minute),
		ExpRef:   now.Add(7 * 24 * time.Hour),
		Issuer:   "glimpze-app",
		Audience: "glimpze-client",
	}

	acc := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		ID:        t.JTIAcc,
		Issuer:    t.Issuer,
		Audience:  jwt.ClaimStrings{t.Audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(t.ExpAcc),
	})

	ref := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		ID:        t.JTIRef,
		Issuer:    t.Issuer,
		Audience:  jwt.ClaimStrings{t.Audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(t.ExpRef),
	})

	var err error
	t.Access, err = acc.SignedString([]byte(env.GetString("ACCESS_SECRET", "change-this")))
	if err != nil {
		return nil, err
	}
	t.Refresh, err = ref.SignedString([]byte(env.GetString("REFRESH_SECRET", "change-this")))
	if err != nil {
		return nil, err
	}
	return t, nil
}

func Persist(ctx context.Context, v *cache.Storage, t *Tokens) error {
	if err := v.Sessions.Set(ctx, "access:"+t.JTIAcc, t.UserID, t.ExpAcc); err != nil {
		return err
	}
	if err := v.Sessions.Set(ctx, "refresh:"+t.JTIRef, t.UserID, t.ExpRef); err != nil {
		return err
	}
	return nil
}

func SetAuthCookies(w http.ResponseWriter, t *Tokens) {
	access_cookie := &http.Cookie{
		Name:     "access_token",
		Value:    t.Access,
		MaxAge:   int(time.Until(t.ExpAcc).Seconds()),
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	refresh_cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    t.Refresh,
		MaxAge:   int(time.Until(t.ExpAcc).Seconds()),
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, access_cookie)
	http.SetCookie(w, refresh_cookie)
}

func ClearAuthCookies(w http.ResponseWriter) {
	access_cookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	refresh_cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, access_cookie)
	http.SetCookie(w, refresh_cookie)
}

func ParseAccess(tokenStr string) (*jwt.RegisteredClaims, error) {
	secret := env.GetString("ACCESS_SECRET", "change-this")
	return parseWithSecret(tokenStr, secret)
}

func ParseRefresh(tokenStr string) (*jwt.RegisteredClaims, error) {
	secret := env.GetString("REFRESH_SECRET", "change-this")
	return parseWithSecret(tokenStr, secret)
}

func parseWithSecret(tokenStr, secret string) (*jwt.RegisteredClaims, error) {
	if secret == "" {
		return nil, errors.New("jwt secret not configured")
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	token, err := parser.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
