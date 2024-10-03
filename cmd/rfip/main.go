package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strconv"

     "github.com/ikweku/RFIP-MVP/pkg/blockchain"
)

type CLI struct{}

func (cli *CLI) createBlockchain(address string) {
    if !blockchain.ValidateAddress(address) {
        log.Panic("ERROR: Address is not valid")
    }
    bc := blockchain.CreateBlockchain(address)
    defer bc.db.Close()

    UTXOSet := blockchain.UTXOSet{bc}
    UTXOSet.Reindex()

    fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
    if !blockchain.ValidateAddress(address) {
        log.Panic("ERROR: Address is not valid")
    }
    bc := blockchain.NewBlockchain()
    UTXOSet := blockchain.UTXOSet{bc}
    defer bc.db.Close()

    balance := 0
    pubKeyHash := blockchain.Base58Decode([]byte(address))
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
    UTXOs := UTXOSet.FindUTXO(pubKeyHash)

    for _, out := range UTXOs {
        balance += out.Value
    }

    fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
    if !blockchain.ValidateAddress(from) {
        log.Panic("ERROR: Sender address is not valid")
    }
    if !blockchain.ValidateAddress(to) {
        log.Panic("ERROR: Recipient address is not valid")
    }

    bc := blockchain.NewBlockchain()
    UTXOSet := blockchain.UTXOSet{bc}
    defer bc.db.Close()

    wallets, err := blockchain.NewWallets()
    if err != nil {
        log.Panic(err)
    }
    wallet := wallets.GetWallet(from)

    tx := blockchain.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
    cbTx := blockchain.NewCoinbaseTX(from, "")
    txs := []*blockchain.Transaction{cbTx, tx}

    newBlock := bc.MineBlock(txs)
    UTXOSet.Update(newBlock)
    fmt.Println("Success!")
}

func (cli *CLI) createWallet() {
    wallets, _ := blockchain.NewWallets()
    address := wallets.CreateWallet()
    wallets.SaveToFile()

    fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) listAddresses() {
    wallets, err := blockchain.NewWallets()
    if err != nil {
        log.Panic(err)
    }
    addresses := wallets.GetAddresses()

    for _, address := range addresses {
        fmt.Println(address)
    }
}

func (cli *CLI) printChain() {
    bc := blockchain.NewBlockchain()
    defer bc.db.Close()

    bci := bc.Iterator()

    for {
        block := bci.Next()

        fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
        fmt.Printf("Hash: %x\n", block.Hash)
        pow := blockchain.NewProofOfWork(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }
}

func (cli *CLI) reindexUTXO() {
    bc := blockchain.NewBlockchain()
    UTXOSet := blockchain.UTXOSet{bc}
    UTXOSet.Reindex()

    count := UTXOSet.CountTransactions()
    fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CLI) Run() {
    cli.validateArgs()

    getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
    createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
    createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
    listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
    sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
    printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
    reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

    getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
    createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
    sendFrom := sendCmd.String("from", "", "Source wallet address")
    sendTo := sendCmd.String("to", "", "Destination wallet address")
    sendAmount := sendCmd.Int("amount", 0, "Amount to send")

    switch os.Args[1] {
    case "getbalance":
        err := getBalanceCmd.Parse(os.Args[2:])
    case "createblockchain":
        err := createBlockchainCmd.Parse(os.Args[2:])
    case "createwallet":
        err := createWalletCmd.Parse(os.Args[2:])
    case "listaddresses":
        err := listAddressesCmd.Parse(os.Args[2:])
    case "printchain":
        err := printChainCmd.Parse(os.Args[2:])
    case "send":
        err := sendCmd.Parse(os.Args[2:])
    case "reindexutxo":
        err := reindexUTXOCmd.Parse(os.Args[2:])
    default:
        cli.printUsage()
        os.Exit(1)
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

    if createWalletCmd.Parsed() {
        cli.createWallet()
    }

    if listAddressesCmd.Parsed() {
        cli.listAddresses()
    }

    if printChainCmd.Parsed() {
        cli.printChain()
    }

    if sendCmd.Parsed() {
        if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
            sendCmd.Usage()
            os.Exit(1)
        }

        cli.send(*sendFrom, *sendTo, *sendAmount)
    }

    if reindexUTXOCmd.Parsed() {
        cli.reindexUTXO()
    }
}

func (cli *CLI) validateArgs() {
    if len(os.Args) < 2 {
        cli.printUsage()
        os.Exit(1)
    }
}

func (cli *CLI) printUsage() {
    fmt.Println("Usage:")
    fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
    fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
    fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
    fmt.Println("  listaddresses - Lists all addresses from the wallet file")
    fmt.Println("  printchain - Print all the blocks of the blockchain")
    fmt.Println("  reindexutxo - Rebuilds the UTXO set")
    fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func main() {
    cli := CLI{}
    cli.Run()
}
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strconv"

    "your-project-path/blockchain"
)

type CLI struct{}

func (cli *CLI) createBlockchain(address string) {
    if !blockchain.ValidateAddress(address) {
        log.Panic("ERROR: Address is not valid")
    }
    bc := blockchain.CreateBlockchain(address)
    defer bc.db.Close()

    UTXOSet := blockchain.UTXOSet{bc}
    UTXOSet.Reindex()

    fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
    if !blockchain.ValidateAddress(address) {
        log.Panic("ERROR: Address is not valid")
    }
    bc := blockchain.NewBlockchain()
    UTXOSet := blockchain.UTXOSet{bc}
    defer bc.db.Close()

    balance := 0
    pubKeyHash := blockchain.Base58Decode([]byte(address))
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
    UTXOs := UTXOSet.FindUTXO(pubKeyHash)

    for _, out := range UTXOs {
        balance += out.Value
    }

    fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
    if !blockchain.ValidateAddress(from) {
        log.Panic("ERROR: Sender address is not valid")
    }
    if !blockchain.ValidateAddress(to) {
        log.Panic("ERROR: Recipient address is not valid")
    }

    bc := blockchain.NewBlockchain()
    UTXOSet := blockchain.UTXOSet{bc}
    defer bc.db.Close()

    wallets, err := blockchain.NewWallets()
    if err != nil {
        log.Panic(err)
    }
    wallet := wallets.GetWallet(from)

    tx := blockchain.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
    cbTx := blockchain.NewCoinbaseTX(from, "")
    txs := []*blockchain.Transaction{cbTx, tx}

    newBlock := bc.MineBlock(txs)
    UTXOSet.Update(newBlock)
    fmt.Println("Success!")
}

func (cli *CLI) createWallet() {
    wallets, _ := blockchain.NewWallets()
    address := wallets.CreateWallet()
    wallets.SaveToFile()

    fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) listAddresses() {
    wallets, err := blockchain.NewWallets()
    if err != nil {
        log.Panic(err)
    }
    addresses := wallets.GetAddresses()

    for _, address := range addresses {
        fmt.Println(address)
    }
}

func (cli *CLI) printChain() {
    bc := blockchain.NewBlockchain()
    defer bc.db.Close()

    bci := bc.Iterator()

    for {
        block := bci.Next()

        fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
        fmt.Printf("Hash: %x\n", block.Hash)
        pow := blockchain.NewProofOfWork(block)
        fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
        fmt.Println()

        if len(block.PrevBlockHash) == 0 {
            break
        }
    }
}

func (cli *CLI) reindexUTXO() {
    bc := blockchain.NewBlockchain()
    UTXOSet := blockchain.UTXOSet{bc}
    UTXOSet.Reindex()

    count := UTXOSet.CountTransactions()
    fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CLI) Run() {
    cli.validateArgs()

    getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
    createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
    createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
    listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
    sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
    printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
    reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

    getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
    createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
    sendFrom := sendCmd.String("from", "", "Source wallet address")
    sendTo := sendCmd.String("to", "", "Destination wallet address")
    sendAmount := sendCmd.Int("amount", 0, "Amount to send")

    switch os.Args[1] {
    case "getbalance":
        err := getBalanceCmd.Parse(os.Args[2:])
    case "createblockchain":
        err := createBlockchainCmd.Parse(os.Args[2:])
    case "createwallet":
        err := createWalletCmd.Parse(os.Args[2:])
    case "listaddresses":
        err := listAddressesCmd.Parse(os.Args[2:])
    case "printchain":
        err := printChainCmd.Parse(os.Args[2:])
    case "send":
        err := sendCmd.Parse(os.Args[2:])
    case "reindexutxo":
        err := reindexUTXOCmd.Parse(os.Args[2:])
    default:
        cli.printUsage()
        os.Exit(1)
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

    if createWalletCmd.Parsed() {
        cli.createWallet()
    }

    if listAddressesCmd.Parsed() {
        cli.listAddresses()
    }

    if printChainCmd.Parsed() {
        cli.printChain()
    }

    if sendCmd.Parsed() {
        if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
            sendCmd.Usage()
            os.Exit(1)
        }

        cli.send(*sendFrom, *sendTo, *sendAmount)
    }

    if reindexUTXOCmd.Parsed() {
        cli.reindexUTXO()
    }
}

func (cli *CLI) validateArgs() {
    if len(os.Args) < 2 {
        cli.printUsage()
        os.Exit(1)
    }
}

func (cli *CLI) printUsage() {
    fmt.Println("Usage:")
    fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
    fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
    fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
    fmt.Println("  listaddresses - Lists all addresses from the wallet file")
    fmt.Println("  printchain - Print all the blocks of the blockchain")
    fmt.Println("  reindexutxo - Rebuilds the UTXO set")
    fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func main() {
    cli := CLI{}
    cli.Run()
}

