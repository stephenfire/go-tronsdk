package go_tronsdk

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

var (
	ErrInvalidTx          = errors.New("invalid transaction")
	ErrNotTriggerContract = errors.New("not trigger contract")
	ErrTxNotFound         = errors.New("transation not found")
	ErrTxResultNotFound   = errors.New("transaction result not found")
)

type Receipt struct {
	BlockNum           int64
	Timestamp          int64
	TxId               []byte
	From               []byte
	To                 []byte
	Input              []byte
	Output             []byte
	EnergyUsage        int64
	EnergyPenaltyTotal int64
	EnergyTotal        int64
	EnergyFee          int64
	OriginEnergyUsage  int64
	NetUsage           int64
	NetFee             int64
	Logs               []*core.TransactionInfo_Log
	Succeed            bool
	Err                error
}

type Tx core.Transaction

func (t *Tx) MerkleHash() ([]byte, error) {
	if t == nil {
		return nil, ErrInvalidTx
	}
	return HashMessage((*core.Transaction)(t))
}

func (t *Tx) IsTriggerSmartContract() (*core.TriggerSmartContract, error) {
	if t == nil || t.RawData == nil || len(t.RawData.Contract) == 0 {
		return nil, ErrNotTriggerContract
	}
	for _, contract := range t.RawData.Contract {
		switch contract.Type {
		case core.Transaction_Contract_TriggerSmartContract:
			tsc := new(core.TriggerSmartContract)
			err := contract.Parameter.UnmarshalTo(tsc)
			if err != nil {
				return nil, err
			}
			return tsc, nil
		}
	}
	return nil, ErrNotTriggerContract
}

// ToResult mustSuccess default as false, which means Transaction_Result_DEFAULT result is ok.
// set to true, ContractRet must be Transaction_Result_UCCESS
func (t *Tx) ToResult(mustSuccess ...bool) (fee int64, err error) {
	if t == nil {
		return 0, ErrTxNotFound
	}
	if len(t.Ret) == 0 || t.Ret[0] == nil {
		return 0, ErrTxResultNotFound
	}
	if len(mustSuccess) > 0 && mustSuccess[0] {
		if t.Ret[0].Ret == core.Transaction_Result_SUCESS && t.Ret[0].ContractRet == core.Transaction_Result_SUCCESS {
			return t.Ret[0].Fee, nil
		}
	} else {
		if t.Ret[0].Ret == core.Transaction_Result_SUCESS && t.Ret[0].ContractRet <= core.Transaction_Result_SUCCESS {
			return t.Ret[0].Fee, nil
		}
	}
	return 0, fmt.Errorf("ret:%s contractRet:%s", t.Ret[0].Ret.String(), t.Ret[0].ContractRet.String())
}

type TxInfo core.TransactionInfo

func (i *TxInfo) ToReceipt() (*Receipt, error) {
	if i == nil {
		return nil, errors.New("no tx info found")
	}
	rpt := new(Receipt)
	rpt.BlockNum = i.BlockNumber
	rpt.Timestamp = i.BlockTimeStamp
	rpt.TxId = common.CopyBytes(i.Id)
	rpt.To = common.CopyBytes(i.ContractAddress)
	if len(i.ContractResult) > 0 {
		rpt.Output = i.ContractResult[0]
	}
	if len(i.Log) > 0 {
		rpt.Logs = i.Log
	}
	if i.Receipt != nil {
		rpt.EnergyTotal = i.Receipt.EnergyUsageTotal
		rpt.EnergyPenaltyTotal = i.Receipt.EnergyPenaltyTotal
		rpt.OriginEnergyUsage = i.Receipt.OriginEnergyUsage
		rpt.EnergyUsage = i.Receipt.EnergyUsage
		rpt.EnergyFee = i.Receipt.EnergyFee
		rpt.NetUsage = i.Receipt.NetUsage
		rpt.NetFee = i.Receipt.NetFee
	}
	if i.Result != core.TransactionInfo_SUCESS ||
		(i.Receipt != nil && i.Receipt.Result > core.Transaction_Result_SUCCESS) {
		if i.Receipt != nil {
			rpt.Err = fmt.Errorf("%s contractRet:%s", i.Result.String(), i.Receipt.Result.String())
			return rpt, nil
		} else {
			rpt.Err = fmt.Errorf("%s", i.Result.String())
			return rpt, nil
		}
	}
	rpt.Succeed = true
	return rpt, nil
}

type TxReturn api.Return

func (ret *TxReturn) Err() error {
	if ret != nil && (!ret.Result || ret.Code > 0) {
		return fmt.Errorf("result(%d) %s", ret.Code, string(ret.Message))
	}
	return nil
}

type TxEx api.TransactionExtention

func (txx *TxEx) ToResult(mustSuccess ...bool) (energy int64, output []byte, err error) {
	if txx == nil || txx.Transaction == nil {
		return 0, nil, ErrTxNotFound
	}
	if err := (*TxReturn)(txx.Result).Err(); err != nil {
		return 0, nil, err
	}
	// if _, err := c.ContractTxResult(txx.Transaction, false); err != nil {
	if _, err := (*Tx)(txx.Transaction).ToResult(mustSuccess...); err != nil {
		return 0, nil, err
	}
	if len(txx.ConstantResult) > 0 {
		output = txx.ConstantResult[0]
	}
	return txx.EnergyUsed - txx.EnergyPenalty, output, nil
}
