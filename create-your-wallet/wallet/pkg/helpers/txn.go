package helpers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"wallet/pkg/models"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func CreateTransaction(utxos map[string]models.UTXO, destinationAddress string, amount, fee float64) (string, error) {
	tx := wire.NewMsgTx(wire.TxVersion)

	totalInput := 0.0

	// Entradas
	for _, utxo := range utxos {
		txHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return "", fmt.Errorf("falha ao decodificar txid: %w", err)
		}
		outPoint := wire.NewOutPoint(txHash, uint32(utxo.VoutIndex))
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)

		totalInput += utxo.Value
	}

	// Verifica se o saldo cobre a transação e a taxa
	if totalInput < (amount + fee) {
		return "", fmt.Errorf("saldo insuficiente para cobrir a transação e a taxa")
	}
	change := totalInput - (amount + fee)

	// Destino
	destinationAddr, err := btcutil.DecodeAddress(destinationAddress, &chaincfg.MainNetParams)
	if err != nil {
		return "", fmt.Errorf("endereço de destino inválido: %w", err)
	}
	pkScript, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return "", fmt.Errorf("falha ao criar Pay-to-Addr Script: %w", err)
	}
	txOut := wire.NewTxOut(int64(amount*1e8), pkScript)
	tx.AddTxOut(txOut)

	// Troco
	if change > 0 {
		var fromAddress btcutil.Address
		for _, utxo := range utxos {
			fromAddress, err = btcutil.DecodeAddress(utxo.Address, &chaincfg.MainNetParams)
			if err != nil {
				return "", fmt.Errorf("falha ao decodificar endereço de troco: %w", err)
			}
			break // Pega o primeiro UTXO encontrado
		}
		if err != nil {
			return "", fmt.Errorf("falha ao decodificar endereço de troco: %w", err)
		}
		changeAddrScript, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			return "", fmt.Errorf("falha ao criar script de troco: %w", err)
		}
		changeOut := wire.NewTxOut(int64(change*1e8), changeAddrScript)
		tx.AddTxOut(changeOut)
	}

	// Assinar entradas
	for i, txIn := range tx.TxIn {
		utxo := utxos[fmt.Sprintf("%s:%d", txIn.PreviousOutPoint.Hash.String(), txIn.PreviousOutPoint.Index)]

		// Decode private key
		privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), utxo.PrivateKey)

		decodedAddress, err := btcutil.DecodeAddress(utxo.Address, &chaincfg.MainNetParams)
		if err != nil {
			return "", fmt.Errorf("falha ao decodificar endereço %s: %w", utxo.Address, err)
		}

		pkScript, err := txscript.PayToAddrScript(decodedAddress)
		if err != nil {
			return "", fmt.Errorf("falha ao criar Pay-to-Addr Script para o endereço %s: %w", utxo.Address, err)
		}

		// Calcular os hashes de assinatura
		sigHashes := txscript.NewTxSigHashes(tx)

		// Criar a assinatura Witness
		sig, err := txscript.RawTxInWitnessSignature(tx, sigHashes, i, int64(utxo.Value*1e8), pkScript, txscript.SigHashAll, privateKey)
		if err != nil {
			return "", fmt.Errorf("falha ao criar assinatura Witness: %w", err)
		}

		// Adiciona a assinatura e a chave pública a Witness
		txIn.Witness = wire.TxWitness{sig, privateKey.PubKey().SerializeCompressed()}

		// Verificar a assinatura
		vm, err := txscript.NewEngine(pkScript, tx, i, txscript.StandardVerifyFlags, nil, nil, int64(utxo.Value*1e8))
		if err != nil {
			return "", fmt.Errorf("falha ao verificar: %w", err)
		}
		if err := vm.Execute(); err != nil {
			return "", fmt.Errorf("verificação de assinatura falhou: %w", err)
		}
	}

	var buf bytes.Buffer
	err = tx.Serialize(&buf)
	if err != nil {
		return "", fmt.Errorf("falha ao serializar transação: %w", err)
	}
	return hex.EncodeToString(buf.Bytes()), nil
}
