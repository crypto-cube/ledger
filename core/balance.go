package core

type Balance struct {
	Balance  int64     `json:"balance"`
	Account  string    `json:"account"`
	Asset    string    `json:"asset"`
	Children []Balance `json:"children"`
}
