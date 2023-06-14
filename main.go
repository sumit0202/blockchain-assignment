package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"
	"os"
	"github.com/syndtr/goleveldb/leveldb"
	blockpkg "github.com/sumit0202/assignment/block"
)

func main() {
	rand.Seed(time.Now().UnixNano())	//random number generator with the current time as the seed
	startTime := time.Now()
	db, err := leveldb.OpenFile("leveldb", nil)
	if err != nil {
		fmt.Println("Error opening LevelDB:", err)
		return
	}
	defer db.Close()

	for i := 1; i <= 10000; i++ {
		key := fmt.Sprintf("SIM%d", i)
		value := blockpkg.Entry{
			Val:   float64(i),
			Ver:   1.0,
			Valid: true,
		}

		value.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%f", key, value.Val))))

		data, err := json.Marshal(value)
		if err != nil {
			log.Fatal("Error serializing entry:", err)
		}

		err = db.Put([]byte(key), data, nil)
		if err != nil {
			log.Fatal("Error inserting entry into LevelDB:", err)
		}
	}

	transactionsPerBlocks := os.Getenv("transperblock")
	transactionsPerBlock,_ := strconv.Atoi(transactionsPerBlocks)

	totalTransactionstr := os.Getenv("envar")
	totalTransactions,_ := strconv.Atoi(totalTransactionstr)

	totalBlocks := int(math.Ceil(float64(totalTransactions) / float64(transactionsPerBlock)))

	prevBlockHash := "0xabc123" 
	blockWriter := make(blockpkg.BlockWriter)//creating a channel of type blockpkg.BlockWriter using the make function

	go func() {
        var blocks []blockpkg.MyBlock
        for block := range blockWriter {
            blocks = append(blocks, block)
            err := blockpkg.WriteBlocksToFile(blocks)
            if err != nil {
                fmt.Println("Error writing blocks to file:", err)
                return
            }
        }
    }()

	for blockNumber := 1; blockNumber <= totalBlocks; blockNumber++ {
		currBlockHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d", prevBlockHash, blockNumber))))

		block := blockpkg.MyBlock{
			BlockNumber:   blockNumber,
			Status:        blockpkg.Pending,
			PrevBlockHash: prevBlockHash,
			CurrBlockHash: currBlockHash,
			Txns:          make(map[string]blockpkg.Transaction),
		}

        block.ProcessingTime = time.Since(startTime)
		startIndex := (blockNumber - 1) * transactionsPerBlock+1
		endIndex := blockNumber*transactionsPerBlock
		if endIndex > totalTransactions {
			endIndex = totalTransactions
		}

		InputTxns := make([]map[string]blockpkg.Entry, endIndex-startIndex)
		//length of the slice is determined by endIndex - startIndex
		for i := startIndex; i < endIndex; i++ {
			txn := make(map[string]blockpkg.Entry)
			key := fmt.Sprintf("SIM%d", i+1)
			value := blockpkg.Entry{
				Val:   float64(i + 1),
				Ver:   float64(rand.Intn(10) + 1),
				Valid: true,
			}

			txn[key] = value
			InputTxns[i-startIndex] = txn
		}

		var wg sync.WaitGroup
		txnChan := make(chan blockpkg.Transaction)

		for i := 0; i < len(InputTxns); i += transactionsPerBlock {
			wg.Add(1)
			go block.PushValidTransactions(db, InputTxns[i:int(math.Min(float64(i+transactionsPerBlock), float64(len(InputTxns))))], txnChan, &wg)
		}

		go func() {
			wg.Wait()
			close(txnChan)
		}()

		for txn := range txnChan {
			block.Txns[txn.TxnId] = txn
		}

		block.UpdateBlockStatus(blockpkg.Committed)
		blockWriter.WriteBlock(block) 
		prevBlockHash = currBlockHash

	}
	close(blockWriter) 

	err = blockpkg.BlockDetails()
	
	if err !=nil{
		return
	}
	allBlocks, err := blockpkg.FetchAllBlockDetails()
		if err != nil {
			fmt.Println("Error fetching all block details:", err)
			return
		}

	fmt.Println("\nAll Blocks:")
	blockpkg.PrintBlockDetails(allBlocks)

	fmt.Println("Your envr var for total txn is:",totalTransactions)
	fmt.Println("Your txn per block is:",transactionsPerBlock)
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	fmt.Println("\nTotal Execution Time:", elapsedTime)
}