package tx

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/stafi-substrate-go/expand"
	"github.com/JFJun/stafi-substrate-go/utils"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
	"golang.org/x/crypto/blake2b"
	"strings"
)

type Transaction struct {
	SenderPubkey       string `json:"sender_pubkey"`    // from address public key ,0x开头
	RecipientPubkey    string `json:"recipient_pubkey"` // to address public key ,0x开头
	Amount             uint64 `json:"amount"`           // 转账金额
	Nonce              uint64 `json:"nonce"`            //nonce值
	Fee                uint64 `json:"fee"`              //手续费
	BlockHeight        uint64 `json:"block_height"`     //最新区块高度
	BlockHash          string `json:"block_hash"`       //最新区块hash
	GenesisHash        string `json:"genesis_hash"`     //
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
	CallId             string `json:"call_id"` //
	UtilityBatchCallId 	string
	PubkeyAmount 		map[string]uint64	`json:"pubkey_amount"`	//用于utilityBatch
}

func CreateTransaction(from, to string, amount, nonce uint64)*Transaction{
	return &Transaction{
		SenderPubkey: utils.AddressToPublicKey(from),
		RecipientPubkey: utils.AddressToPublicKey(to),
		Amount:          amount,
		Nonce:           nonce,
	}
}

func CreateUtilityBatchTransaction(from string,nonce uint64,address_amount map[string]uint64,utilityBatchCallId string)*Transaction{
	tx:=new(Transaction)
	tx.SenderPubkey = utils.AddressToPublicKey(from)
	tx.Nonce = nonce
	pub_amount:=make(map[string]uint64)
	for address,amount:=range address_amount{
		pub_amount[utils.AddressToPublicKey(address)] = amount
	}
	tx.PubkeyAmount = pub_amount
	tx.UtilityBatchCallId = utilityBatchCallId
	return tx
}

func (tx *Transaction) SetGenesisHashAndBlockHash(genesisHash, blockHash string) {
	tx.GenesisHash = utils.Remove0X(genesisHash)
	tx.BlockHash = utils.Remove0X(blockHash)
}

func (tx *Transaction) SetSpecVersionAndCallId(specVersion, transactionVersion uint32,callIdx string) {
	tx.SpecVersion = specVersion
	tx.TransactionVersion = transactionVersion
	tx.CallId = callIdx
}
func (tx *Transaction)checkTxParams()error{
	if tx.SenderPubkey=="" {
		return errors.New("send public key is null")
	}
	if	tx.BlockHash==""{
		return errors.New("block hash is null")
	}
	if tx.GenesisHash=="" {
		return errors.New("genesis hash is null")
	}
	if tx.CallId==""  {
		return errors.New("callIdx is null")
	}
	if tx.UtilityBatchCallId!="" {
		if len(tx.PubkeyAmount)==0 {
			return errors.New("public key Amount map is null")
		}
	}else{
		if tx.RecipientPubkey=="" {
			return errors.New("recipient public key is null")
		}
	}
	return nil
}

/*
signType: 0-->ecdsa   1--> sr25519
*/
func (tx *Transaction)SignTransaction(privateKey string,signType int)(string,error){
	var (
		call types.Call
		err  error
	)
	//1.  check params
	err = tx.checkTxParams()
	if err != nil {
		return "", fmt.Errorf("check params error: $v",err)
	}
	//2. create types.Call

	if tx.UtilityBatchCallId=="" {
		// Balances.transfer交易或者Balances.transfer_keep_alive交易
		call,err = expand.NewCall(tx.CallId,types.NewAddressFromAccountID(types.MustHexDecodeString(
			tx.RecipientPubkey)),
			types.NewUCompactFromUInt(tx.Amount))
	}else{
		// Utility.batch 交易
		var args []interface{}
		for address,amount:=range tx.PubkeyAmount{
			balanceTransferCall,err:=expand.NewCall(tx.CallId,types.NewAddressFromAccountID(types.MustHexDecodeString(address)),
				types.NewUCompactFromUInt(amount))
			if err != nil {
				return "", fmt.Errorf("create utility.batch calls error: %v",err)
			}
			args = append(args,balanceTransferCall)
		}
		call,err =expand.NewCall(tx.UtilityBatchCallId,args)
	}
	if err != nil {
		return "", fmt.Errorf("create types.Call error: %v",err)
	}
	ext :=types.NewExtrinsic(call)
	o := types.SignatureOptions{
		BlockHash: types.NewHash(types.MustHexDecodeString(tx.BlockHash)),
		// Era: ExtrinsicEra{IsImmortalEra: true},
		GenesisHash:        types.NewHash(types.MustHexDecodeString(tx.GenesisHash)),
		Nonce:              types.NewUCompactFromUInt(tx.Nonce),
		SpecVersion:        types.NewU32(uint32(tx.SpecVersion)),
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: types.NewU32(uint32(tx.TransactionVersion)),
	}
	e:=&ext
	//签名
	err = tx.signTx(e,o,privateKey,signType)
	if err != nil {
		return "", fmt.Errorf("sign error: %v",err)
	}
	return types.EncodeToHexString(e)
}

func (tx *Transaction)signTx(e *types.Extrinsic,o types.SignatureOptions,privateKey string,signType int)error{
	if e.Type() != types.ExtrinsicVersion4 {
		return fmt.Errorf("unsupported extrinsic version: %v (isSigned: %v, type: %v)", e.Version, e.IsSigned(), e.Type())
	}
	mb, err := types.EncodeToBytes(e.Method)
	if err != nil {
		return err
	}
	era := o.Era
	if !o.Era.IsMortalEra {
		era = types.ExtrinsicEra{IsImmortalEra: true}
	}
	payload := types.ExtrinsicPayloadV4{
		ExtrinsicPayloadV3: types.ExtrinsicPayloadV3{
			Method:      mb,
			Era:         era,
			Nonce:       o.Nonce,
			Tip:         o.Tip,
			SpecVersion: o.SpecVersion,
			GenesisHash: o.GenesisHash,
			BlockHash:   o.BlockHash,
		},
		TransactionVersion: o.TransactionVersion,
	}
	// sign
	data,err:=types.EncodeToBytes(payload)
	if err != nil {
		return fmt.Errorf("encode payload error: %v",err)
	}
	// if data is longer than 256 bytes, hash it first
	if len(data)>256 {
		h := blake2b.Sum256(data)
		data  = h[:]
	}
	privateKey = strings.TrimPrefix(privateKey,"0x")
	priv,err:=hex.DecodeString(privateKey)
	if err != nil {
		return fmt.Errorf("hex decode private key error: %v",err)
	}

	defer utils.ZeroBytes(priv)
	sig,err:=crypto.Sign(priv,data,signType)
	if err != nil {
		return fmt.Errorf("sign error: %v",err)
	}
	signerPubKey := types.NewAddressFromAccountID(types.MustHexDecodeString(
		tx.SenderPubkey))

	var ss types.MultiSignature
	if signType==crypto.EcdsaType {
		ss = types.MultiSignature{IsEcdsa: true, AsEcdsa: types.NewBytes(sig)}
	}else if signType==crypto.Sr25519Type {
		ss = types.MultiSignature{IsSr25519: true, AsSr25519: types.NewSignature(sig)}
	}else{
		return fmt.Errorf("unsupport sign type : %d",signType)
	}
	extSig := types.ExtrinsicSignatureV4{
		Signer:    signerPubKey,
		Signature: ss,
		Era:       era,
		Nonce:     o.Nonce,
		Tip:       o.Tip,
	}
	e.Signature = extSig
	e.Version |= types.ExtrinsicBitSigned
	return nil
}