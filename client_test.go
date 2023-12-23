package go_tronsdk

import (
	"context"
	"encoding/hex"
	"testing"
)

func TestTronClient_FilterLogs(t *testing.T) {
	// client, err := NewTronClient(context.Background(), "https://api.nileex.io/", "grpc.nile.trongrid.io:50051", "https://nile.trongrid.io/jsonrpc", 10, 5)
	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	addr, _ := hex.DecodeString("4c290fab628c32d932c7f48f8f533928e33ae13c")
	topic, _ := hex.DecodeString("44ff77018688dad4b245e8ab97358ed57ed92269952ece7ffd321366ce078622")
	logs, err := client.FilterLogs(context.Background(), 39945134, 39945155, addr, [][]byte{topic})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(logs)
}

func TestTronClient_GetTransactionById(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.nileex.io/", "grpc.nile.trongrid.io:50051", "https://nile.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	txId, _ := hex.DecodeString("e1ec2d4ea861c4b07b8470d468a63a096e65f0c2011cd24ab35f74fa3eb51bc6")
	tx, err := client.GetTransactionById(context.Background(), txId)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tx)
}

// func TestTriggerContract(t *testing.T) {
// 	priv, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
// 	cbs, _ := hex.DecodeString("4127abb48702183bdf4672fb40808ef8ebfbb40602")
// 	data, _ := hex.DecodeString("981aff4e000000000000000000000000da50a11cd3fd79d638f3552257c0f6f27d239af7000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000186a000000000000000000000000000000000000000000000000000000000000001bc0000000000000000000000000000000000000000000000000000000000000014da50a11cd3fd79d638f3552257c0f6f27d239af7000000000000000000000000")
//
// 	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer func() {
// 		_ = client.Close()
// 	}()
//
// 	txx, err := client.TriggerContract(context.Background(), 10000000000, priv, cbs, 0, data)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Logf("%x sent", txx.Txid)
// 	fee, err := client.ParseContractTxResult(client.TryTxByHash(context.Background(), txx.Txid))
// 	if err != nil {
// 		t.Fatalf("%x failed: %v", txx.Txid, err)
// 	} else {
// 		t.Logf("fee: %d", fee)
// 	}
// }

func TestGetTxResult(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()
	txId, _ := hex.DecodeString("ba391614fbd4e2cac2f0da10445ce6c79573413e077ee70d84c1a2008d3e8d5d")
	// txId, _ := hex.DecodeString("31e54edd14b3b6acd5d26fddd7e42ea44c22e708d366a2caa94fdccd7beccab8")
	fee, err := ParseContractTxResult(client.TryTxByHash(context.Background(), txId))
	if err != nil {
		t.Fatalf("%x failed: %v", txId, err)
	} else {
		t.Logf("fee: %d", fee)
	}
}

func TestTriggerConstantContract(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	//
	from, _ := hex.DecodeString("410000000000000000000000000000000000000000")
	// contract, _ := hex.DecodeString("413de9ab8f268ae3ee3cb7908db7249ba173e15c0f")
	// data, _ := hex.DecodeString("6e9960c3")

	// OrderTest.sol
	contract, _ := hex.DecodeString("41878c79fc51056162b69b07aa1d8695e1e5a63326")
	// (953fe6a7) function iOrder() view returns(address)
	data, _ := hex.DecodeString("953fe6a7")
	tx, err := client.CallContract(context.Background(), from, contract, data)
	if err != nil {
		t.Fatal(err)
	}
	energy, output, err := (*TxEx)(tx).ToResult()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("energy:%d output:%x", energy, output)
}

func TestTronClient_GetMaintenanceTimeInterval(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()
	mtiTime, err := client.GetMaintenanceTimeInterval(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s, %dms", mtiTime, mtiTime.Milliseconds())
}

func TestTronClient_GetTransactionInfoById(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	txId, _ := hex.DecodeString("4c4368c99f1d5575643b0b57a5058189522f2800c84d3fb04d915d251f0ae56c")
	tx, err := client.GetTransactionInfoById(context.Background(), txId)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tx)
}

func TestTronClient_GetContract(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	addr, _ := hex.DecodeString("41878c79fc51056162b69b07aa1d8695e1e5a63326")
	contract, err := client.GetContract(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(contract)
}

func TestTronClient_GetAccount(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.shasta.trongrid.io", "grpc.shasta.trongrid.io:50051", "https://api.shasta.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()
	addr, _ := hex.DecodeString("41878c79fc51056162b69b07aa1d8695e1e5a63326")
	acc, err := client.GetAccount(context.Background(), addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(acc)
}
