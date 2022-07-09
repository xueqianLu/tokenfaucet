package models

type Param struct {
	Network  string `json:"network"`
	From     string `json:"from"`
	To       string `json:"to"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Data     string `json:"data"`
	Value    string `json:"value"`
}
