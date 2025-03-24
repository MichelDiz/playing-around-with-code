package helpers

import (
	"crypto/sha256"
	"fmt"
	"wallet/pkg/models"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/btcsuite/btcutil/hdkeychain"
	"golang.org/x/crypto/ripemd160"
)

func GetPrivateKeyForAddress(state *models.WalletState, address string) ([]byte, bool) {
	for i, addr := range state.Addresses {
		if addr[0] == address {
			// Retorna a chave privada correspondente
			return state.PrivateKeys[i], true
		}
	}
	// Se não encontrado, retorna nil e false
	return nil, false
}

func GetP2WPKHProgram(pubKey []byte, version byte) ([]byte, error) {
	if len(pubKey) != 33 {
		return nil, fmt.Errorf("chave pública inválida: esperado 33 bytes, recebido %d bytes", len(pubKey))
	}

	fmt.Printf("Public Key (hex): %x\n", pubKey)

	sha256Hash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	_, err := ripemdHasher.Write(sha256Hash[:])
	if err != nil {
		return nil, fmt.Errorf("erro ao calcular RIPEMD-160: %w", err)
	}
	hash160 := ripemdHasher.Sum(nil)

	scriptPubKey := make([]byte, 2+len(hash160))
	scriptPubKey[0] = version
	scriptPubKey[1] = byte(len(hash160))
	copy(scriptPubKey[2:], hash160)

	return scriptPubKey, nil
}

func DeriveKeyPairs(xprv string, count int, state *models.WalletState) error {
	masterKey, err := hdkeychain.NewKeyFromString(xprv)
	if err != nil {
		return fmt.Errorf("erro ao parsear xprv: %w", err)
	}

	// Caminho de derivação: /84h/1h/0h
	purposeKey, err := masterKey.Child(84 + hdkeychain.HardenedKeyStart)
	if err != nil {
		return fmt.Errorf("erro ao derivar propósito (84h): %w", err)
	}

	coinTypeKey, err := purposeKey.Child(1 + hdkeychain.HardenedKeyStart) // Testnet (1h)
	if err != nil {
		return fmt.Errorf("erro ao derivar moeda (1h): %w", err)
	}

	accountKey, err := coinTypeKey.Child(0 + hdkeychain.HardenedKeyStart) // Conta (0h)
	if err != nil {
		return fmt.Errorf("erro ao derivar conta (0h): %w", err)
	}

	// Derivar chaves /0/* (recebimento)
	receiveKey, err := accountKey.Child(0) // Branch (0)
	if err != nil {
		return fmt.Errorf("erro ao derivar branch (0): %w", err)
	}

	for i := 0; i < count; i++ {
		childKey, err := receiveKey.Child(uint32(i))
		if err != nil {
			return fmt.Errorf("erro na derivação do índice %d: %w", i, err)
		}

		// Chave privada
		privKey, err := childKey.ECPrivKey()
		if err != nil {
			return fmt.Errorf("erro ao extrair chave privada: %w", err)
		}
		state.PrivateKeys = append(state.PrivateKeys, privKey.Serialize())

		// Chave pública
		pubKey := privKey.PubKey()
		state.PublicKeys = append(state.PublicKeys, pubKey.SerializeCompressed())

		// Gerar o endereço bech32
		address, err := GenerateSegWitAddress(pubKey.SerializeCompressed())
		if err != nil {
			return fmt.Errorf("erro ao gerar endereço: %w", err)
		}
		state.Addresses = append(state.Addresses, []string{address})
	}

	return nil
}

func GenerateSegWitAddress(pubKey []byte) (string, error) {
	if len(pubKey) != 33 {
		return "", fmt.Errorf("chave pública inválida: esperado 33 bytes, recebido %d bytes", len(pubKey))
	}

	// Hash SHA256 da chave pública
	sha256Hash := sha256.Sum256(pubKey)

	// Hash RIPEMD-160 do resultado do SHA256
	ripemdHasher := ripemd160.New()
	_, err := ripemdHasher.Write(sha256Hash[:])
	if err != nil {
		return "", fmt.Errorf("erro ao calcular RIPEMD-160: %w", err)
	}
	hash160 := ripemdHasher.Sum(nil)

	// Codificação bech32 para SegWit
	witnessVersion := byte(0)                            // Versão 0 para SegWit nativo
	data, err := bech32.ConvertBits(hash160, 8, 5, true) // Converte para base 32
	if err != nil {
		return "", fmt.Errorf("erro ao converter para bits base 32: %w", err)
	}

	// Codificar para formato bech32 (testnet usa "tb", mainnet usa "bc")
	address, err := bech32.Encode("tb", append([]byte{witnessVersion}, data...))
	if err != nil {
		return "", fmt.Errorf("erro ao codificar endereço bech32: %w", err)
	}

	return address, nil
}
