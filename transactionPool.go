package main

import "errors"

var transactionPool []Transaction = []Transaction{}

func getTransactionPool() []Transaction {
	b := append(transactionPool[:0:0], transactionPool...)

	return b
}

func addToTransactionPool(tx *Transaction, unspentTxOuts []UnspentTxOut) error {
	if tx == nil {
		return errors.New("Trying to add invalid tx to pool")
	}

	if !validateTransaction(*tx, unspentTxOuts) {
		return errors.New("Trying to add invalid tx to pool")
	}

	if !isValidTxForPool(*tx, transactionPool) {
		return errors.New("Trying to add invalid tx to pool")
	}

	transactionPool = append(transactionPool, *tx)

	return nil
}

func isValidTxForPool(tx Transaction, aTransactionPool []Transaction) bool {
	txPoolIns := getTxPoolIns(aTransactionPool)

	containsTxIn := func(txIns []TxIn, txIn TxIn) bool {
		for _, txPoolIn := range txIns {
			if txIn.TxOutIndex == txPoolIn.TxOutIndex && txIn.TxOutID == txPoolIn.TxOutID {
				return true
			}
		}
		return false
	}

	for _, txIn := range tx.TxIns {
		if containsTxIn(txPoolIns, txIn) {
			return false
		}
	}

	return true
}

func getTxPoolIns(aTransactionPool []Transaction) []TxIn {
	result := []TxIn{}

	for _, t := range aTransactionPool {
		for _, txIn := range t.TxIns {
			result = append(result, txIn)
		}
	}

	return result
}

func hasTxIn(txIn TxIn, unspentTxOuts []UnspentTxOut) bool {
	for _, uTxO := range unspentTxOuts {
		if uTxO.TxOutID == txIn.TxOutID && uTxO.TxOutIndex == txIn.TxOutIndex {
			return true
		}
	}

	return false
}

func updateTransactionPool(unspentTxOuts []UnspentTxOut) error {
	invalidTxs := map[string]bool{}

	for _, tx := range transactionPool {
		for _, txIn := range tx.TxIns {
			if !hasTxIn(txIn, unspentTxOuts) {
				invalidTxs[tx.ID] = true
			}
		}
	}

	if len(invalidTxs) > 0 {
		newTransactionPool := []Transaction{}

		for _, tx := range transactionPool {
			if _, ok := invalidTxs[tx.ID]; !ok {
				newTransactionPool = append(newTransactionPool, tx)
			}
		}
	}

	return nil
}
