package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0xMishra/makerble/internal/data"
	"github.com/0xMishra/makerble/internal/validator"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}

			next.ServeHTTP(w, r)
		}()
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Importantly, unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	// runs on every request
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// Extract the client's IP address from the request.
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)

				return
			}

			// Lock the mutex to prevent this code from being executed concurrently.
			mu.Lock()

			// Check to see if the IP address already exists in the map. If it doesn't, then
			// initialize a new rate limiter and add the IP address and limiter to the map.
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}

			clients[ip].lastSeen = time.Now()
			// Call the Allow() method on the rate limiter for the current IP address. If
			// the request isn't allowed, unlock the mutex and send a 429 Too Many Requests
			// response, just like before.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			// Very importantly, unlock the mutex before calling the next handler in the
			// chain. Notice that we DON'T use defer to unlock the mutex, as that would mean
			// that the mutex isn't unlocked until all the handlers downstream of this
			// middleware have also returned.
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

/*
If your code makes a decision about what to return based on the content of a request header,
you should include that header name in your Vary response header — even if the request
didn’t include that header.
*/
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")

		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(role1, role2 string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			bearerToken := r.Header.Get("Authorization")
			sliceToken := strings.Split(bearerToken, " ")
			token := sliceToken[1]

			v := validator.New()

			if sliceToken[0] != "Bearer" {
				app.invalidAuthenticationTokenResponse(w, r)
				return
			}

			if data.ValidateTokenPlaintext(v, token); !v.Valid() {
				app.failedValidationResponse(w, r, v.Errors)
				return
			}

			var t *data.Token
			t, err := app.models.Tokens.GetUserForToken(token)
			if err != nil {
				switch {
				case errors.Is(err, data.ErrRecordNotFound):
					app.invalidAuthenticationTokenResponse(w, r)

				default:
					app.serverErrorResponse(w, r, err)
				}
				return
			}

			if t.Role != role1 && t.Role != role2 {
				app.invalidCredentialsResponse(w, r)
				return
			}

			currentTime := time.Now()

			if currentTime.After(t.Expiry) {
				err = app.models.Tokens.DeleteAllForUser(data.ScopeAuthentication, t.Email)
				if err != nil {
					app.serverErrorResponse(w, r, err)
					return
				}
			}

			next.ServeHTTP(w, r)
		},
	)
}
