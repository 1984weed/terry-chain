// class UnspentTxOut {
// public readonly txOutId: string;
// public readonly txOutIndex: number;
// public readonly address: string;
// public readonly amount: number;

// constructor(txOutId: string, txOutIndex: number, address: string, amount: number) {
//     this.txOutId = txOutId;
//     this.txOutIndex = txOutIndex;
//     this.address = address;
//     this.amount = amount;
// }
// }
package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"math/big"
)

type UnspentTxOut struct {
	TxOutID    string
	TxOutIndex int
	Address    string
	Amount     int
}

func generageUnspentTxOut(txOutID string, txOutIndex int, address string, amount int) *UnspentTxOut {
	return &UnspentTxOut{
		TxOutID:    txOutID,
		TxOutIndex: txOutIndex,
		Address:    address,
		Amount:     amount,
	}
}

type TxIn struct {
	TxOutID    string
	TxOutIndex int
	Signature  string
}

type TxOut struct {
	Address string
	Amount  int
}

type Transaction struct {
	ID     string
	TxIns  []TxIn
	TxOuts []TxOut
}

func getTransactionID(transaction Transaction) string {
	txInContent := ""
	for _, txIn := range transaction.TxIns {
		txInContent += txIn.TxOutID + string(txIn.TxOutIndex)
	}

	txOutContent := ""

	for _, txOut := range transaction.TxOuts {
		txOutContent += txOut.Address + string(txOut.Amount)
	}

	h := sha256.New()

	h.Write([]byte(txInContent + txOutContent))

	bs := h.Sum(nil)

	return string(bs)
}

func signTxIn(transaction Transaction, txInIndex int, privateKey string, aUnspentTxOuts []UnspentTxOut) string {
	// txIn := transaction.TxIns[txInIndex]
	dataToSign := transaction.ID

	// referencedUnspentTxOut := findUnspentTxOut(txIn.TxOutID, txIn.TxOutIndex, aUnspentTxOuts)
	// referencedAddress := referencedUnspentTxOut.Address

	hash := sha256.Sum256([]byte(dataToSign))

	ecdsaPrivateKey, err := ParseRsaPrivateKeyFromPemStr(privateKey)
	if err != nil {
		return ""
	}

	r, s, err := ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(ecdsaPrivateKey), hash[:])

	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)

	return string(signature)
}

func stringToBigInt(key string) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(key, 16)
	if !ok {
		return nil
	}

	return n
}

func updateUnspentTxOuts(newTransactions []Transaction, aUnspentTxOuts []UnspentTxOut) []UnspentTxOut {
	newUnspentTxOuts := []UnspentTxOut{}

	for _, t := range newTransactions {
		for index, out := range t.TxOuts {
			newUnspentTxOuts = append(newUnspentTxOuts, UnspentTxOut{
				TxOutID:    t.ID,
				TxOutIndex: index,
				Address:    out.Address,
				Amount:     out.Amount,
			})
		}
	}

	consumedTxOuts := []UnspentTxOut{}
	for _, t := range newTransactions {
		for _, in := range t.TxIns {
			consumedTxOuts = append(consumedTxOuts, UnspentTxOut{
				TxOutID:    in.TxOutID,
				TxOutIndex: in.TxOutIndex,
				Address:    "",
				Amount:     0,
			})
		}
	}
	resultingUnspentTxOuts := newUnspentTxOuts

	for _, t := range aUnspentTxOuts {
		if findUnspentTxOut(t.TxOutID, t.TxOutIndex, consumedTxOuts) != nil {
			resultingUnspentTxOuts = append(resultingUnspentTxOuts, t)
		}
	}

	return resultingUnspentTxOuts
}

func findUnspentTxOut(transactionID string, index int, aUnspentTxOuts []UnspentTxOut) *UnspentTxOut {
	for _, t := range aUnspentTxOuts {
		if t.TxOutID == transactionID && t.TxOutIndex == index {
			return &t
		}
	}
	return nil
}

func validateTransaction(transaction Transaction, aUnspentTxOuts []UnspentTxOut) bool {
	if getTransactionID(transaction) != transaction.ID {
		return false
	}

	hasValidTxIns := true

	for _, t := range transaction.TxIns {
		hasValidTxIns = hasValidTxIns && validateTxIn(t, transaction, aUnspentTxOuts)
	}

	if hasValidTxIns {
		return false
	}

	return true
}

func validateTxIn(txIn TxIn, transaction Transaction, aUnspentTxOuts []UnspentTxOut) bool {
	var referencedUTxOut *UnspentTxOut
	for _, t := range aUnspentTxOuts {
		if t.TxOutID == txIn.TxOutID && t.TxOutID == txIn.TxOutID {
			referencedUTxOut = &t
		}
	}

	if referencedUTxOut == nil {
		return false
	}

	address := referencedUTxOut.Address

	serializedPubKey, err := hex.DecodeString(address)
	if err != nil {
		return false
	}

	publicKey, _ := ParsePubKey(serializedPubKey)

	sig, _ := parseSig([]byte(txIn.Signature))

	return sig.Verify([]byte(transaction.ID), publicKey)
}

const COINBASE_AMOUNT = 50

func validateCoinbaseTx(transaction Transaction, blockIndex int) bool {
	if getTransactionID(transaction) != transaction.ID {
		return false
	}

	if len(transaction.TxIns) != 1 {
		return false
	}

	if transaction.TxIns[0].TxOutIndex != blockIndex {
		return false
	}

	if len(transaction.TxOuts) != 1 {
		return false
	}
	if transaction.TxOuts[0].Amount != COINBASE_AMOUNT {
		return false
	}

	return true
}
func ProcessTransactions(aTransactions []Transaction, aUnspentTxOuts []UnspentTxOut, blockIndex int) []UnspentTxOut {
	if !validateBlockTransactions(aTransactions, aUnspentTxOuts, blockIndex) {
		fmt.Println("invalid block transactions")
		return []UnspentTxOut{}
	}
	return updateUnspentTxOuts(aTransactions, aUnspentTxOuts)
}

func validateBlockTransactions(aTransactions []Transaction, aUnspentTxOuts []UnspentTxOut, blockIndex int) bool {
	coinbaseTx := aTransactions[0]

	if !validateCoinbaseTx(coinbaseTx, blockIndex) {
		return false
	}
	txIns := []TxIn{}

	for _, tx := range aTransactions {
		for _, txIn := range tx.TxIns {
			txIns = append(txIns, txIn)
		}
	}

	if hasDuplicates(txIns) {
		return false
	}
	normalTransaction := aTransactions[1:]
	result := true
	for _, tx := range normalTransaction {
		result = result && validateTransaction(tx, aUnspentTxOuts)
	}

	return result
}

func hasDuplicates(txIns []TxIn) bool {
	keyCountMap := make(map[string]int)

	for _, txIn := range txIns {
		key := txIn.TxOutID + txIn.TxOutID
		if _, ok := keyCountMap[key]; ok {
			return true
		}
		keyCountMap[key] = 1
	}
	return false
}

func getCoinBaseTransaction(address string, blockIndex int) Transaction {
	t := Transaction{}
	txIn := TxIn{
		Signature:  "",
		TxOutID:    "",
		TxOutIndex: blockIndex,
	}

	t.TxIns = []TxIn{txIn}
	t.TxOuts = []TxOut{TxOut{
		Address: address,
		Amount:  COINBASE_AMOUNT,
	}}
	t.ID = getTransactionID(t)

	return t
}

func getPublicKey(aPrivateKey string) string {
	ecdsaPrivateKey, err := ParseRsaPrivateKeyFromPemStr(aPrivateKey)

	if err != nil {
		return ""
	}
	publicKey := PublicKey(ecdsaPrivateKey.PublicKey)

	return string(publicKey.SerializeUncompressed())
}

// const getPublicKey = (aPrivateKey: string): string => {
//     return ec.keyFromPrivate(aPrivateKey, 'hex').getPublic().encode('hex');
// };
// ProcessTransactions, signTxIn, getTransactionId,
// UnspentTxOut, TxIn, TxOut, getCoinbaseTransaction, getPublicKey,
// Transaction

// const validateCoinbaseTx = (transaction: Transaction, blockIndex: number): boolean => {
//     if (transaction == null) {
//         console.log('the first transaction in the block must be coinbase transaction');
//         return false;
//     }
//     if (getTransactionId(transaction) !== transaction.id) {
//         console.log('invalid coinbase tx id: ' + transaction.id);
//         return false;
//     }
//     if (transaction.txIns.length !== 1) {
//         console.log('one txIn must be specified in the coinbase transaction');
//         return;
//     }
//     if (transaction.txIns[0].txOutIndex !== blockIndex) {
//         console.log('the txIn signature in coinbase tx must be the block height');
//         return false;
//     }
//     if (transaction.txOuts.length !== 1) {
//         console.log('invalid number of txOuts in coinbase transaction');
//         return false;
//     }
//     if (transaction.txOuts[0].amount != COINBASE_AMOUNT) {
//         console.log('invalid coinbase amount in coinbase transaction');
//         return false;
//     }
//     return true;
// };

// const validateTransaction = (transaction: Transaction, aUnspentTxOuts: UnspentTxOut[]): boolean => {

//     if (getTransactionId(transaction) !== transaction.id) {
//         console.log('invalid tx id: ' + transaction.id);
//         return false;
//     }
//     const hasValidTxIns: boolean = transaction.txIns
//         .map((txIn) => validateTxIn(txIn, transaction, aUnspentTxOuts))
//         .reduce((a, b) => a && b, true);

//     if (!hasValidTxIns) {
//         console.log('some of the txIns are invalid in tx: ' + transaction.id);
//         return false;
//     }

//     const totalTxInValues: number = transaction.txIns
//         .map((txIn) => getTxInAmount(txIn, aUnspentTxOuts))
//         .reduce((a, b) => (a + b), 0);

//     const totalTxOutValues: number = transaction.txOuts
//         .map((txOut) => txOut.amount)
//         .reduce((a, b) => (a + b), 0);

//     if (totalTxOutValues !== totalTxInValues) {
//         console.log('totalTxOutValues !== totalTxInValues in tx: ' + transaction.id);
//         return false;
//     }

//     return true;
// };

// const validateTxIn = (txIn: TxIn, transaction: Transaction, aUnspentTxOuts: UnspentTxOut[]): boolean => {
//     const referencedUTxOut: UnspentTxOut =
//         aUnspentTxOuts.find((uTxO) => uTxO.txOutId === txIn.txOutId && uTxO.txOutId === txIn.txOutId);
//     if (referencedUTxOut == null) {
//         console.log('referenced txOut not found: ' + JSON.stringify(txIn));
//         return false;
//     }
//     const address = referencedUTxOut.address;

//     const key = ec.keyFromPublic(address, 'hex');
//     return key.verify(transaction.id, txIn.signature);
// };

// const validateTransaction = (transaction: Transaction, aUnspentTxOuts: UnspentTxOut[]): boolean => {

//     if (getTransactionId(transaction) !== transaction.id) {
//         console.log('invalid tx id: ' + transaction.id);
//         return false;
//     }
//     const hasValidTxIns: boolean = transaction.txIns
//         .map((txIn) => validateTxIn(txIn, transaction, aUnspentTxOuts))
//         .reduce((a, b) => a && b, true);

//     if (!hasValidTxIns) {
//         console.log('some of the txIns are invalid in tx: ' + transaction.id);
//         return false;
//     }

//     const totalTxInValues: number = transaction.txIns
//         .map((txIn) => getTxInAmount(txIn, aUnspentTxOuts))
//         .reduce((a, b) => (a + b), 0);

//     const totalTxOutValues: number = transaction.txOuts
//         .map((txOut) => txOut.amount)
//         .reduce((a, b) => (a + b), 0);

//     if (totalTxOutValues !== totalTxInValues) {
//         console.log('totalTxOutValues !== totalTxInValues in tx: ' + transaction.id);
//         return false;
//     }

//     return true;
// };

// func isValidTransactionStructure(transaction Transaction) bool {
// 	// validIns := true
// 	// for _, t := range transaction.TxIns {
// 	// 	validIns
// 	// }
// 	// if !validIns {
// 	// 	return false
// 	// }

// }

// const isValidTransactionStructure = (transaction: Transaction) => {
//     if (typeof transaction.id !== 'string') {
//         console.log('transactionId missing');
//         return false;
//     }
//     if (!(transaction.txIns instanceof Array)) {
//         console.log('invalid txIns type in transaction');
//         return false;
//     }
//     if (!transaction.txIns
//             .map(isValidTxInStructure)
//             .reduce((a, b) => (a && b), true)) {
//         return false;
//     }

//     if (!(transaction.txOuts instanceof Array)) {
//         console.log('invalid txIns type in transaction');
//         return false;
//     }

//     if (!transaction.txOuts
//             .map(isValidTxOutStructure)
//             .reduce((a, b) => (a && b), true)) {
//         return false;
//     }
//     return true;
// };

// const findUnspentTxOut = (transactionId: string, index: number, aUnspentTxOuts: UnspentTxOut[]): UnspentTxOut => {
//     return aUnspentTxOuts.find((uTxO) => uTxO.txOutId === transactionId && uTxO.txOutIndex === index);
// };

// const updateUnspentTxOuts = (newTransactions: Transaction[], aUnspentTxOuts: UnspentTxOut[]): UnspentTxOut[] => {
//     const newUnspentTxOuts: UnspentTxOut[] = newTransactions
//         .map((t) => {
//             return t.txOuts.map((txOut, index) => new UnspentTxOut(t.id, index, txOut.address, txOut.amount));
//         })
//         .reduce((a, b) => a.concat(b), []);

//     const consumedTxOuts: UnspentTxOut[] = newTransactions
//         .map((t) => t.txIns)
//         .reduce((a, b) => a.concat(b), [])
//         .map((txIn) => new UnspentTxOut(txIn.txOutId, txIn.txOutIndex, '', 0));

//     const resultingUnspentTxOuts = aUnspentTxOuts
//         .filter(((uTxO) => !findUnspentTxOut(uTxO.txOutId, uTxO.txOutIndex, consumedTxOuts)))
//         .concat(newUnspentTxOuts);

//     return resultingUnspentTxOuts;
// };

// func newUnspentTxOuts
// const newUnspentTxOuts: UnspentTxOut[] = newTransactions
// .map((t) => {
// 	return t.txOuts.map((txOut, index) => new UnspentTxOut(t.id, index, txOut.address, txOut.amount));
// })
// .reduce((a, b) => a.concat(b), []);
