/*
 * Copyright (C) 2021 Laurynas Četyrkinas <stnby@kernal.eu>
 * Copyright (C) 2021 İrem Kuyucu <siren@kernal.eu>
 *
 * This file is part of MoneroPay.
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

package router

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"gitlab.com/moneropay/moneropay/internal/moneropayd/controllers"
)

func Run(bind string) {
	r := mux.NewRouter()
	r.HandleFunc("/health", controllers.HealthHandler).Methods("GET", "HEAD")
	r.HandleFunc("/balance", controllers.BalanceHandler).Methods("GET")
	r.HandleFunc("/receive", controllers.ReceivePostHandler).Methods("POST")
	r.HandleFunc("/receive/{address}", controllers.ReceiveGetHandler).Methods("GET")
	r.HandleFunc("/transfer", controllers.TransferPostHandler).Methods("POST").Headers(
	    "Content-Type", "application/json")
	r.HandleFunc("/transfer/{tx_hash}", controllers.TransferGetHandler).Methods("GET")
	srv := &http.Server{
		Handler: r,
		Addr: bind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
