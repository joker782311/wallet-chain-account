package solana

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"math"
	"strconv"
	"strings"
	"time"

	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/mr-tron/base58"

	account2 "github.com/dapplink-labs/chain-explorer-api/common/account"
	"github.com/dapplink-labs/wallet-chain-account/chain"
	"github.com/dapplink-labs/wallet-chain-account/config"
	"github.com/dapplink-labs/wallet-chain-account/rpc/account"
	common2 "github.com/dapplink-labs/wallet-chain-account/rpc/common"
)

const ChainName = "Solana"

type ChainAdaptor struct {
	solCli  SolClient
	solData *SolData
}

func NewChainAdaptor(conf *config.Config) (chain.IChainAdaptor, error) {
	rpcUrl := conf.WalletNode.Sol.RpcUrl
	cli, err := NewSolHttpClient(rpcUrl)
	if err != nil {
		return nil, err
	}
	sol, err := NewSolScanClient(conf.WalletNode.Sol.DataApiUrl, conf.WalletNode.Sol.DataApiKey, time.Duration(conf.WalletNode.Sol.TimeOut))
	if err != nil {
		return nil, err
	}
	return &ChainAdaptor{
		solCli:  cli,
		solData: sol,
	}, nil

}

func (c *ChainAdaptor) GetSupportChains(req *account.SupportChainsRequest) (*account.SupportChainsResponse, error) {
	response := &account.SupportChainsResponse{
		Code:    common2.ReturnCode_ERROR,
		Msg:     "",
		Support: false,
	}
	if ok, msg := validateChainAndNetwork(req.Chain, req.Network); !ok {
		err := fmt.Errorf("GetSupportChains validateChainAndNetwork fail, err msg = %s", msg)
		log.Error("err", err)
		response.Msg = err.Error()
		return response, err
	}

	response.Msg = "Support this chain"
	response.Code = common2.ReturnCode_SUCCESS
	response.Support = true
	return response, nil
}

func (c *ChainAdaptor) ConvertAddress(req *account.ConvertAddressRequest) (*account.ConvertAddressResponse, error) {
	response := &account.ConvertAddressResponse{
		Code:    common2.ReturnCode_ERROR,
		Msg:     "",
		Address: "",
	}
	if ok, msg := validateChainAndNetwork(req.Chain, req.Network); !ok {
		err := fmt.Errorf("ConvertAddress validateChainAndNetwork fail, err msg = %s", msg)
		log.Error("err", err)
		response.Msg = err.Error()
		return nil, err
	}
	pubKeyHex := req.PublicKey
	if ok, msg := validatePublicKey(pubKeyHex); !ok {
		err := fmt.Errorf("ConvertAddress validatePublicKey fail, err msg = %s", msg)
		log.Error("err", err)
		response.Msg = err.Error()
		return nil, err
	}
	accountAddress, err := PubKeyHexToAddress(pubKeyHex)
	if err != nil {
		err := fmt.Errorf("ConvertAddress PubKeyHexToAddress failed: %w", err)
		log.Error("err", err)
		response.Msg = err.Error()
		return nil, err
	}
	response.Code = common2.ReturnCode_SUCCESS
	response.Msg = "convert address success"
	response.Address = accountAddress
	return response, nil
}

func (c *ChainAdaptor) ValidAddress(req *account.ValidAddressRequest) (*account.ValidAddressResponse, error) {
	response := &account.ValidAddressResponse{
		Code:  common2.ReturnCode_ERROR,
		Msg:   "",
		Valid: false,
	}

	if ok, msg := validateChainAndNetwork(req.Chain, req.Network); !ok {
		err := fmt.Errorf("ConvertAddress validateChainAndNetwork failed: %s", msg)
		log.Error("err", err)
		response.Msg = err.Error()
		return response, err
	}
	address := req.Address
	if len(address) == 0 {
		err := fmt.Errorf("ConvertAddress address is empty")
		log.Error("err", err)
		response.Msg = err.Error()
		return response, err
	}
	if len(address) != 43 && len(address) != 44 {
		err := fmt.Errorf("invalid Solana address length: expected 43 or 44 characters, got %d", len(address))
		response.Msg = err.Error()
		return response, err
	}
	response.Code = common2.ReturnCode_SUCCESS
	response.Valid = true
	return response, nil
}

func (c *ChainAdaptor) GetBlockByNumber(req *account.BlockNumberRequest) (*account.BlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) GetBlockByHash(req *account.BlockHashRequest) (*account.BlockResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) GetBlockHeaderByHash(req *account.BlockHeaderHashRequest) (*account.BlockHeaderResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) GetBlockHeaderByNumber(req *account.BlockHeaderNumberRequest) (*account.BlockHeaderResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) GetAccount(req *account.AccountRequest) (*account.AccountResponse, error) {
	//req.ContractAddress as nonceAddress
	//nonceResult, err := c.solCli.GetNonce(req.ContractAddress)
	//if err != nil {
	//	log.Error("get nonce by address fail", "err", err)
	//	return &account.AccountResponse{
	//		Code: common2.ReturnCode_ERROR,
	//		Msg:  "get nonce by address fail",
	//	}, nil
	//}
	balanceResult, err := c.solCli.GetBalance(req.Address)
	if err != nil {
		return &account.AccountResponse{
			Code:    common2.ReturnCode_ERROR,
			Msg:     "get token balance fail",
			Balance: "0",
		}, err
	}
	log.Info("balance result", "balance=", balanceResult, "balanceStr=", balanceResult)
	return &account.AccountResponse{
		Code:          common2.ReturnCode_SUCCESS,
		Msg:           "get account response success",
		AccountNumber: "0",
		//Sequence:      nonceResult,
		//Balance:       balanceResult,
	}, nil
}

func (c *ChainAdaptor) GetFee(req *account.FeeRequest) (*account.FeeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) SendTx(req *account.SendTxRequest) (*account.SendTxResponse, error) {
	//TxResponse, err := c.solCli.SendTx(req.RawTx)
	//if err != nil {
	//	return &account.SendTxResponse{
	//		Code:   common2.ReturnCode_ERROR,
	//		Msg:    "get tx response error",
	//		TxHash: "0",
	//	}, nil
	//}
	return &account.SendTxResponse{
		Code: common2.ReturnCode_SUCCESS,
		Msg:  "get tx response success",
		//TxHash: TxResponse,
	}, nil
}

func (c *ChainAdaptor) GetTxByAddress(req *account.TxAddressRequest) (*account.TxAddressResponse, error) {
	var resp *account2.TransactionResponse[account2.AccountTxResponse]
	var err error
	fmt.Println("req.ContractAddress", req.ContractAddress)
	if req.ContractAddress != "0x00" && req.ContractAddress != "" {
		log.Info("Spl token transfer record")
		resp, err = c.solData.GetTxByAddress(uint64(req.Page), uint64(req.Pagesize), req.Address, "spl")
	} else {
		log.Info("Sol transfer record")
		resp, err = c.solData.GetTxByAddress(uint64(req.Page), uint64(req.Pagesize), req.Address, "sol")
	}
	if err != nil {
		log.Error("get GetTxByAddress error", "err", err)
		return &account.TxAddressResponse{
			Code: common2.ReturnCode_ERROR,
			Msg:  "get tx list fail",
			Tx:   nil,
		}, err
	} else {
		txs := resp.TransactionList
		list := make([]*account.TxMessage, 0, len(txs))
		for i := 0; i < len(txs); i++ {
			list = append(list, &account.TxMessage{
				Hash:   txs[i].TxId,
				Tos:    []*account.Address{{Address: txs[i].To}},
				Froms:  []*account.Address{{Address: txs[i].From}},
				Fee:    txs[i].TxId,
				Status: account.TxStatus_Success,
				Values: []*account.Value{{Value: txs[i].Amount}},
				Type:   1,
				Height: txs[i].Height,
			})
		}
		fmt.Println("resp", resp)
		return &account.TxAddressResponse{
			Code: common2.ReturnCode_SUCCESS,
			Msg:  "get tx list success",
			Tx:   list,
		}, nil
	}
}

func (c *ChainAdaptor) GetTxByHash(req *account.TxHashRequest) (*account.TxHashResponse, error) {
	//tx, err := c.solCli.GetTxByHash(req.Hash)
	//if err != nil {
	//	return &account.TxHashResponse{
	//		Code: common2.ReturnCode_ERROR,
	//		Msg:  err.Error(),
	//		Tx:   nil,
	//	}, err
	//}
	//var value_list []*account.Value
	//value_list = append(value_list, &account.Value{Value: tx.Value})
	return &account.TxHashResponse{
		Tx: &account.TxMessage{
			//Hash:  tx.Hash,
			//Tos:   []*account.Address{{Address: tx.To}},
			//Froms: []*account.Address{{Address: tx.From}},

			//Fee:    tx.Fee,
			//Status: account.TxStatus_Success,
			//Values: value_list,
			//Type:   tx.Type,
			//Height: tx.Height,
		},
	}, nil
}

func (c *ChainAdaptor) GetBlockByRange(req *account.BlockByRangeRequest) (*account.BlockByRangeResponse, error) {
	//TODO implement me
	panic("implement me")
}
func (c *ChainAdaptor) CreateUnSignTransaction(req *account.UnSignTransactionRequest) (*account.UnSignTransactionResponse, error) {

	jsonBytes, err := base64.StdEncoding.DecodeString(req.Base64Tx)
	if err != nil {
		log.Error("decode string fail", "err", err)
		return nil, err
	}
	var data TxStructure
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		log.Error("parse json fail", "err", err)
		return nil, err
	}
	valueFloat, err := strconv.ParseFloat(data.Value, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value: %w", err)
	}
	value := uint64(valueFloat * 1000000000)
	if err != nil {
		return nil, err
	}

	fromPrikey, err := getPrivateKey(data.FromPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	fromPubkey, err := solana.PublicKeyFromBase58(data.FromAddress)
	if err != nil {
		return nil, err
	}
	toPubkey, err := solana.PublicKeyFromBase58(data.ToAddress)
	if err != nil {
		return nil, err
	}
	var tx *solana.Transaction
	if isSOLTransfer(data.ContractAddress) {
		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					value,
					fromPubkey,
					toPubkey,
				).Build(),
			},
			solana.MustHashFromBase58(data.Nonce),
			solana.TransactionPayer(fromPubkey),
		)

	} else {

		mintPubkey := solana.MustPublicKeyFromBase58(data.ContractAddress)

		fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			fromPubkey,
			mintPubkey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to FindAssociatedTokenAddress: %w", err)
		}

		toTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			toPubkey,
			mintPubkey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to FindAssociatedTokenAddress: %w", err)
		}

		tokenInfo, err := c.solCli.Client.GetTokenSupply(context.Background(), data.ContractAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get token info: %w", err)
		}
		decimals := tokenInfo.Decimals

		valueFloat, err := strconv.ParseFloat(data.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value: %w", err)
		}
		actualValue := uint64(valueFloat * math.Pow10(int(decimals)))

		transferInstruction := token.NewTransferInstruction(
			actualValue,
			fromTokenAccount,
			toTokenAccount,
			fromPubkey,
			[]solana.PublicKey{},
		).Build()

		accountInfo, err := c.solCli.Client.GetAccountInfo(context.Background(), toTokenAccount.String())
		if err != nil || accountInfo.Data == nil {

			createATAInstruction := associatedtokenaccount.NewCreateInstruction(
				fromPubkey,
				toPubkey,
				mintPubkey,
			).Build()

			tx, err = solana.NewTransaction(
				[]solana.Instruction{createATAInstruction, transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubkey),
			)
		} else {
			// 直接创建转账交易
			tx, err = solana.NewTransaction(
				[]solana.Instruction{transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubkey),
			)
		}
	}

	//https://github.com/gagliardetto/solana-go/tree/main?tab=readme-ov-file#transfer-sol-from-one-wallet-to-another-wallet
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			return &fromPrikey
		},
	)

	//	signedTx := tx.String()
	spew.Dump(tx)
	if err := tx.VerifySignatures(); err != nil {
		log.Info("Invalid signatures: %v", err)
	}
	//https://github.com/gagliardetto/solana-go/tree/main?tab=readme-ov-file#transfer-sol-from-one-wallet-to-another-wallet
	return &account.UnSignTransactionResponse{
		Code:     common2.ReturnCode_SUCCESS,
		Msg:      "create un sign tx success",
		UnSignTx: tx.String(),
	}, nil
}
func (c ChainAdaptor) BuildSignedTransaction(req *account.SignedTransactionRequest) (*account.SignedTransactionResponse, error) {
	jsonBytes, err := base64.StdEncoding.DecodeString(req.Base64Tx)
	if err != nil {
		log.Error("decode string fail", "err", err)
		return nil, err
	}
	var data TxStructure
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		log.Error("parse json fail", "err", err)
		return nil, err
	}
	valueFloat, err := strconv.ParseFloat(data.Value, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value: %w", err)
	}
	value := uint64(valueFloat * 1000000000)
	if err != nil {
		return nil, err
	}

	fromPrikey, err := getPrivateKey(data.FromPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	fromPubkey, err := solana.PublicKeyFromBase58(data.FromAddress)
	if err != nil {
		return nil, err
	}
	toPubkey, err := solana.PublicKeyFromBase58(data.ToAddress)
	if err != nil {
		return nil, err
	}
	var tx *solana.Transaction
	if isSOLTransfer(data.ContractAddress) {
		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					value,
					fromPubkey,
					toPubkey,
				).Build(),
			},
			solana.MustHashFromBase58(data.Nonce),
			solana.TransactionPayer(fromPubkey),
		)

	} else {

		mintPubkey := solana.MustPublicKeyFromBase58(data.ContractAddress)

		fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			fromPubkey,
			mintPubkey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to FindAssociatedTokenAddress: %w", err)
		}

		toTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			toPubkey,
			mintPubkey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to FindAssociatedTokenAddress: %w", err)
		}

		tokenInfo, err := c.solCli.Client.GetTokenSupply(context.Background(), data.ContractAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get token info: %w", err)
		}
		decimals := tokenInfo.Decimals

		valueFloat, err := strconv.ParseFloat(data.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value: %w", err)
		}
		actualValue := uint64(valueFloat * math.Pow10(int(decimals)))

		transferInstruction := token.NewTransferInstruction(
			actualValue,
			fromTokenAccount,
			toTokenAccount,
			fromPubkey,
			[]solana.PublicKey{},
		).Build()

		accountInfo, err := c.solCli.Client.GetAccountInfo(context.Background(), toTokenAccount.String())
		if err != nil || accountInfo.Data == nil {

			createATAInstruction := associatedtokenaccount.NewCreateInstruction(
				fromPubkey,
				toPubkey,
				mintPubkey,
			).Build()

			tx, err = solana.NewTransaction(
				[]solana.Instruction{createATAInstruction, transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubkey),
			)
		} else {

			tx, err = solana.NewTransaction(
				[]solana.Instruction{transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubkey),
			)
		}
	}

	//https://github.com/gagliardetto/solana-go/tree/main?tab=readme-ov-file#transfer-sol-from-one-wallet-to-another-wallet
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			return &fromPrikey
		},
	)

	//	signedTx := tx.String()
	spew.Dump(tx)
	if err := tx.VerifySignatures(); err != nil {
		log.Info("Invalid signatures: %v", err)
	}
	serializedTx, err := tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}
	encodedTx := base64.StdEncoding.EncodeToString(serializedTx)

	binaryTx, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		return nil, fmt.Errorf("Base64 decode error: %w", err)
	}

	base58Tx := base58.Encode(binaryTx)
	return &account.SignedTransactionResponse{
		Code:     common2.ReturnCode_SUCCESS,
		Msg:      "create un sign tx success",
		SignedTx: base58Tx,
	}, nil
}

func (c *ChainAdaptor) DecodeTransaction(req *account.DecodeTransactionRequest) (*account.DecodeTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ChainAdaptor) VerifySignedTransaction(req *account.VerifyTransactionRequest) (*account.VerifyTransactionResponse, error) {

	txBytes, err := base58.Decode(req.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	tx, err := solana.TransactionFromBytes(txBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %w", err)
	}

	if err := tx.VerifySignatures(); err != nil {
		log.Info("Invalid signatures", "err", err)
		return &account.VerifyTransactionResponse{
			Code:   common2.ReturnCode_ERROR,
			Msg:    "invalid signature",
			Verify: false,
		}, nil
	}

	return &account.VerifyTransactionResponse{
		Code:   common2.ReturnCode_SUCCESS,
		Msg:    "valid signature",
		Verify: true,
	}, nil
}

func (c *ChainAdaptor) GetExtraData(req *account.ExtraDataRequest) (*account.ExtraDataResponse, error) {
	//TODO implement me
	panic("implement me")
}
func isSOLTransfer(coinAddress string) bool {

	return coinAddress == "" ||
		coinAddress == "So11111111111111111111111111111111111111112"
}
func getPrivateKey(keyStr string) (solana.PrivateKey, error) {
	// base58
	if prikey, err := solana.PrivateKeyFromBase58(keyStr); err == nil {
		return prikey, nil
	}

	// 16
	if bytes, err := hex.DecodeString(keyStr); err == nil {
		return solana.PrivateKey(bytes), nil
	}

	return nil, fmt.Errorf("invalid private key format")
}

func validateChainAndNetwork(chain, network string) (bool, string) {
	if chain != ChainName {
		return false, "invalid chain"
	}
	//if network != NetworkMainnet && network != NetworkTestnet {
	//	return false, "invalid network"
	//}
	return true, ""
}

func validatePublicKey(pubKey string) (bool, string) {
	if pubKey == "" {
		return false, "public key cannot be empty"
	}
	pubKeyWithoutPrefix := strings.TrimPrefix(pubKey, "0x")

	if len(pubKeyWithoutPrefix) != 64 {
		return false, "invalid public key length"
	}
	if _, err := hex.DecodeString(pubKeyWithoutPrefix); err != nil {
		return false, "invalid public key format: must be hex string"
	}

	return true, ""
}
