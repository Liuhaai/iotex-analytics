// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package mimo

type Exchange struct {
	Address             string `json:"address"`
	Token               Token  `json:"token"`
	Liquidity           string `json:"liquidity"`
	VolumeInPast24Hours string `json:"volumeInPast24Hours"`
	BalanceOfToken      string `json:"balanceOfToken"`
	BalanceOfIotx       string `json:"balanceOfIOTX"`
}

type Pagination struct {
	Skip  int `json:"skip"`
	First int `json:"first"`
}

type Token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type Volume struct {
	Amount string `json:"amount"`
	Date   string `json:"date"`
}
