/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2022 Laurynas Četyrkinas <stnby@kernal.eu>
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
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
)

func initRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewareServerHeader)
	r.Use(middlewareXMoneroPayAddressHeader)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(100 * time.Second))
	r.Get("/health", controller.HealthHandler)
	r.Get("/balance", controller.BalanceHandler)
	r.Post("/receive", controller.ReceivePostHandler)
	r.Get("/receive/{address}", controller.ReceiveGetHandler)
	r.Post("/transfer", controller.TransferPostHandler)
	r.Get("/transfer/{tx_hash}", controller.TransferGetHandler)
	return r
}

func Run() {
	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:         daemon.Config.BindAddr,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		Handler:      h2c.NewHandler(initRouter(), h2s),
	}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
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
