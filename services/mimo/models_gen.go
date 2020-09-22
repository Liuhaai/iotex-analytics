// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package mimo

type AmountInOneDay struct {
	Amount string `json:"amount"`
	Date   string `json:"date"`
}

type Exchange struct {
	Address             string `json:"address"`
	Token               Token  `json:"token"`
	Liquidity           string `json:"liquidity"`
	VolumeInPast24Hours string `json:"volumeInPast24Hours"`
	VolumeInPast7Days   string `json:"volumeInPast7Days"`
	BalanceOfToken      string `json:"balanceOfToken"`
	BalanceOfIotx       string `json:"balanceOfIOTX"`
}

type Pagination struct {
	Skip  int `json:"skip"`
	First int `json:"first"`
}

type Stats struct {
	NumOfTransations int    `json:"numOfTransations"`
	Volume           string `json:"volume"`
}

type Token struct {
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
}