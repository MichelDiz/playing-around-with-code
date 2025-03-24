package models

// Estado da carteira
type WalletState struct {
	UTXOs           map[string]UTXO
	WitnessPrograms [][]byte
	PublicKeys      [][]byte
	PrivateKeys     [][]byte
	Addresses       [][]string
	Balance         float64
}

type UTXO struct {
	TxID       string // ID da transação
	VoutIndex  int    // Índice do vout
	Address    string // Endereço associado ao UTXO
	PrivateKey []byte
	Value      float64 // Valor do UTXO
}
