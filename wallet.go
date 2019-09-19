package main

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
)

const privateKeyLocation = "./node/wallet/private_key"

func GetPrivateFromWallet() (string, error) {
	f, err := os.Open(privateKeyLocation)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func GetPublicFromWallet() (string, error) {
	privateKey, err := GetPrivateFromWallet()

	if err != nil {
		return "", err
	}

	key, err := ParseRsaPrivateKeyFromPemStr(privateKey)

	if err != nil {
		return "", err
	}

	x := key.PublicKey.X
	y := key.PublicKey.Y

	publicKey := x.Bytes()
	publicKey = append(publicKey, y.Bytes()...)

	dst := make([]byte, hex.DecodedLen(len(publicKey)))
	n, err := hex.Decode(dst, publicKey)

	if err != nil {
		return "", err
	}

	return string(dst[:n]), nil
}

func generatePrivateKey() string {
	privateKey, err := NewPrivateKey()
	if err != nil {
		return ""
	}

	return privateKey.GetSerialize()
}

func initWallet() error {
	if fileExists(privateKeyLocation) {
		return errors.New("It has already exited")
	}

	newPrivateKey := generatePrivateKey()

	d1 := []byte(newPrivateKey)
	if err := ioutil.WriteFile(privateKeyLocation, d1, 0644); err != nil {
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getBalance(address string, unspentTxOuts []UnspentTxOut) int {
	sum := 0
	for _, uTxO := range unspentTxOuts {
		if uTxO.Address == address {
			sum += uTxO.Amount
		}
	}

	return sum
}

type TxOutsForAmount struct {
	IncludedUnspentTxOuts []UnspentTxOut
	LeftOverAmount        int
}

func findTxOutsForAmount(amount int, myUnspentTxOuts []UnspentTxOut) (*TxOutsForAmount, error) {
	currentAmount := 0
	includedUnspentTxOuts := []UnspentTxOut{}
	for _, myUnspentTxOut := range myUnspentTxOuts {
		includedUnspentTxOuts = append(includedUnspentTxOuts, myUnspentTxOut)
		currentAmount = currentAmount + myUnspentTxOut.Amount
		if currentAmount >= amount {
			// leftOverAmount  :=
			return &TxOutsForAmount{
				IncludedUnspentTxOuts: includedUnspentTxOuts,
				LeftOverAmount:        currentAmount - amount,
			}, nil
		}
	}

	return nil, errors.New("Not enough coins to send transaction")
}

func createTxOuts(receiverAddress string, myAddress string, amount, leftOverAmount int) []TxOut {
	txOut1 := TxOut{
		Address: receiverAddress,
		Amount:  amount,
	}

	if leftOverAmount == 0 {
		return []TxOut{txOut1}
	}

	leftOverTx := TxOut{
		Address: myAddress,
		Amount:  leftOverAmount,
	}

	return []TxOut{txOut1, leftOverTx}
}

func createTransaction(receiveAddress string, amount int, privateKey string, unspentTxOuts []UnspentTxOut, txPool []Transaction) (*Transaction, error) {
	myAddress := getPublicKey(privateKey)
	myUnspentTxOuts := []UnspentTxOut{}
	for _, utx := range unspentTxOuts {
		if utx.Address == myAddress {
			myUnspentTxOuts = append(myUnspentTxOuts, utx)
		}
	}

	txOutsForAmount, err := findTxOutsForAmount(amount, myUnspentTxOuts)

	if err != nil {
		return nil, err
	}

	toUnsignedTxIn := func(unspentTxOut UnspentTxOut) TxIn {
		return TxIn{
			TxOutID:    unspentTxOut.TxOutID,
			TxOutIndex: unspentTxOut.TxOutIndex,
		}
	}

	unsignedTxIns := []TxIn{}

	for _, uTxO := range txOutsForAmount.IncludedUnspentTxOuts {
		unsignedTxIns = append(unsignedTxIns, toUnsignedTxIn(uTxO))
	}

	tx := &Transaction{
		TxIns:  unsignedTxIns,
		TxOuts: createTxOuts(receiveAddress, myAddress, amount, txOutsForAmount.LeftOverAmount),
	}
	tx.ID = getTransactionID(*tx)

	for index, txIn := range tx.TxIns {
		txIn.Signature = signTxIn(*tx, index, privateKey, unspentTxOuts)
	}

	return tx, nil
}
