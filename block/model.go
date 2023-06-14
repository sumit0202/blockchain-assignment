package block

import (
	"time"
)

type Transaction struct {
	TxnId   string
	Value   []Entry
	IsValid bool
	Data    map[string]Entry
}

type Entry struct {
	Val   float64 `json:"val"`
	Ver   float64 `json:"ver"`
	Valid bool    `json:"valid"`
	Hash  string  `json:"hash"`
}

type Block interface {
	PushValidTxns(txns []Transaction)
	UpdateBlockStatus(status BlockStatus)
	GetPrevBlockHash() string
}

type MyBlock struct {
	BlockNumber     int
	PrevBlockHash   string
	CurrBlockHash	string
	Status          BlockStatus
	ProcessingTime  time.Duration
	Txns            map[string]Transaction
}