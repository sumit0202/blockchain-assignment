package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
    "os"
	"github.com/syndtr/goleveldb/leveldb"
)

type BlockStatus int

const (
	Committed BlockStatus = iota
	Pending
)

type Txn struct {
	Key   string `json:"key"`
	Value Value  `json:"value"`
	Valid bool   `json:"valid"`
	Hash  string `json:"hash"`
}

type Value struct {
	Val int     `json:"val"`
	Ver float64 `json:"ver"`
}

type Block struct {
	BlockNumber    int         `json:"blockNumber"`
	PrevBlockHash  string      `json:"prevBlockHash"`
	Txns           []Txn       `json:"txns"`
	Timestamp      int64       `json:"timestamp"`
	BlockStatus    BlockStatus `json:"blockStatus"`
}

type BlockInterface interface {
	PushTxns(txns []Txn) error
	UpdateBlockStatus(status BlockStatus) error
}

type BlockImpl struct {
	db *leveldb.DB
}

func NewBlockImpl(db *leveldb.DB) *BlockImpl {
	return &BlockImpl{db: db}
}

func (b *BlockImpl) PushTxns(txns []Txn) error {
	var wg sync.WaitGroup
	for i := range txns {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			hash := sha256.Sum256([]byte(fmt.Sprintf("%v", txns[i])))
			txns[i].Hash = fmt.Sprintf("%x", hash)
			if val, err := b.db.Get([]byte(txns[i].Key), nil); err == nil {
				var value Value
				if err := json.Unmarshal(val, &value); err == nil {
					if value.Ver == txns[i].Value.Ver {
						txns[i].Valid = true
						b.db.Put([]byte(txns[i].Key), []byte(fmt.Sprintf(`{"val":%d,"ver":%f}`, txns[i].Value.Val, txns[i].Value.Ver)), nil)
					} else {
                        txns[i].Valid = false
                    }
				}
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func (b *BlockImpl) UpdateBlockStatus(status BlockStatus) error {
	return nil
}

func writeBlockToFile(blockChannel chan Block) {
	file, err := os.OpenFile("./db/ledger.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for block := range blockChannel {
		blockJSON, err := json.Marshal(block)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := file.WriteString(string(blockJSON) + "\n"); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	db, err := leveldb.OpenFile("./db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	blockImpl := NewBlockImpl(db)

    blockChannel := make(chan Block)
	go writeBlockToFile(blockChannel)

	txns := []Txn{
		{Key: "SIM1", Value: Value{Val: 2, Ver: 1.0}},
		{Key: "SIM2", Value: Value{Val: 3, Ver: 1.0}},
		{Key: "SIM3", Value: Value{Val: 4, Ver: 2.0}},
	}

	block := Block{
		BlockNumber:   1,
		PrevBlockHash: "0xabc123",
		Txns:          txns,
		Timestamp:     time.Now().Unix(),
		BlockStatus:   Pending,
	}

    blockChannel <- block

	if err := blockImpl.PushTxns(block.Txns); err != nil {
		log.Fatal(err)
	}

    //blockChannel <- block

	blockJSON, err := json.Marshal(block)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(blockJSON))
}
