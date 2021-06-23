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

package callback

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gitlab.com/moneropay/go-monero/walletrpc"

	"gitlab.com/moneropay/moneropay/internal/moneropayd/database"
	"gitlab.com/moneropay/moneropay/internal/moneropayd/wallet"
)

func callback(url string, transfer []byte) error {
	b := bytes.NewBuffer(transfer)
	c := &http.Client{Timeout: time.Second * 3}
	req, _ := http.NewRequest("POST", url, b)
	req.Header.Set("Content-Type", "application/json")
	if _, err := c.Do(req); err != nil {
		return err
	}
	return nil
}

/*
 * Get the failed callbacks every minute.
 * Retry callback and if successful this time, remove it.
 * Else reschedule it and increment attempts.
 */
func retryFailed() {
	db := database.DB
	var url, body, interval string
	var attempts, id int
	for {
		rows, err := db.Query(context.Background(),
		    "SELECT callback_url, id, request_body, attempts FROM receivers, failed_callbacks" +
		    " WHERE CURRENT_TIMESTAMP > next_retry")
		if err != nil {
			log.Println("[retry callback]", err.Error())
			return
		}
		for rows.Next() {
			rows.Scan(&url, &id, &body, &attempts)
			if err = callback(url, []byte(body)); err != nil {
				switch(attempts) {
				case 1:
					interval = "5"
				case 2:
					interval = "10"
				case 3:
					interval = "30"
				case 4:
					interval = "60"
				}
				db.Exec(context.Background(), "UPDATE failed_callbacks SET attempts = attempts + 1," +
				    "next_retry = CURRENT_TIMESTAMP + interval '" + interval + " min' WHERE id = $1", id)
			} else {
				db.Exec(context.Background(), "DELETE FROM failed_callbacks WHERE id = $1", id)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

/*
 * Fetch new transfers from wallet-rpc and match with database.
 * If there is a match do callback.
 * If callback fails store the callback data in the failed_callbacks table.
 */
func poll() {
	var h uint64
	db := database.DB
	if err := db.QueryRow(context.Background(),
	    "SELECT last_polled_block FROM metadata").Scan(&h); err != nil {
		log.Println("[callback]", err.Error())
		return
	}
	rpc := wallet.Wallet
	resp, err := rpc.GetTransfers(&walletrpc.GetTransfersRequest{
		In: true,
		FilterByHeight: true,
		MinHeight: h,
	})
	a := []uint64{}
	for _, t := range resp.In {
		a = append(a, t.SubaddrIndex.Minor)
	}
	rows, err := db.Query(context.Background(), "SELECT subaddress_index, callback_url FROM receivers WHERE subaddress_index = ANY ($1)", a)
	var i uint64
	var u string
	urls := make(map[uint64]string)
	for rows.Next() {
		if err = rows.Scan(&i, &u); err != nil {
			continue
		}
		urls[i] = u
	}
	for _, t := range resp.In {
		url, ok := urls[t.SubaddrIndex.Minor]
		if !ok {
			continue
		}
		j, _ := json.Marshal(t)
		log.Printf("[callback] callback url:%s data: %v", url, t)
		if err = callback(url, j); err != nil {
			log.Println("[callback] inserting to failed_callbacks")
			db.Exec(context.Background(),
			    "INSERT INTO failed_callbacks (subaddress_index, request_body, next_retry)" +
			    "VALUES ($1, $2, CURRENT_TIMESTAMP + interval '1 min')",
			     t.SubaddrIndex.Minor, string(j))
		}
		if t.Height > h {
			h = t.Height
		}
	}
	_, err = db.Exec(context.Background(), "UPDATE metadata SET last_polled_block=$1", h)
	if err != nil {
		log.Println("[callback]", err.Error())
	}
}

func Run() {
	go retryFailed()
	for {
		poll()
		time.Sleep(1 * time.Minute)
	}
}
