package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// CLI is the structure which implements the cli.
type CLI struct{}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  generateaddress - Generates a wallet address")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  sendcoins -address ADDRESS -amount AMOUNT - Send AMOUNT of coins to ADDRESS")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
}

func (cli *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain([]byte(address))
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) sendCoins(from string, to string, amount int) {
	bc := NewBlockchain([]byte(from))
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	// TODO: Fix this
	bc := NewBlockchain([]byte(""))
	defer bc.db.Close()

	bci := bc.Iterator()
	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Transactions: %+v\n", block.Transactions)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

//Run is the function used by the cli to accept commands and process them
func (cli *CLI) Run() {
	cli.validateArgs()

	printUsageCmd := flag.NewFlagSet("help", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "help":
		err := printUsageCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error adding getBalanceCmd to CLI")
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error adding getBalanceCmd to CLI")
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error adding createBlockchainCmd to CLI")
		}
	case "sendcoins":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error adding sendCoins to CLI")
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error adding printChainCmd to CLI")
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	// The following lines are calls to Parsed()
	// which checks whether or not Parse() has
	// been called, thus printing theirresult
	if printUsageCmd.Parsed() {
		cli.printUsage()
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.sendCoins(*sendFrom, *sendTo, *sendAmount)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}

}
