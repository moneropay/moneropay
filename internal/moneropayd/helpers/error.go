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

package helpers

import (
	"encoding/json"
	"net/http"

	"gitlab.com/moneropay/moneropay/pkg/models"
)

func WriteError(w http.ResponseWriter, status int, code *int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Status: status,
		Code: code,
		Message: message,
	})
}
