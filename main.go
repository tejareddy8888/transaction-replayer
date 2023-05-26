package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

func newUint64(val uint64) *uint64 { return &val }

func buildTx(tx *types.Transaction) *types.Transaction {

	// Extract necessary information from the original transaction
	toAddress := tx.To()
	data := tx.Data()
	value := tx.Value()
	gasLimit := tx.Gas()
	nonce := tx.Nonce()

	txType := tx.Type()
	var newTx *types.Transaction

	switch txType {
	case 0:
		gasPrice := tx.GasPrice()
		// Construct a new transaction object
		newTx = types.NewTransaction(nonce, *toAddress, value, gasLimit, gasPrice, data)
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
		// Construct a new transaction object
		newTx = types.NewTx(&types.DynamicFeeTx{Nonce: nonce,
			To:         toAddress,
			Value:      value,
			Gas:        gasLimit,
			GasTipCap:  gasTipCap,
			GasFeeCap:  gasFeeCap,
			Data:       data,
			AccessList: accessList})

	default:
		log.Fatal("Transaction not found on both chains")
	}

	// Set the V, R, and S values from the original transaction
	v, r, s := tx.RawSignatureValues()

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
		ShanghaiTime:                  newUint64(0),
		Ethash:                        new(params.EthashConfig),
	}

	signer := types.MakeSigner(config, big.NewInt(0))

	response, err := newTx.WithSignature(signer, append(append(r.Bytes(), s.Bytes()...), v.Bytes()...))

	if err != nil {
		log.Fatal(err)
	}
	return response
}

func main() {
	// EthashConfig is the consensus engine configs for proof-of-work based sealing.

	// Connect to the Ethereum clients for each chain split
	clientA, err := ethclient.Dial("http://localhost:8545") // Replace with the appropriate URL for chain A
	if err != nil {
		log.Fatal(err)
	}

	clientB, err := ethclient.Dial("http://localhost:8546") // Replace with the appropriate URL for chain B
	if err != nil {
		log.Fatal(err)
	}

	// Get the latest block number
	latestBlockNumber, err := clientA.BlockNumber(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over the blocks
	for i := int64(0); i <= int64(latestBlockNumber); i += i + 1 {
		// Get the block by number
		block, err := clientA.BlockByNumber(context.Background(), big.NewInt(i))
		if err != nil {
			log.Fatal(err)
		}

		transactions := block.Transactions()

		// Process the block data
		for _, tx := range transactions {

			// Transaction details
			txHash := tx.Hash() // Replace with the transaction hash you want to replay

			// Fetch transaction details from chain A
			tx, status, err := clientA.TransactionByHash(context.Background(), txHash)
			if err != nil {
				log.Fatal(err)
			}

			// Check if the transaction exists on both chains
			if tx == nil || status {
				log.Fatal("Transaction not found on chains")
			}

			// Broadcast the replayed transaction on chain B
			err = clientB.SendTransaction(context.Background(), tx)
			if err != nil {
				log.Fatal(err)
			}

			receiptB, err := clientB.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Transaction replayed on chain B:", receiptB.Status)
		}
	}

}
