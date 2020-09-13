// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Exchange struct {
	Token          string `json:"token"`
	Liquidity      string `json:"liquidity"`
	BalanceOfToken string `json:"balanceOfToken"`
	BalanceOfIotx  string `json:"balanceOfIOTX"`
}

type Pagination struct {
	Skip  int `json:"skip"`
	First int `json:"first"`
}
