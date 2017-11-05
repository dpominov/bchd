// Copyright (c) 2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bchjson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/bchsuite/bchd/bchjson"
)

// TestWalletSvrWsNtfns tests all of the chain server websocket-specific
// notifications marshal and unmarshal into valid results include handling of
// optional fields being omitted in the marshalled command, while optional
// fields with defaults have the default assigned on unmarshalled commands.
func TestWalletSvrWsNtfns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		newNtfn      func() (interface{}, error)
		staticNtfn   func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "accountbalance",
			newNtfn: func() (interface{}, error) {
				return bchjson.NewCmd("accountbalance", "acct", 1.25, true)
			},
			staticNtfn: func() interface{} {
				return bchjson.NewAccountBalanceNtfn("acct", 1.25, true)
			},
			marshalled: `{"jsonrpc":"1.0","method":"accountbalance","params":["acct",1.25,true],"id":null}`,
			unmarshalled: &bchjson.AccountBalanceNtfn{
				Account:   "acct",
				Balance:   1.25,
				Confirmed: true,
			},
		},
		{
			name: "bchdconnected",
			newNtfn: func() (interface{}, error) {
				return bchjson.NewCmd("bchdconnected", true)
			},
			staticNtfn: func() interface{} {
				return bchjson.NewBchdConnectedNtfn(true)
			},
			marshalled: `{"jsonrpc":"1.0","method":"bchdconnected","params":[true],"id":null}`,
			unmarshalled: &bchjson.BchdConnectedNtfn{
				Connected: true,
			},
		},
		{
			name: "walletlockstate",
			newNtfn: func() (interface{}, error) {
				return bchjson.NewCmd("walletlockstate", true)
			},
			staticNtfn: func() interface{} {
				return bchjson.NewWalletLockStateNtfn(true)
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletlockstate","params":[true],"id":null}`,
			unmarshalled: &bchjson.WalletLockStateNtfn{
				Locked: true,
			},
		},
		{
			name: "newtx",
			newNtfn: func() (interface{}, error) {
				return bchjson.NewCmd("newtx", "acct", `{"account":"acct","address":"1Address","category":"send","amount":1.5,"bip125-replaceable":"unknown","fee":0.0001,"confirmations":1,"trusted":true,"txid":"456","walletconflicts":[],"time":12345678,"timereceived":12345876,"vout":789,"otheraccount":"otheracct"}`)
			},
			staticNtfn: func() interface{} {
				result := bchjson.ListTransactionsResult{
					Abandoned:         false,
					Account:           "acct",
					Address:           "1Address",
					BIP125Replaceable: "unknown",
					Category:          "send",
					Amount:            1.5,
					Fee:               bchjson.Float64(0.0001),
					Confirmations:     1,
					TxID:              "456",
					WalletConflicts:   []string{},
					Time:              12345678,
					TimeReceived:      12345876,
					Trusted:           true,
					Vout:              789,
					OtherAccount:      "otheracct",
				}
				return bchjson.NewNewTxNtfn("acct", result)
			},
			marshalled: `{"jsonrpc":"1.0","method":"newtx","params":["acct",{"abandoned":false,"account":"acct","address":"1Address","amount":1.5,"bip125-replaceable":"unknown","category":"send","confirmations":1,"fee":0.0001,"time":12345678,"timereceived":12345876,"trusted":true,"txid":"456","vout":789,"walletconflicts":[],"otheraccount":"otheracct"}],"id":null}`,
			unmarshalled: &bchjson.NewTxNtfn{
				Account: "acct",
				Details: bchjson.ListTransactionsResult{
					Abandoned:         false,
					Account:           "acct",
					Address:           "1Address",
					BIP125Replaceable: "unknown",
					Category:          "send",
					Amount:            1.5,
					Fee:               bchjson.Float64(0.0001),
					Confirmations:     1,
					TxID:              "456",
					WalletConflicts:   []string{},
					Time:              12345678,
					TimeReceived:      12345876,
					Trusted:           true,
					Vout:              789,
					OtherAccount:      "otheracct",
				},
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the notification as created by the new static
		// creation function.  The ID is nil for notifications.
		marshalled, err := bchjson.MarshalCmd(nil, test.staticNtfn())
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		// Ensure the notification is created without error via the
		// generic new notification creation function.
		cmd, err := test.newNtfn()
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected NewCmd error: %v ",
				i, test.name, err)
		}

		// Marshal the notification as created by the generic new
		// notification creation function.    The ID is nil for
		// notifications.
		marshalled, err = bchjson.MarshalCmd(nil, cmd)
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		var request bchjson.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = bchjson.UnmarshalCmd(&request)
		if err != nil {
			t.Errorf("UnmarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !reflect.DeepEqual(cmd, test.unmarshalled) {
			t.Errorf("Test #%d (%s) unexpected unmarshalled command "+
				"- got %s, want %s", i, test.name,
				fmt.Sprintf("(%T) %+[1]v", cmd),
				fmt.Sprintf("(%T) %+[1]v\n", test.unmarshalled))
			continue
		}
	}
}
