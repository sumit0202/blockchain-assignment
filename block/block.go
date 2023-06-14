package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"github.com/syndtr/goleveldb/leveldb"
)

func (b *MyBlock) GetPrevBlockHash() string {
	return b.PrevBlockHash
}

func (b *MyBlock) PushValidTransactions(db *leveldb.DB, InputTxns []map[string]Entry, txnChan chan Transaction, wg *sync.WaitGroup) {
	defer wg.Done()

	newdb := make(map[string]Entry)
	for _, txn := range InputTxns {
		for key, value := range txn {
			hashInput := fmt.Sprintf("%s:%s", b.GetPrevBlockHash(), key)
			hash := sha256.Sum256([]byte(hashInput))

			data, err := db.Get([]byte(key), nil)
			if err == nil {
				var entry Entry
				err = json.Unmarshal(data, &entry)
				if err == nil {
					if entry.Ver == value.Ver {
						entry.Val = value.Val
						entry.Ver++
						entry.Valid = true
					} else {
						entry.Valid = false
					}
					entry.Hash = fmt.Sprintf("%x", hash)
					newdb[key] = entry
				}
			}

			if _, ok := newdb[key]; !ok {
				value.Valid = false
				value.Hash = fmt.Sprintf("%x", hash)
				newdb[key] = value
			}
		}
	}

	transaction := Transaction{
		TxnId:   fmt.Sprintf("%x", sha256.Sum256([]byte(b.PrevBlockHash))),
		Data:    newdb,
		IsValid: true,
	}

	txnChan <- transaction
}

func (b *MyBlock) UpdateBlockStatus(status BlockStatus) {
	b.Status = status
}

type BlockWriter chan MyBlock

func (bw BlockWriter) WriteBlock(block MyBlock) {
	bw <- block
}

func WriteBlocksToFile(blocks []MyBlock) error {
	fileName := "blocks.json"

	blockJSON, err := json.Marshal(blocks)
	if err != nil {
		return fmt.Errorf("error marshaling blocks to JSON: %v", err)
	}

	err = ioutil.WriteFile(fileName, blockJSON, 0644)
	if err != nil {
		return fmt.Errorf("error writing blocks to file: %v", err)
	}

	return nil
}

func FetchBlockDetailsByNumber(blockNumber int) (MyBlock, error) {
	fileName := "blocks.json"

	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return MyBlock{}, err
	}

	var blocks []MyBlock
	err = json.Unmarshal(fileContent, &blocks)
	if err != nil {
		return MyBlock{}, err
	}

	for _, block := range blocks {
		if block.BlockNumber == blockNumber {
			return block, nil
		}
	}

	return MyBlock{}, fmt.Errorf("block with block number %d not found", blockNumber)
}

func FetchAllBlockDetails() ([]MyBlock, error) {
	fileName := "blocks.json"

	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var blocks []MyBlock
	err = json.Unmarshal(fileContent, &blocks)
	if err != nil {
		return nil, err
	}

	return blocks, nil
}

func PrintBlockDetails(blocks []MyBlock) {
	for _, b := range blocks {
		fmt.Println("Block Number:", b.BlockNumber)
		fmt.Println("Prev Block Hash:", b.PrevBlockHash)
		fmt.Println("Block Status:", b.Status)
		fmt.Println("Processing Time:", b.ProcessingTime)
		fmt.Println("Transactions:")
		for txnID, txn := range b.Txns {
			fmt.Println("  Transaction ID:", txnID)
			fmt.Println("  IsValid:", txn.IsValid)
			fmt.Println("  Data:")
			for key, entry := range txn.Data {
				fmt.Println("    Key:", key)
				fmt.Println("    Value:", entry.Val)
				fmt.Println("    Ver:", entry.Ver)
				fmt.Println("    Valid:", entry.Valid)
				fmt.Println("    Hash:", entry.Hash)
			}
		}
	}
}

func BlockDetails() error {
	for{
		var blockNumber int
		fmt.Print("Enter the block number: ")
		_, err := fmt.Scanln(&blockNumber)
		if err != nil {
			log.Fatal("Error reading block number:", err)
			return err
		}
		
		fetchedBlock, err := FetchBlockDetailsByNumber(blockNumber)
		if err != nil {
			fmt.Println("Error fetching block details:", err)
			return err
		}
		fmt.Println("Fetched Block Details:")
		fmt.Println("Block Number:", fetchedBlock.BlockNumber)
		fmt.Println("Prev Block Hash:", fetchedBlock.PrevBlockHash)
		fmt.Println("Curr Block Hash:", fetchedBlock.CurrBlockHash)
		fmt.Println("Block Status:", fetchedBlock.Status)
		fmt.Println("Processing Time:", fetchedBlock.ProcessingTime)
		fmt.Printf("Transactions:\n")

		for txnID, txn := range fetchedBlock.Txns {
			fmt.Printf("  TxnID: %s\t", txnID)
			fmt.Printf("  IsValid: %t\t", txn.IsValid)
			fmt.Printf("  Data:\t")
			for key, entry := range txn.Data {
				fmt.Printf("    Key: %s\t", key)
				fmt.Printf("    Value: %v\t", entry)
			}
		}
		fmt.Println()

	var input string
		fmt.Print("Do you want to fetch another block? (yes/no): ")
		_, err = fmt.Scanln(&input)
		if err != nil {
			log.Fatal("Error reading user input:", err)
		}

		if input != "yes" {
			return nil
		}
	}
}