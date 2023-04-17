package main

import (	
		"fmt"
		// "context"
		"strconv"
		"encoding/csv"
		"os"
		// "bufio"
		"math/big"
		// "reflect"
		// "encoding/binary"
		// "github.com/syndtr/goleveldb/leveldb"	
		// "github.com/ethereum/go-ethereum/rlp"
		"github.com/ethereum/go-ethereum/common"
		"github.com/ethereum/go-ethereum/core/rawdb"
		"github.com/ethereum/go-ethereum/core/types"
		// "github.com/ethereum/go-ethereum/params"
		// "github.com/ethereum/go-ethereum/accounts/abi"

		// "github.com/sirupsen/logrus"
		// "github.com/tsuna/gohbase"
		// "github.com/tsuna/gohbase/hrpc"
		// "github.com/tsuna/gohbase/pb"
)

var (
	// Path of ldb folder
	dbPath      = "/eth/geth/chaindata"

	// Path of ancient database
	ancientPath = dbPath + "/ancient"

	// The begin of the scan block height
	upNum       = 12500000 // 5773947 4634748 46147 12500000

	// The end of block height
	endNum      = 14615251
)

type Bar struct {
    percent int64  //百分比
    cur     int64  //当前进度位置
    total   int64  //总进度
    rate    string //进度条
    graph   string //显示符号
}

func (bar *Bar) NewOption(start, total int64) {
    bar.cur = start
    bar.total = total
    if bar.graph == "" {
        bar.graph = "█"
    }
    bar.percent = bar.getPercent()
    for i := 0; i < int(bar.percent); i += 2 {
        bar.rate += bar.graph //初始化进度条位置
    }
}

func (bar *Bar) getPercent() int64 {
    return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Bar) NewOptionWithGraph(start, total int64, graph string) {
    bar.graph = graph
    bar.NewOption(start, total)
}

func (bar *Bar) Play(cur int64) {
    bar.cur = cur
    last := bar.percent
    bar.percent = bar.getPercent()
    if bar.percent != last && bar.percent%2 == 0 {
        bar.rate += bar.graph
    }
    fmt.Printf("\r[%-50s]%3d%%  %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}

func (bar *Bar) Finish(){
    fmt.Println()
}

func main() {

	ancientDb, _ := rawdb.NewLevelDBDatabaseWithFreezer(dbPath, 16, 1, ancientPath, "", true)

	tranFile, _ := os.OpenFile("tran.csv", os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)

	tw := csv.NewWriter(tranFile)

	var bar Bar    

    bar.NewOption(int64(upNum), int64(endNum))

	for i := upNum; i <= endNum; i = i + 1 {

		bar.Play(int64(i))
		// Get blockHash by height
		blkHash := rawdb.ReadCanonicalHash(ancientDb, uint64(i))

		blkHeader := rawdb.ReadHeader(ancientDb, blkHash, uint64(i))

		blkBody := rawdb.ReadBody(ancientDb, blkHash, uint64(i))

		if len(blkBody.Transactions) != 0 {

			for _, tx := range blkBody.Transactions {

				var to string
				
				if tx.To() == nil {
					to = ""
				} else {
					to = tx.To().Hex()
				}
				
				tw.Write([]string{
					blkHeader.Number.String(),                 // blockHeigh
					strconv.Itoa(int(blkHeader.Time)),		   // time
					strconv.Itoa(int(tx.Nonce())),             // nonce
					tx.Hash().Hex(),						   // hash
					getFromAddr(tx, blkHeader.Number).Hex(),   // from 
					to,							               // to
					fmt.Sprintf("%x", tx.Data()),              // data
					strconv.Itoa(int(tx.Gas())),               // gas
				 	tx.GasPrice().String(),                    // gasPrice
					tx.Value().String(),                       // value
					tx.Cost().String(),                        // cost
				})
		    }
		}
	}
	bar.Finish()
	tw.Flush()
	tranFile.Close()
}


func getFromAddr(tx *types.Transaction, Number *big.Int) common.Address {
	
	var chainid *big.Int = big.NewInt(1)	

	var signer types.Signer

	if big.NewInt(2675001).Cmp(Number) >= 0 {
		signer = types.FrontierSigner{}
	} else if big.NewInt(12250640).Cmp(Number) >= 0 {
		signer = types.NewEIP155Signer(chainid)
	} else {
		signer = types.NewEIP2930Signer(chainid)
	}

	addr, err := types.Sender(signer, tx)

	if err != nil {
		fmt.Println(Number)
		panic(err)
	}
	return addr
}
