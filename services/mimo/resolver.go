// Copyright (c) 2019 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package mimo

import (
	"context"
	"math/big"

	"github.com/pkg/errors"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.
const (
	// HexPrefix is the prefix of ERC20 address in hex string
	HexPrefix = "0x"
	// DefaultPageSize is the size of page when pagination parameters are not set
	DefaultPageSize = 20
	// MaximumPageSize is the maximum size of page
	MaximumPageSize = 256
)

var (
	// ErrPaginationNotFound is the error indicating that pagination is not specified
	ErrPaginationNotFound = errors.New("pagination information is not found")
	// ErrPaginationInvalidOffset is the error indicating that pagination's offset parameter is invalid
	ErrPaginationInvalidOffset = errors.New("invalid pagination offset number")
	// ErrPaginationInvalidSize is the error indicating that pagination's size parameter is invalid
	ErrPaginationInvalidSize = errors.New("invalid pagination size number")
	// ErrInvalidParameter is the error indicating that invalid size
	ErrInvalidParameter = errors.New("invalid parameter number")
)

type queryResolver struct {
	service *mimoService
}

// Exchanges returns all exchanges
func (r *queryResolver) Exchanges(ctx context.Context, height string, pagination Pagination) ([]*Exchange, error) {
	if pagination.Skip < 0 {
		return nil, ErrPaginationInvalidOffset
	}
	if pagination.First <= 0 || pagination.First > MaximumPageSize {
		return nil, ErrPaginationInvalidSize
	}
	h, ok := new(big.Int).SetString(height, 10)
	if !ok {
		return nil, errors.Errorf("failed to parse height %s", height)
	}
	pairs, err := r.service.exchanges(h.Uint64(), uint32(pagination.Skip), uint8(pagination.First))
	if err != nil {
		return nil, err
	}
	exchanges := make([]string, len(pairs))
	tokens := make([]string, len(pairs))
	reversePairs := make([]AddressPair, len(pairs))
	for i, pair := range pairs {
		exchanges[i] = pair.Address1
		tokens[i] = pair.Address2
		reversePairs[i] = AddressPair{
			Address1: pair.Address2,
			Address2: pair.Address1,
		}
	}
	balances, err := r.service.balances(h.Uint64(), exchanges)
	if err != nil {
		return nil, err
	}
	supplies, err := r.service.supplies(h.Uint64(), exchanges)
	if err != nil {
		return nil, err
	}
	tokenInfos, err := r.service.tokens(h.Uint64(), tokens)
	if err != nil {
		return nil, err
	}
	tokenBalances, err := r.service.tokenBalances(h.Uint64(), reversePairs)
	if err != nil {
		return nil, err
	}
	volumes, err := r.service.volumesInPast24Hours(exchanges)
	if err != nil {
		return nil, err
	}
	ret := make([]*Exchange, 0, len(exchanges))
	for _, pair := range reversePairs {
		token := pair.Address1
		exchange := pair.Address2
		balance, ok := balances[exchange]
		if !ok {
			balance = big.NewInt(0)
		}
		tokenBalance, ok := tokenBalances[pair]
		if !ok {
			tokenBalance = big.NewInt(0)
		}
		supply, ok := supplies[exchange]
		if !ok {
			supply = big.NewInt(0)
		}
		volume, ok := volumes[exchange]
		if !ok {
			volume = big.NewInt(0)
		}
		info, ok := tokenInfos[token]
		if !ok {
			info = Token{Address: token}
		}
		ret = append(ret, &Exchange{
			Address:             exchange,
			Token:               info,
			VolumeInPast24Hours: volume.String(),
			Liquidity:           supply.String(),
			BalanceOfIotx:       balance.String(),
			BalanceOfToken:      tokenBalance.String(),
		})
	}
	return ret, nil
}

func (r *queryResolver) TipHeight(ctx context.Context) (string, error) {
	tip, err := r.service.latestHeight()
	if err != nil {
		return "", err
	}
	return tip.String(), nil
}

func (r *queryResolver) TotalVolumes(ctx context.Context, days int) ([]*Volume, error) {
	if days < 0 {
		days = 30
	}
	if days > 256 {
		days = 256
	}
	volumes, err := r.service.totalVolumes(uint8(days))
	if err != nil {
		return nil, err
	}
	ret := []*Volume{}
	for date, volume := range volumes {
		ret = append(ret, &Volume{
			Amount: volume.String(),
			Date:   date,
		})
	}
	return ret, nil
}
