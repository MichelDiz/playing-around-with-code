package main

import (
	"fmt"
	"time"

	"wallet/internal/storage"
	"wallet/pkg/helpers"
	"wallet/pkg/models"
	"wallet/pkg/progress"
)

//!Edil: Sugestão, quando encontrar um outpoint (uma utxo) que você controla, já guarda o script dela junto. Você precisa do script para gastar essa utxo.

func main() {

	start := time.Now()

	db, err := storage.Setup("./internal/badgerdb")
	if err != nil {
		fmt.Printf("Erro ao configurar o BadgerDB: %v\n", err)
		return
	}
	defer db.Close()

	//? Limpeza

	// if err := db.ClearHashes(); err != nil {
	// 	log.Fatalf("Erro ao limpar hashes: %v", err)
	// }

	// if err := db.ClearProcessedBlocks(); err != nil {
	// 	log.Fatalf("Erro ao limpar hashes: %v", err)
	// }

	// if err := db.ClearWalletHashes(); err != nil {
	// 	log.Fatalf("Erro ao limpar hashes: %v", err)
	// }

	// if err := db.Compact(); err != nil {
	// 	log.Printf("Erro ao compactar o banco: %v", err)
	// }

	lastProcessed, state, err := progress.LoadProgress(db.GetBadgerDB())
	if err != nil {
		fmt.Printf("Erro ao carregar progresso: %v\n", err)
		return
	}

	// Inicializar estado se não houver progresso salvo
	if state == nil {
		state = &models.WalletState{
			UTXOs:           make(map[string]models.UTXO),
			WitnessPrograms: [][]byte{},
			PublicKeys:      [][]byte{},
			PrivateKeys:     [][]byte{},
			Addresses:       [][]string{},
			Balance:         0,
		}
		fmt.Println("Inicializando estado da carteira...")
	}

	xprv := "tprv8ZgxMBicQKsPdt2JSGYoFa3bag1DMeGF8zdJC3ECLwCbUWdoZMq2wkqrN3zMaY9ep1RpD6yqLLmPohMgptXQ56YHr5NBLoUoXxLv97MjDcz"
	if state == nil {
	}

	helpers.DeriveKeyPairs(xprv, 2000, state)
	fmt.Println("Break point")

	// Gerar scriptPubKeys
	for i := 0; i < 2000; i++ {
		scriptPubKey, err := helpers.GetP2WPKHProgram(state.PublicKeys[i], 0)
		if err != nil {
			fmt.Printf("Erro no scriptPubKey: %v\n", err)
			return
		}
		state.WitnessPrograms = append(state.WitnessPrograms, scriptPubKey)
	}

	// Exibir algumas chaves derivadas
	for i := 0; i < 5; i++ {
		fmt.Printf("Par #%d:\n", i)
		fmt.Printf("  Chave privada: %x\n", state.PrivateKeys[i])
		fmt.Printf("  Chave pública: %x\n", state.PublicKeys[i])
		fmt.Printf("  Endereço: %s\n", state.Addresses[i][0])
		fmt.Printf("  WitnessProgram: %x\n", state.WitnessPrograms[i])
	}

	// helpers.Gen(&state)

	// Continuar do último bloco processado + 1
	startBlock := lastProcessed + 1
	targetBlock := 301
	for blockHeight := startBlock; blockHeight <= targetBlock; blockHeight++ {
		block, err := storage.FetchAndStoreBlock(db, blockHeight)
		if err != nil {
			fmt.Printf("Erro ao buscar o bloco %d: %v\n", blockHeight, err)
			continue
		}
		fmt.Printf("Processando bloco: %d\n", blockHeight)

		if txList, ok := block["tx"].([]interface{}); ok {
			for _, tx := range txList {
				txMap, ok := tx.(map[string]interface{})
				if !ok {
					fmt.Println("Erro ao converter transação")
					continue
				}

				// Processar saídas (vout) para adicionar UTXOs
				if voutList, ok := txMap["vout"].([]interface{}); ok {
					for voutIndex, vout := range voutList {
						voutMap, ok := vout.(map[string]interface{})
						if !ok {
							continue
						}

						scriptPubKey, ok := voutMap["scriptPubKey"].(map[string]interface{})
						if !ok {
							continue
						}
						address, ok := scriptPubKey["address"].(string)
						if !ok {
							continue
						}

						privateKey, found := helpers.GetPrivateKeyForAddress(state, address)
						if found {
							value, ok := voutMap["value"].(float64)
							if !ok {
								continue
							}
							txid, ok := txMap["txid"].(string)
							if !ok {
								continue
							}
							// fmt.Printf("Chave privada associada ao endereço %s: %x\n", address, privateKey)
							helpers.UpdateUTXO(state, txid, voutIndex, value, address, privateKey)
						}
					}
				}

				// Processar entradas (vin) para remover UTXOs gastos
				if vinList, ok := txMap["vin"].([]interface{}); ok {
					for _, vin := range vinList {
						vinMap, ok := vin.(map[string]interface{})
						if !ok {
							continue
						}

						txid, ok := vinMap["txid"].(string)
						if !ok {
							continue
						}

						voutIndex, ok := vinMap["vout"].(float64) // Índice do vout
						if !ok {
							continue
						}

						utxoKey := fmt.Sprintf("%s:%d", txid, int(voutIndex))
						if _, exists := state.UTXOs[utxoKey]; exists {
							delete(state.UTXOs, utxoKey)
							fmt.Printf("Removendo UTXO gasto: %s\n", utxoKey)
						}
					}
				}
			}
		}
	}

	// Salva snapshot da blockchain
	if targetBlock > lastProcessed {
		if err := progress.SaveProgress(db.GetBadgerDB(), targetBlock, state); err != nil {
			fmt.Printf("Erro ao salvar progresso ao final: %v\n", err)
		}
	}

	// Calcular saldo ao final
	helpers.CalculateBalance(state)

	// Imprimir saldo final
	//wallet_126 31.31556107
	//wpkh(tprv8ZgxMBicQKsPd5uXxEzM9s95NKFHfUrhj4dujNN9KcLdR2YXVobPxaSZt7rTNpNtVZRqqRAJW2oesqSJtEETbzgmQjg9aPR4Xs95BGV8EHQ/84h/1h/0h/0/*)#fc9pwrdn
	fmt.Printf("wallet_126 %.8f\n", state.Balance)

	elapsed := time.Since(start) // Final da medição de tempo
	fmt.Printf("Tempo total de execução: %s\n", elapsed)
	fmt.Printf("Último bloco processado: %d\n", lastProcessed)

	destinationAddress := "tb1q2z0yg87sxpeqftrj7cpx7zd3q0cthh22vda6la"

	amount := 0.01 // Valor a enviar
	fee := 0.0001  // Taxa de transação

	selectedUTXOs := make(map[string]models.UTXO)
	totalSelected := 0.0
	for key, utxo := range state.UTXOs {
		if totalSelected >= amount+fee {
			break
		}
		selectedUTXOs[key] = utxo
		totalSelected += utxo.Value
	}

	rawTx, err := helpers.CreateTransaction(selectedUTXOs, destinationAddress, amount, fee)
	if err != nil {
		fmt.Printf("Erro ao criar transação: %v\n", err)
		return
	}

	fmt.Printf("Transação criada com sucesso: %s\n", rawTx)

	fmt.Println("Break point")
}
