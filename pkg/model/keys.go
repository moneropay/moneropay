/*
 * MoneroPay is a Monero payment processor.
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

package model

type KeysResponse struct {
	Mnemonic        string `json:"mnemonic"`
	RestoreHeight   uint64 `json:"restore_height"`
	SpendKeyPublic  string `json:"spend_key_public"`
	SpendKeyPrivate string `json:"spend_key_private"`
	ViewKeyPublic   string `json:"view_key_public"`
	ViewKeyPrivate  string `json:"view_key_private"`
}
