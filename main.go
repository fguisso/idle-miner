package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrd/wire"
)

func handlerChan(block chan []byte, myLastBlock *chainhash.Hash) *rpcclient.NotificationHandlers {
	return &rpcclient.NotificationHandlers{
		OnBlockConnected: func(blockHeader []byte, transactions [][]byte) {
			log.Printf("Block connected: ")
			go func() {
				buffer := bytes.NewBuffer(blockHeader)
				var bh wire.BlockHeader
				bh.Deserialize(buffer)
				blockHash := bh.BlockHash()
				if blockHash != *myLastBlock {
					block <- blockHeader
				}
			}()
		},
	}
}

func generateBlock(client *rpcclient.Client) (chainhash.Hash, error) {
	var emptyChainHash chainhash.Hash
	minerStatus, err := client.GetGenerate()
	if err != nil {
		log.Fatal(err)
	}
	if minerStatus {
		return emptyChainHash, errors.New("Another generate function is on load!")
	}
	blocks, err := client.Generate(1)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated blocks hashes: %s", blocks[0].String())
	return *blocks[0], nil
}

func main() {
	dcrdHomeDir := dcrutil.AppDataDir("dcrd", false)
	certs, err := ioutil.ReadFile(filepath.Join(dcrdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}
	var myLastBlock chainhash.Hash
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:19556",
		Endpoint:     "ws",
		User:         "8Qpf4pG/dysKEdIOm89k/1ZyLfU=",
		Pass:         "2kBX9bCS9mUe0nKQ2FlhNl16q1s=",
		Certificates: certs,
	}
	block := make(chan []byte)
	client, err := rpcclient.New(connCfg, handlerChan(block, &myLastBlock))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.NotifyBlocks(); err != nil {
		log.Fatal(err)
	}
	log.Println("NotifyBlocks: Registration Complete.")

	go func() {
		for {
			hash, _, err := client.GetBestBlock()
			if err != nil {
				log.Fatal(err)
			}

			blockHeader, err := client.GetBlockHeader(hash)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Best block: %d %s \n", blockHeader.Height, blockHeader.BlockHash())

			plustentime := blockHeader.Timestamp.Add(30 * time.Second)
			timeNow := time.Now()
			diference := plustentime.Sub(timeNow)

			log.Printf("I'm go to sleep %v minutes.", diference)
			timer := time.NewTimer(diference)
			log.Println("timer: ", timer != nil)
			select {
			case <-block:
				//log.Printf("My last block: %v %T", myLastBlock, myLastBlock)
				//log.Printf("Last block: %v %T", hash, hash)
				//log.Println("hash != myLastBlock: ", *hash != *myLastBlock)
				timer.Stop()
				log.Println("Recived a block: ")
			case <-timer.C:
				log.Println("Generate a block before wait diference.")
				blockDone, err := generateBlock(client)
				if err != nil {
					log.Fatal(err)
				}
				myLastBlock = blockDone

			}
		}
	}()

	log.Println("Wait for shutdown...")
	client.WaitForShutdown()
	log.Println("Shutdown!")
}
