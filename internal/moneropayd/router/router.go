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

	v1 "gitlab.com/moneropay/moneropay/internal/moneropayd/v1/controllers"
)

func Run(bind string) {
	r := mux.NewRouter()
	s1 := r.PathPrefix("/v1").Subrouter()
	s1.HandleFunc("/health", v1.HealthHandler).Methods("GET", "HEAD")
	s1.HandleFunc("/balance", v1.BalanceHandler).Methods("GET")
	s1.HandleFunc("/receive", v1.ReceivePostHandler).Methods("POST")
	s1.HandleFunc("/receive/{address}", v1.ReceiveGetHandler).Methods("GET")
	s1.HandleFunc("/transfer", v1.TransferPostHandler).Methods("POST").Headers("Content-Type", "application/json")
	s1.HandleFunc("/transfer/{tx_hash}", v1.TransferGetHandler).Methods("GET")
	srv := &http.Server{
		Handler: r,
		Addr: bind,
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
