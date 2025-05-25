package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/centrifugal/centrifuge"
	"github.com/dgrijalva/jwt-go"
)

func ValidateAuthWS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenKey := []byte(os.Getenv("CENTRIFUGO_TOKEN_HMAC_SECRET_KEY"))

		cookie, err := r.Cookie("access_token")
		if err != nil {
			fmt.Println("Unauthorized: no token")
			http.Error(w, "Unauthorized: no token", http.StatusUnauthorized)
			return
		}
		tokenString := cookie.Value

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			method, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok || method.Alg() != "HS256" {
				return nil, errors.New("bad sign method")
			}
			return tokenKey, nil
		})
		if err != nil || !token.Valid {
			fmt.Println("Token is invalid")
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("User claims type cast err")
			http.Error(w, "User claims type cast err", http.StatusUnauthorized)
			return
		}

		claimsUser, ok := claims["user"].(map[string]interface{})
		if !ok {
			fmt.Println("User claims are missing")
			http.Error(w, "User claims are missing", http.StatusUnauthorized)
			return
		}

		userIDString, ok := claimsUser["id"].(string)
		if !ok {
			fmt.Println("Couldn't parse user id")
			http.Error(w, "Couldn't parse user id", http.StatusUnauthorized)
			return
		}

		cred := &centrifuge.Credentials{
			UserID: userIDString,
		}

		ctx := r.Context()
		newCtx := centrifuge.SetCredentials(ctx, cred)
		r = r.WithContext(newCtx)

		fmt.Println("WebSocket connection, passing UserID to Centrifugo:", userIDString)

		h.ServeHTTP(w, r)
	})
}
