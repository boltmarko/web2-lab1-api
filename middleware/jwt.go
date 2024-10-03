package middleware

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

func EnsureValidToken() func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
	)

	if err != nil {
		log.Fatalf("Failed to create the jwt validator: %v", err)
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered an error while validating the token: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_token"}`))
	}

	middleware := jwtmiddleware.New(jwtValidator.ValidateToken, jwtmiddleware.WithErrorHandler(errorHandler))

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}

}
