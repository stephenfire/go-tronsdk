package go_tronsdk

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

const (
	DefaultTimeoutSeconds       = 30
	DefaultGetTxIntervalSeconds = 5
)

var (
	ErrTxNotFound       = errors.New("transation not found")
	ErrTxResultNotFound = errors.New("transaction result not found")
)

type TronClient struct {
	httpUrl       string
	http          *HttpClient
	grpcUrl       string
	fullnodeConn  *grpc.ClientConn
	fullnodeGrpc  api.WalletClient
	ethUrl        string
	eth           *ethclient.Client
	chainid       *big.Int
	timeout       time.Duration
	getTxInterval time.Duration
}

func _timeoutRun[T any](ctx context.Context, d time.Duration, f func(context.Context) (T, error)) (t T, err error) {
	cctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	select {
	case <-cctx.Done():
		return t, cctx.Err()
	default:
		return f(cctx)
	}
}

func NewTronClient(cctx context.Context, httpurl, grpcurl, ethurl string,
	timeoutSeconds, getTxIntervalSeconds int64) (tc *TronClient, errr error) {
	c := &TronClient{
		httpUrl: httpurl,
		grpcUrl: grpcurl,
		ethUrl:  ethurl,
	}
	defer func() {
		if errr != nil {
			_ = c.Close()
		}
	}()
	to := int64(DefaultTimeoutSeconds)
	if timeoutSeconds > 0 {
		to = timeoutSeconds
	}
	c.timeout = time.Duration(to) * time.Second
	interval := int64(DefaultGetTxIntervalSeconds)
	if interval > 0 {
		interval = getTxIntervalSeconds
	}
	c.getTxInterval = time.Duration(interval) * time.Second

	c.eth, errr = _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*ethclient.Client, error) {
		return ethclient.DialContext(ctx, ethurl)
	})
	if errr != nil {
		return nil, errr
	}
	c.chainid, errr = _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*big.Int, error) {
		return c.eth.ChainID(ctx)
	})
	if errr != nil {
		return nil, errr
	}

	c.fullnodeConn, errr = _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*grpc.ClientConn, error) {
		return grpc.DialContext(ctx, grpcurl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	})
	if errr != nil {
		return nil, errr
	}
	c.fullnodeGrpc = api.NewWalletClient(c.fullnodeConn)

	c.http = NewHttpClient(httpurl, to)

	return c, nil
}

func (c *TronClient) Close() (err error) {
	if c.fullnodeConn != nil {
		err = c.fullnodeConn.Close()
	}
	if c.http != nil {
		c.http.Close()
	}
	if c.eth != nil {
		c.eth.Close()
	}
	return
}

func (c *TronClient) ChainId() *big.Int {
	return new(big.Int).Set(c.chainid)
}

func (c *TronClient) String() string {
	if c == nil {
		return "TronClient<nil>"
	}
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("TronClient{ChainID:%s Timeout:%s", c.chainid, c.timeout))
	buf.WriteString(fmt.Sprintf(" HTTP:%s", c.httpUrl))
	if c.http != nil {
		buf.WriteString("(CONN)")
	}
	buf.WriteString(fmt.Sprintf(" GRPC:%s", c.grpcUrl))
	if c.fullnodeGrpc != nil {
		buf.WriteString("(CONN)")
	}
	buf.WriteString(fmt.Sprintf(" ETH:%s", c.ethUrl))
	if c.eth != nil {
		buf.WriteString("(CONN)")
	}
	buf.WriteByte('}')
	return buf.String()
}

func (c *TronClient) GetNextMaintenanceTime(cctx context.Context) (time.Time, error) {
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) (time.Time, error) {
		nm, err := c.fullnodeGrpc.GetNextMaintenanceTime(ctx, &api.EmptyMessage{})
		if err != nil {
			return time.Time{}, err
		}
		if nm == nil {
			return time.Time{}, errors.New("empty maintenance time")
		}
		if nm.Num > 9999999999 {
			return time.UnixMilli(nm.Num), nil
		} else {
			return time.Unix(nm.Num, 0), nil
		}
	})
}

func (c *TronClient) WitnessPermissions(ctx context.Context, addr address.Address) (*WitnessPerm, error) {
	return _timeoutRun(ctx, c.timeout, func(cctx context.Context) (*WitnessPerm, error) {
		acc, err := c.fullnodeGrpc.GetAccount(cctx, &core.Account{Address: addr})
		if err != nil {
			return nil, err
		}
		ret := &WitnessPerm{OwnerAddr: addr}
		if acc != nil && acc.WitnessPermission != nil && len(acc.WitnessPermission.Keys) > 0 {
			ret.WitnessAddr = acc.WitnessPermission.Keys[0].Address
		}
		return ret, nil
	})
}

const (
	MaxCommitteeSize           = 27
	ConfirmedSize              = 19
	MaintenanceTimeIntervalKey = "getMaintenanceTimeInterval"
)

func (c *TronClient) ListCommittees(ctx context.Context) ([]*WitnessPerm, error) {
	witnesses, err := _timeoutRun(ctx, c.timeout, func(cctx context.Context) (*api.WitnessList, error) {
		return c.fullnodeGrpc.ListWitnesses(cctx, &api.EmptyMessage{})
	})
	if err != nil {
		return nil, err
	}
	if witnesses == nil || len(witnesses.Witnesses) == 0 {
		return nil, errors.New("no witnesses found")
	}
	var wps []*WitnessPerm
	for i, witness := range witnesses.Witnesses {
		if witness != nil && witness.IsJobs {
			if !address.Address(witness.Address).IsValid() {
				return nil, fmt.Errorf("invalid address of (%d)witness:{Address:%x IsJobs:%t}", i, witness.Address, witness.IsJobs)
			}
			addr := address.Address(witness.Address)
			wp, err := c.WitnessPermissions(ctx, witness.Address)
			if err != nil {
				return nil, fmt.Errorf("get permission for witness %s(%s) failed: %w", addr.Hex(), addr.String(), err)
			}
			wps = append(wps, wp)
		}
	}
	return wps, nil
}

func (c *TronClient) GetMaintenanceTimeInterval(cctx context.Context) (time.Duration, error) {
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) (time.Duration, error) {
		chainparams, err := c.fullnodeGrpc.GetChainParameters(ctx, &api.EmptyMessage{})
		if err != nil {
			return 0, err
		}
		if chainparams != nil && len(chainparams.ChainParameter) > 0 {
			for _, cp := range chainparams.ChainParameter {
				if cp != nil && cp.Key == MaintenanceTimeIntervalKey {
					return time.Duration(cp.Value) * time.Millisecond, nil
				}
			}
		} else {
			return 0, errors.New("no chain parameter found")
		}
		return 0, fmt.Errorf("chain parameters key:%s not found", MaintenanceTimeIntervalKey)
	})
}

func (c *TronClient) GetBlockHeader(cctx context.Context, num int64) (*api.BlockExtention, error) {
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*api.BlockExtention, error) {
		req := &api.BlockReq{
			IdOrNum: fmt.Sprintf("%d", num),
			Detail:  false,
		}
		return c.fullnodeGrpc.GetBlock(ctx, req)
	})
}

func (c *TronClient) GetBlock(cctx context.Context, num int64) (*api.BlockExtention, error) {
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*api.BlockExtention, error) {
		req := &api.BlockReq{
			IdOrNum: fmt.Sprintf("%d", num),
			Detail:  true,
		}
		return c.fullnodeGrpc.GetBlock(ctx, req)
	})
}

func (c *TronClient) GetNowBlock(cctx context.Context) (*api.BlockExtention, error) {
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*api.BlockExtention, error) {
		return c.fullnodeGrpc.GetNowBlock2(ctx, &api.EmptyMessage{})
	})
}

// [start, end)
func (c *TronClient) GetBlocks(cctx context.Context, start, end int64) ([]*api.BlockExtention, error) {
	if start < 0 || end < 0 || start >= end {
		return nil, errors.New("invalid range")
	}
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) ([]*api.BlockExtention, error) {
		list, err := c.fullnodeGrpc.GetBlockByLimitNext2(ctx, &api.BlockLimit{
			StartNum: start,
			EndNum:   end,
		})
		if err != nil {
			return nil, err
		}
		if list == nil {
			return nil, nil
		}
		return list.Block, nil
	})
}

// [from, to]
func (c *TronClient) FilterLogs(cctx context.Context, from, to int64, addr []byte, topicss ...[][]byte) ([]types.Log, error) {
	var tss [][]ethcommon.Hash
	for _, topics := range topicss {
		if len(topics) > 0 {
			var ts []ethcommon.Hash
			for _, topic := range topics {
				ts = append(ts, ethcommon.BytesToHash(topic))
			}
		}
	}
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(from),
		ToBlock:   big.NewInt(to),
		Addresses: []ethcommon.Address{ethcommon.BytesToAddress(addr)},
		Topics:    tss,
	}
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) ([]types.Log, error) {
		return c.eth.FilterLogs(ctx, query)
	})
}

func (c *TronClient) GetTransactionById(cctx context.Context, txHash []byte) (*core.Transaction, error) {
	return _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*core.Transaction, error) {
		return c.fullnodeGrpc.GetTransactionById(ctx, &api.BytesMessage{Value: txHash})
	})
}

func (c *TronClient) _hash(m proto.Message) ([]byte, error) {
	bs, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	hasher := sha256.New()
	hasher.Write(bs)
	return hasher.Sum(nil), nil
}

func (c *TronClient) CallContract(cctx context.Context, from, contract address.Address, data []byte) (*api.TransactionExtention, error) {
	txx, err := _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*api.TransactionExtention, error) {
		return c.fullnodeGrpc.TriggerConstantContract(ctx, &core.TriggerSmartContract{
			OwnerAddress:    from,
			ContractAddress: contract,
			CallValue:       0,
			Data:            data,
			CallTokenValue:  0,
			TokenId:         0,
		})
	})
	if err != nil {
		return nil, err
	}
	return txx, nil
}

func (c *TronClient) TriggerContract(cctx context.Context, feeLimit int64,
	fromPriv []byte, contract address.Address, data []byte) ([]byte, error) {
	privKey, err := BytesToPrivateKey(fromPriv)
	if err != nil || privKey == nil {
		return nil, errors.New("unknown private key")
	}
	ethfrom := crypto.PubkeyToAddress(privKey.PublicKey)
	from := address.BytesToAddress(ethfrom[:])
	tsc := &core.TriggerSmartContract{
		OwnerAddress:    from[:],
		ContractAddress: contract[:],
		CallValue:       0,
		Data:            data,
		CallTokenValue:  0,
		TokenId:         0,
	}
	txx, err := _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*api.TransactionExtention, error) {
		return c.fullnodeGrpc.TriggerContract(ctx, tsc)
	})
	if txx.Result.Code > 0 {
		return nil, fmt.Errorf("%s", string(txx.Result.Message))
	}
	if feeLimit > 0 {
		txx.Transaction.RawData.FeeLimit = feeLimit
	}
	txId, err := c._hash(txx.Transaction.RawData)
	if err != nil {
		return nil, err
	}
	txx.Txid = txId

	sig, err := crypto.Sign(txId, privKey)
	if err != nil {
		return nil, err
	}
	txx.Transaction.Signature = append(txx.Transaction.Signature, sig)

	ret, err := _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*api.Return, error) {
		return c.fullnodeGrpc.BroadcastTransaction(ctx, txx.Transaction)
	})
	if err != nil {
		return nil, err
	}
	if err = c.ParseReturn(ret); err != nil {
		return nil, fmt.Errorf("broadcast failed: %w", err)
	}
	return txId, nil
}

func (c *TronClient) ParseReturn(ret *api.Return) error {
	if ret != nil && (!ret.Result || ret.Code > 0) {
		return fmt.Errorf("result(%d) %s", ret.Code, string(ret.Message))
	}
	return nil
}

func (c *TronClient) TriggerContractResult(cctx context.Context, txId []byte) (*core.Transaction, error) {
	tx, err := _timeoutRun(cctx, c.timeout, func(ctx context.Context) (*core.Transaction, error) {
		return c.fullnodeGrpc.GetTransactionById(ctx, &api.BytesMessage{Value: txId})
	})
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, nil
	}
	id, _ := c._hash(tx.RawData)
	if bytes.Equal(id, txId) {
		return tx, nil
	}
	return nil, errors.New("not found")
}

func (c *TronClient) TryTxByHash(cctx context.Context, txId []byte) (*core.Transaction, error) {
	for i := 0; i < 5; i++ {
		time.Sleep(c.getTxInterval)
		select {
		case <-cctx.Done():
			return nil, cctx.Err()
		default:
			tx, err := c.TriggerContractResult(cctx, txId)
			if err != nil || tx == nil {
				continue
			}
			if tx != nil {
				return tx, nil
			}
		}
	}
	return nil, ErrTxNotFound
}

func (c *TronClient) ParseContractTxExResult(txx *api.TransactionExtention, runErr error) (energy int64, output []byte, err error) {
	if runErr != nil {
		return 0, nil, runErr
	}
	if txx == nil || txx.Transaction == nil {
		return 0, nil, ErrTxNotFound
	}
	if err := c.ParseReturn(txx.Result); err != nil {
		return 0, nil, err
	}
	if _, err := c.ContractTxResult(txx.Transaction, false); err != nil {
		return 0, nil, err
	}
	if len(txx.ConstantResult) > 0 {
		output = txx.ConstantResult[0]
	}
	return txx.EnergyUsed - txx.EnergyPenalty, output, nil
}

func (c *TronClient) ParseContractTxResult(tx *core.Transaction, runErr error) (fee int64, err error) {
	if runErr != nil {
		return 0, runErr
	}
	return c.ContractTxResult(tx, true)
}

func (c *TronClient) ContractTxResult(tx *core.Transaction, runOrCall bool) (fee int64, err error) {
	if tx == nil {
		return 0, ErrTxNotFound
	}
	if len(tx.Ret) == 0 || tx.Ret[0] == nil {
		return 0, ErrTxResultNotFound
	}
	if runOrCall {
		if tx.Ret[0].Ret == core.Transaction_Result_SUCESS && tx.Ret[0].ContractRet == core.Transaction_Result_SUCCESS {
			return tx.Ret[0].Fee, nil
		}
	} else {
		if tx.Ret[0].Ret == core.Transaction_Result_SUCESS && tx.Ret[0].ContractRet <= core.Transaction_Result_SUCCESS {
			return tx.Ret[0].Fee, nil
		}
	}
	return 0, fmt.Errorf("ret:%s contractRet:%s", tx.Ret[0].Ret.String(), tx.Ret[0].ContractRet.String())
}
