package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

func newUint64(val uint64) *uint64 { return &val }

func getClientASigner(i int64) types.Signer {
	config := &params.ChainConfig{
		ChainID:                 big.NewInt(8888),
		HomesteadBlock:          big.NewInt(0),
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		TerminalTotalDifficulty: big.NewInt(60000000),
		Ethash:                  new(params.EthashConfig),
	}

	return types.MakeSigner(config, big.NewInt(i))
}

func getClientBSigner(i int64) types.Signer {
	config := &params.ChainConfig{
		ChainID:                       big.NewInt(8888),
		HomesteadBlock:                big.NewInt(0),
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		TerminalTotalDifficulty:       big.NewInt(0),
		TerminalTotalDifficultyPassed: true,
		ShanghaiTime:                  newUint64(1687439607),
		Ethash:                        new(params.EthashConfig),
	}

	return types.MakeSigner(config, big.NewInt(i))
}

func buildNewTx(ctx context.Context, sender common.Address, signer *types.Signer, tx *types.Transaction, client *ethclient.Client, privKey string, args ...string) *types.Transaction {
	// Extract necessary information from the original transaction

	var toAddress = tx.To()
	var value = tx.Value()
	if len(args) > 0 {
		overridenAddress := common.HexToAddress(args[0])
		toAddress = &overridenAddress
		n := new(big.Int)
		n, ok := n.SetString(args[1], 10)
		if !ok {
			fmt.Println("SetString: error %s", n)
		}
		value = n
	}
	fmt.Println("Updated tx value %s", value)
	fmt.Println("Updated tx to address %s", toAddress)
	data := tx.Data()
	gasLimit := estimateGas(ctx, sender, toAddress, data, value, client) + 100000
	nonce := tx.Nonce()
	txType := tx.Type()

	var newTx *types.Transaction

	fmt.Printf("\n tx type given: %d \n", txType)

	switch txType {
	case 0:
		gasPrice := tx.GasPrice()
		fmt.Printf("\n gas price given: %d \n", gasPrice)
		// Construct a new transaction object
		newTx = types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       toAddress,
			Value:    value,
			Gas:      gasLimit,
			GasPrice: gasPrice,
			Data:     data,
		})
	case 1:
		gasPrice := tx.GasPrice()
		chainId := tx.ChainId()
		accessList := tx.AccessList()
		// Construct a new transaction object
		newTx = types.NewTx(&types.AccessListTx{Nonce: nonce,
			To:         toAddress,
			Value:      value,
			Gas:        gasLimit,
			GasPrice:   gasPrice,
			Data:       data,
			AccessList: accessList,
			ChainID:    chainId})

	case 2:
		gasTipCap := tx.GasTipCap()
		gasFeeCap := tx.GasFeeCap()
		accessList := tx.AccessList()
		chainId := tx.ChainId()

		// Construct a new transaction object
		newTx = types.NewTx(&types.DynamicFeeTx{Nonce: nonce,
			To:         toAddress,
			Value:      value,
			Gas:        gasLimit,
			GasTipCap:  gasTipCap,
			GasFeeCap:  gasFeeCap,
			Data:       data,
			AccessList: accessList,
			ChainID:    chainId,
		})

	default:
		log.Fatal("Transaction Type does not exists")
	}

	fmt.Printf("\n priv in hex given: %s \n", privKey)

	priv, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal(err)
	}

	// // Set the V, R, and S values from the original transaction
	// v, r, s := tx.RawSignatureValues()
	signedTx, err := types.SignTx(newTx, *signer, priv)

	if err != nil {
		log.Fatal(err)
	}

	return signedTx
}

func estimateGas(ctx context.Context, fromAddress common.Address, toAddress *common.Address, data []byte, value *big.Int, client *ethclient.Client) uint64 {
	transaction := ethereum.CallMsg{
		From:  fromAddress,
		To:    toAddress,
		Value: value,
		Data:  data,
	}

	fmt.Printf("\n transaction to be estimated: %s \n", transaction)
	// Estimate the gas for the transaction
	gas, err := client.EstimateGas(ctx, transaction)

	fmt.Printf("\n gas estimated: %d \n", gas)

	if err != nil {
		log.Fatal(err)
	}
	return gas
}

func main() {
	// EthashConfig is the consensus engine configs for proof-of-work based sealing.
	var addressArray = [3]string{"0xd9dC96857daD6E570a771E8E8Ef6a94B08E55D9A", "0x75452375fe402c548671B66Af159452e4018D53D", "0x702332E028e45103a036Cf37E4cc6a9B55978A93"}

	privKeyAddressMapping := make(map[string]string)

	// Adding key-value pairs to the map

	// Connect to the Ethereum clients for each chain split
	clientA, err := ethclient.Dial("https://rpc.uzheths.ifi.uzh.ch/") // Replace with the appropriate URL for chain A
	if err != nil {
		log.Fatal(err)
	}

	// Get the latest block number
	latestBlockNumber, err := clientA.BlockNumber(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	clientB, err := ethclient.Dial("http://localhost:8545") // Replace with the appropriate URL for chain B
	if err != nil {
		log.Fatal(err)
	}

	// Get the latest block number
	clientBlatestBlockNumber, err := clientB.BlockNumber(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ClientA BlockNumber %d \n", latestBlockNumber)
	fmt.Printf("Connected to ClientB without errors %d \n", clientBlatestBlockNumber)

	var totalTransactionCount int = 0

	// Iterate over the blocks
	for i := int64(0); i <= int64(latestBlockNumber); i++ {
		// Get the block by number
		block, err := clientA.BlockByNumber(context.Background(), big.NewInt(i))
		if err != nil {
			log.Fatal(err)
		}

		transactions := block.Transactions()
		var blockTransactionLength int = 0
		if transactions.Len() == 0 {
			continue
		}

		blockTransactionLength = transactions.Len()
		fmt.Printf("\n Processing Block with number: %d and hash: %s with transactions of length: %d \n", i, block.Hash(), blockTransactionLength)
		appendedTransactions := make([]string, 0, blockTransactionLength)

		signer := getClientASigner(i)
		signerB := getClientBSigner(i)

		// Process the block data
		for _, tx := range transactions {

			sender, err := types.Sender(signer, tx)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Transaction is initiated from: %s\n", sender.Hex())
			// Transaction details
			txHash := tx.Hash() // Replace with the transaction hash you want to replay

			// Fetch transaction details from chain A
			tx, status, err := clientA.TransactionByHash(context.Background(), txHash)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Sleeping for 1 seconds for..... \n")
			fmt.Printf("Sending the transaction tx hash: %s \n", tx.Hash())
			time.Sleep(1 * time.Second)

			fmt.Printf("to address: %s \n value: %s \n hash: %s \n   ", tx.To(), tx.Value(), tx.Hash().Hex())

			for _, str := range addressArray {
				if strings.ToLower(sender.Hex()) == strings.ToLower(str) && tx.To() == nil {
					fmt.Printf("Found Contract Creation Transaction: %s\n", tx.Hash())
					fmt.Println(tx)
					newSignedTx := buildNewTx(context.Background(), sender, &signerB, tx, clientB, privKeyAddressMapping[strings.ToLower(sender.Hex())])
					fmt.Println(newSignedTx)
					tx = newSignedTx
				}
			}

			if strings.ToLower(txHash.Hex()) == strings.ToLower("0xbdfe0af8e439c05f2925b8cd30258b031520a4bf0c99d972aec8f359641be07e") {
				blockNumberB, err := clientB.BlockNumber(context.Background())
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("at this block is: %s\n", big.NewInt(int64(blockNumberB)))
				bal, err := clientB.BalanceAt(context.Background(), common.HexToAddress("0xd9dC96857daD6E570a771E8E8Ef6a94B08E55D9A"), big.NewInt(int64(blockNumberB)))

				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Sender's Bal at this block is: %s\n", bal)

				fmt.Printf("Found Problematic Transaction: %s\n", tx.Hash())

				n := new(big.Int)
				n, ok := n.SetString("999987922985332012900325659", 10)
				if !ok {
					fmt.Println("SetString: error")
					return
				}

				fmt.Printf("Subtracted Bal at this block is: %s\n", new(big.Int).Sub(bal, n).String())

				newSignedTx := buildNewTx(context.Background(), sender, &signerB, tx, clientB, privKeyAddressMapping[strings.ToLower(sender.Hex())], common.HexToAddress("0x702332E028e45103a036Cf37E4cc6a9B55978A93").Hex(), new(big.Int).Sub(bal, n).String())
				fmt.Println(newSignedTx)
				tx = newSignedTx
			}

			// Check if the transaction exists on both chains
			if tx == nil || status {
				log.Fatal("Transaction not found on chains")
			}

			// // Broadcast the replayed transaction on chain B
			err = clientB.SendTransaction(context.Background(), tx)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Sent Transaction on chain B in tx hash: %s \n", tx.Hash())

			appendedTransactions = append(appendedTransactions, tx.Hash().Hex())
		}

		fmt.Println(appendedTransactions)
		fmt.Printf("Sleeping for 45 seconds after a block \n")
		time.Sleep(45 * time.Second)
		fmt.Printf("Processed Block with number: %d and hash: %s with transactions of length: %d \n", i, block.Hash(), blockTransactionLength)

		for _, txHash := range appendedTransactions {
			fmt.Printf("The tx hash from array: %s \n", txHash)

			fmt.Printf("Checking if Transaction with hash %s  is on chain B \n", common.HexToHash(txHash))

			receiptB, err := clientB.TransactionReceipt(context.Background(), common.HexToHash(txHash))
			if err != nil {
				log.Fatal(err)
			}

			if receiptB.Status == 0 {
				fmt.Printf("The tx hash with: %s  has  \n", txHash)
			}

			fmt.Printf("\n Transaction with hash %s successfully replayed on chain B: %d \n", txHash, receiptB.Status)
		}
		totalTransactionCount += blockTransactionLength

		fmt.Printf("\n Total Number of Transactions: %d \n", totalTransactionCount)
	}
}
