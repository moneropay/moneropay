/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
 *
 * MoneroPay is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * MoneroPay is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with MoneroPay.  If not, see <https://www.gnu.org/licenses/>.
 */

package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"gitlab.com/moneropay/moneropay/v2/internal/daemon"
	"gitlab.com/moneropay/moneropay/v2/internal/server/controller"
	"gitlab.com/moneropay/moneropay/v2/pkg/model"
)

// middlewareKeysPending blocks all endpoints except /keys and /health
// when in view-only init mode and keys haven't been retrieved yet.
func middlewareKeysPending(d *daemon.Daemon) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if d.IsKeysPending() && path != "/keys" && path != "/health" {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(model.ErrorResponse{
					Status:  http.StatusServiceUnavailable,
					Message: "Retrieve keys via GET /keys first",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// initRouter creates and configures the HTTP router with all routes.
func initRouter(d *daemon.Daemon) *chi.Mux {
	ctrl := controller.New(d)

	r := chi.NewRouter()
	r.Use(middlewareContentType)
	r.Use(middlewareServerHeader(d))
	r.Use(middlewareXMoneroPayAddressHeader(d))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(100 * time.Second))
	r.Use(middlewareKeysPending(d))

	r.Get("/health", ctrl.HealthHandler)
	r.Get("/balance", ctrl.BalanceHandler)
	r.Get("/keys", ctrl.KeysGetHandler)
	r.Post("/receive", ctrl.ReceivePostHandler)
	r.Get("/receive/{address}", ctrl.ReceiveGetHandler)
	r.Delete("/receive/{address}", ctrl.ReceiveDeleteHandler)
	r.Post("/transfer", ctrl.TransferPostHandler)
	r.Get("/transfer/{tx_hash}", ctrl.TransferGetHandler)

	return r
}

// Run starts the HTTP server with the given daemon.
func Run(d *daemon.Daemon) {
	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:         d.Config().BindAddr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		Handler:      h2c.NewHandler(initRouter(d), h2s),
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		d.Shutdown()
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out. Forcing exit.")
			}
		}()
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-serverCtx.Done()
}
