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
	"strconv"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
)

func addAddressRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/address")

	g.GET("/:subaddress_id", func(c *gin.Context) {
		s := c.Param("subaddress_id")
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		resp, err := w.GetTransfers(&wallet.RequestGetTransfers{
			SubaddrIndices: []uint64{i},
			In: true,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		type transferData struct {
			Address string `json:"address"`
			Amount float64 `json:"amount"`
			Confirmations uint64 `json:"confirmations"`
			DoubleSpendSeen bool `json:"double_spend_seen"`
			Fee float64 `json:"fee"`
			Timestamp uint64 `json:"timestamp"`
			TxID string `json:"txid"`
			UnlockTime uint64 `json:"unlock_time"`
		}
		var data []*transferData
		for _, tr := range resp.In {
			td := &transferData{
				Address: tr.Address,
				Amount: wallet.XMRToFloat64(tr.Amount),
				Confirmations: tr.Confirmations,
				DoubleSpendSeen: tr.DoubleSpendSeen,
				Fee: wallet.XMRToFloat64(tr.Fee),
				Timestamp: tr.Timestamp,
				TxID: tr.TxID,
				UnlockTime: tr.UnlockTime,
			}
			data = append(data, td)
		}
		c.JSON(http.StatusOK, gin.H{
			"transfers": data,
		})
	})

	g.POST("/", func(c *gin.Context) {
		resp, err := w.CreateAddress(&wallet.RequestCreateAddress{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"address": resp.Address,
			"index": resp.AddressIndex,
		})
	})
}
