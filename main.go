package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
)

func handlerChan(block chan []byte) *rpcclient.NotificationHandlers {
	return &rpcclient.NotificationHandlers{
		OnBlockConnected: func(blockHeader []byte, transactions [][]byte) {
			log.Printf("Block connected: ")
			go func() {
				block <- blockHeader
			}()
		},
	}
}

func generateBlock(client *rpcclient.Client) {
	minerStatus, err := client.GetGenerate()
	if err != nil {
		log.Fatal(err)
	}
	if minerStatus {
		log.Println("Another generate function is on load!")
		return
	}
	blocks, err := client.Generate(1)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated blocks hashes: %s", blocks[0].String())
	return blocks[0]
}

func main() {
	dcrdHomeDir := dcrutil.AppDataDir("dcrd", false)
	certs, err := ioutil.ReadFile(filepath.Join(dcrdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:19556",
		Endpoint:     "ws",
		User:         "8Qpf4pG/dysKEdIOm89k/1ZyLfU=",
		Pass:         "2kBX9bCS9mUe0nKQ2FlhNl16q1s=",
		Certificates: certs,
	}
	block := make(chan []byte)
	client, err := rpcclient.New(connCfg, handlerChan(block))
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
			time_now := time.Now()
			diference := plustentime.Sub(time_now)

			log.Printf("I'm go to sleep %v minutes.", diference)
			timer := time.NewTimer(diference)
			log.Println("timer: ", timer != nil)
			select {
			case <-block:
				timer.Stop()
				log.Println("Recived a block: ")
			case <-timer.C:
				log.Println("Generate a block before wait diference.")
				generateBlock(client)
			}
		}
	}()

	log.Println("Wait for shutdown...")
	client.WaitForShutdown()
	log.Println("Shutdown!")
}
