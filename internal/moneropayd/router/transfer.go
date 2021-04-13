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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"

	"gitlab.com/kernal/moneropay/internal/moneropayd/config"
)

func addTransferRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/transfer")

	g.POST("/", func(c *gin.Context) {
		if c.ContentType() != "application/json" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bad content-type",
			})
			return
		}
		type Destination struct {
			Amount float64 `json:"amount"`
			Address string `json:"address"`
		}
		var d0 []*Destination
		var d1 []*wallet.Destination
		err := c.BindJSON(&d0)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		for _, i := range d0 {
			d :=  wallet.Destination{
				Amount: wallet.Float64ToXMR(i.Amount),
				Address: i.Address,
			}
			d1 = append(d1, &d)
		}
		resp, err := w.Transfer(&wallet.RequestTransfer{
			Destinations: d1,
			Priority: wallet.Priority(config.Values.TransferPriority),
			Mixing: config.Values.TransferMixin,
			UnlockTime: config.Values.TransferUnlockTime,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"amount": wallet.XMRToFloat64(resp.Amount),
			"fee": wallet.XMRToFloat64(resp.Fee),
			"tx_hash": resp.TxHash,
		})
	})
}
