package modules

// crust build/integration-tests
// ./coreum-modules -chain-id=coreum-mainnet-1 -test.run=TestValidatorGrant > multisig.json

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
)

func TestValidatorGrants(t *testing.T) {
	const amount = 20_300_000_000
	recipients := []string{
		"core1xprcq3xdcuht0a8p082l3srgtwfgl57hsrj7s7",
		"core10zh7zme4m3ash43k5cxz9ztq0refcpvemj0gaa",
		"core1sg0gwgwumymhp0acldjpz7s3k9trgcnndqw593",
	}

	requireT := require.New(t)

	_, chain := integrationtests.NewTestingContext(t)
	codec := chain.ClientContext.Codec()
	keyring := chain.ClientContext.Keyring()

	pubKey1 := &secp256k1.PubKey{}
	pubKey2 := &secp256k1.PubKey{}
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"A2MidxM8OUyemp7UycIVNDR2YxEomyfEYtndydqIuBsV"}`), pubKey1))
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"A3sqCkVWPIHZ64JWwd3rM7Qxj2vjsDfoOJ+JLn7bP4ge"}`), pubKey2))

	multisigKey := multisig.NewLegacyAminoPubKey(
		2,
		[]types.PubKey{pubKey1, pubKey2},
	)
	multisigInfo, err := keyring.SaveMultisig("coreum-foundation-0", multisigKey)
	requireT.NoError(err)

	multisendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: multisigInfo.GetAddress().String(),
				Coins:   sdk.NewCoins(chain.NewCoin(sdk.NewIntFromUint64(amount).MulRaw(int64(len(recipients))))),
			},
		},
	}
	for _, r := range recipients {
		multisendMsg.Outputs = append(multisendMsg.Outputs, banktypes.Output{
			Address: r,
			Coins:   sdk.NewCoins(chain.NewCoin(sdk.NewIntFromUint64(amount))),
		})
	}

	txBuilder, err := chain.TxFactory().
		WithGas(chain.GasLimitByMsgs(multisendMsg) + 10000).
		BuildUnsignedTx(multisendMsg)
	requireT.NoError(err)

	json, err := chain.ClientContext.TxConfig().TxJSONEncoder()(txBuilder.GetTx())
	requireT.NoError(err)

	requireT.NoError(chain.ClientContext.PrintString(fmt.Sprintf("%s\n", json)))
}
