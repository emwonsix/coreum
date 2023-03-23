package modules

// crust build/integration-tests
// ./coreum-modules -chain-id=coreum-mainnet-1 -cored-address=full-node-curium.mainnet-1.coreum.dev:9090 -test.run=TestValidatorDelegations > multisig.json

import (
	"fmt"
	"sort"
	"testing"

	keyring2 "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
)

func TestValidatorDelegations(t *testing.T) {
	delegations := map[string]uint64{
		// HashQuark
		"corevaloper1up3dsjekups4y2rf376a24l2ljtkc8t930hsft": 3000000_000_000,
		// Cosmostation
		"corevaloper1ll9gdh5ur6gyv6swgshcm4zkkw4ttakt0zghmr": 10000000_000_000,
		// B-Harvest
		"corevaloper1vsdgrlmq87qax45k32ghp94rcl3j03fwqe5lv4": 3000000_000_000,
		// ECO Stake
		"corevaloper1py9v5f7e4aaka55lwtthk30scxhnqfa6agwxt8": 3000000_000_000,
		// Stakewolle
		"corevaloper1fq63eks78npnsgtja8m8nvvv5qmahqxawgpafz": 5000000_000_000,
		// Citadel.one
		"corevaloper1pe5d6fkmghjrpgk7qps3nqmsnwth7ejuz7gezv": 4000000_000_000,
		// Smart Stake
		"corevaloper1zv3rz855lan7mznzlmzycf0a3tslar4lntsxk2": 2000000_000_000,
		// Stakecito
		"corevaloper19kg8c9k6quujkh0sflp78n4jsp3vzdy87540mm": 4000000_000_000,
		// StakingCabin
		"corevaloper16h0h2cjul7qa3np664ae9gqzrd0apcp8rqsdfs": 3000000_000_000,
		// Virtual Hive
		"corevaloper1e80r50sasmashmtg0s2dapdafy6c22n66ylaje": 3000000_000_000,
		// StakeLab
		"corevaloper1k0rllvenwr02gvm52fh5056g5m3hly2lpf63z5": 4000000_000_000,
		// Silk Nodes
		"corevaloper1kepnaw38rymdvq5sstnnytdqqkpd0xxwc5eqjk": 3000000_000_000,
		// Keyrock
		"corevaloper19p9mc0lrlndcwejrqk0m8a4jv035prnn0q34g0": 4000000_000_000,
		// EMERGE
		"corevaloper1nm7cx2z2zkfsqdn9vc9xvunczuts4gmd0n07xa": 4000000_000_000,
		// Bitrue
		"corevaloper10zh7zme4m3ash43k5cxz9ztq0refcpvepv3dar": 5000000_000_000,
		// ZenLounge
		"corevaloper1uhrrdv6g6v9t38v4qghjucunnxyk8xt30vr8za": 2000000_000_000,
		// 01node
		"corevaloper1flaz3hzgg3tjszl372lu2zz5jsmxd8pvydl7gg": 2000000_000_000,
		// Forbole
		"corevaloper1k3wy8ztt2e0uq3j5deukjxu2um4a4z5tvz35la": 3000000_000_000,
		// Informal Systems
		"corevaloper1x20lytyf6zkcrv5edpkfkn8sz578qg5sudvagj": 8000000_000_000,
		// Ubik Capital
		"corevaloper1q6mdzggk7feskx3uy90su8sqernswjpce8xjnp": 3000000_000_000,
	}

	requireT := require.New(t)

	ctx, chain := integrationtests.NewTestingContext(t)
	codec := chain.ClientContext.Codec()
	keyring := chain.ClientContext.Keyring()

	pubKey11 := &secp256k1.PubKey{}
	pubKey12 := &secp256k1.PubKey{}
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"At/R9OcXOK5I12Ps7gdjCZtzJ6Y5fIHYNS0X49YWVroQ"}`), pubKey11))
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"A05FyUd9eTuL0jeijQ49Lrpg7HkWepczxXcgEGvfSviw"}`), pubKey12))

	multisigKey1 := multisig.NewLegacyAminoPubKey(
		2,
		[]types.PubKey{pubKey11, pubKey12},
	)

	pubKey21 := &secp256k1.PubKey{}
	pubKey22 := &secp256k1.PubKey{}
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"A2QRVQk1slBv7OfpzWsvsmXv56L8SvniFb1mTjKiI7m9"}`), pubKey21))
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"AzXo4et8Q6iATJ0IV1wbTcVputdcYpv3HY8AYI6X3G0r"}`), pubKey22))

	multisigKey2 := multisig.NewLegacyAminoPubKey(
		2,
		[]types.PubKey{pubKey21, pubKey22},
	)

	pubKey31 := &secp256k1.PubKey{}
	pubKey32 := &secp256k1.PubKey{}
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"A/RAaZADO8/nkSKuSuQL2QAhIWWTPwG8SipA3j+7tHAd"}`), pubKey31))
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"AlxrRC60JP5k4IRjSna38QN8CL/XcAJOI7o6dqWb3n4Z"}`), pubKey32))

	multisigKey3 := multisig.NewLegacyAminoPubKey(
		2,
		[]types.PubKey{pubKey31, pubKey32},
	)

	pubKey41 := &secp256k1.PubKey{}
	pubKey42 := &secp256k1.PubKey{}
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"A9OV1Mske3aG9vhfsafGxV/GwEDJjYX0Yg/XJfIoyVg2"}`), pubKey41))
	requireT.NoError(codec.UnmarshalJSON([]byte(`{"key":"AyIdBfP0/vCdvrYtm13Dy6YfXB/NJn0yXn28Y6LqsLcF"}`), pubKey42))

	multisigKey4 := multisig.NewLegacyAminoPubKey(
		2,
		[]types.PubKey{pubKey41, pubKey42},
	)

	multisigInfo1, err := keyring.SaveMultisig("coreum-foundation-1", multisigKey1)
	requireT.NoError(err)
	multisigInfo2, err := keyring.SaveMultisig("coreum-foundation-2", multisigKey2)
	requireT.NoError(err)
	multisigInfo3, err := keyring.SaveMultisig("coreum-foundation-3", multisigKey3)
	requireT.NoError(err)
	multisigInfo4, err := keyring.SaveMultisig("coreum-foundation-4", multisigKey4)
	requireT.NoError(err)

	infos := []keyring2.Info{multisigInfo1, multisigInfo2, multisigInfo3, multisigInfo4}
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	balances := map[string]uint64{}
	for _, info := range infos {
		resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: info.GetAddress().String(), Denom: chain.NetworkConfig.Denom})
		requireT.NoError(err)

		balances[info.GetAddress().String()] = resp.Balance.Amount.Uint64()
	}

	validators := make([]string, 0, len(delegations))
	for v := range delegations {
		validators = append(validators, v)
	}
	sort.Strings(validators)

	msgs := map[string][]sdk.Msg{}
	for _, v := range validators {
		var found bool
		for _, info := range infos {
			if balances[info.GetAddress().String()] >= delegations[v]+2_000_000 {
				balances[info.GetAddress().String()] -= delegations[v]
				found = true

				msgs[info.GetAddress().String()] = append(msgs[info.GetAddress().String()], &stakingtypes.MsgDelegate{
					DelegatorAddress: info.GetAddress().String(),
					ValidatorAddress: v,
					Amount:           chain.NewCoin(sdk.NewIntFromUint64(delegations[v])),
				})
				break
			}
		}
		if !found {
			requireT.Fail("No funds left")
		}
	}

	for _, info := range infos {
		if msgsv := msgs[info.GetAddress().String()]; len(msgsv) > 0 {
			txBuilder, err := chain.TxFactory().
				WithGas(chain.GasLimitByManyMsgs(msgsv...) + 30_000).
				BuildUnsignedTx(msgsv...)
			requireT.NoError(err)

			json, err := chain.ClientContext.TxConfig().TxJSONEncoder()(txBuilder.GetTx())
			requireT.NoError(err)

			requireT.NoError(chain.ClientContext.PrintString(fmt.Sprintf("%s\n", json)))
			fmt.Println()
		}
	}
}

// crust build/integration-tests
// ./coreum-modules -chain-id=coreum-mainnet-1 -test.run=TestValidatorGrant > multisig.json

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
