package go_tronsdk

import (
	"context"
	"encoding/hex"
	"testing"
)

func TestTronClient_FilterLogs(t *testing.T) {
	client, err := NewTronClient(context.Background(), "https://api.nileex.io/", "grpc.nile.trongrid.io:50051", "https://nile.trongrid.io/jsonrpc", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

	addr, _ := hex.DecodeString("eca9bc828a3005b9a3b909f2cc5c2a54794de05f")
	topic, _ := hex.DecodeString("ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	logs, err := client.FilterLogs(context.Background(), 41814304, 41814304, addr, [][]byte{topic})
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
// 	txId, err := client.TriggerContract(context.Background(), 10000000000, priv, cbs, data)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Logf("%x sent", txId)
// 	fee, err := client.ParseContractTxResult(client.TryTxByHash(context.Background(), txId))
// 	if err != nil {
// 		t.Fatalf("%x failed: %v", txId, err)
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
	fee, err := client.ParseContractTxResult(client.TryTxByHash(context.Background(), txId))
	if err != nil {
		t.Fatalf("%x failed: %v", txId, err)
	} else {
		t.Logf("fee: %d", fee)
	}
}
