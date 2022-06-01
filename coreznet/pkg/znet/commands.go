package znet

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/ioc"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/zstress"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// Activate starts preconfigured bash environment
func Activate(ctx context.Context, configF *ConfigFactory) error {
	config := configF.Config()

	exe := must.String(filepath.EvalSymlinks(must.String(os.Executable())))

	must.OK(ioutil.WriteFile(config.WrapperDir+"/start", []byte(fmt.Sprintf("#!/bin/bash\nexec %s start \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/stop", []byte(fmt.Sprintf("#!/bin/bash\nexec %s stop \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/remove", []byte(fmt.Sprintf("#!/bin/bash\nexec %s remove \"$@\"", exe)), 0o700))
	// `test` can't be used here because it is a reserved keyword in bash
	must.OK(ioutil.WriteFile(config.WrapperDir+"/tests", []byte(fmt.Sprintf("#!/bin/bash\nexec %s test \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/spec", []byte(fmt.Sprintf("#!/bin/bash\nexec %s spec \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/ping-pong", []byte(fmt.Sprintf("#!/bin/bash\nexec %s ping-pong \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/stress", []byte(fmt.Sprintf("#!/bin/bash\nexec %s stress \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/logs", []byte(fmt.Sprintf(`#!/bin/bash
if [ "$1" == "" ]; then
  echo "Provide the name of application"
  exit 1
fi
exec tail -f -n +0 "%s/$1.log"
`, config.LogDir)), 0o700))

	bash := osexec.Command("bash")
	bash.Env = append(os.Environ(),
		fmt.Sprintf("PS1=%s", "("+configF.EnvName+`) [\u@\h \W]\$ `),
		fmt.Sprintf("PATH=%s", config.WrapperDir+":/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin"),
		fmt.Sprintf("COREZNET_ENV=%s", configF.EnvName),
		fmt.Sprintf("COREZNET_MODE=%s", configF.ModeName),
		fmt.Sprintf("COREZNET_HOME=%s", configF.HomeDir),
		fmt.Sprintf("COREZNET_TARGET=%s", configF.Target),
		fmt.Sprintf("COREZNET_BIN_DIR=%s", configF.BinDir),
		fmt.Sprintf("COREZNET_FILTERS=%s", strings.Join(configF.TestFilters, ",")),
	)
	bash.Dir = config.LogDir
	bash.Stdin = os.Stdin
	err := libexec.Exec(ctx, bash)
	if bash.ProcessState != nil && bash.ProcessState.ExitCode() != 0 {
		// bash returns non-exit code if command executed in the shell failed
		return nil
	}
	return err
}

// Start starts environment
func Start(ctx context.Context, target infra.Target, mode infra.Mode) (retErr error) {
	return target.Deploy(ctx, mode)
}

// Stop stops environment
func Stop(ctx context.Context, target infra.Target, spec *infra.Spec) (retErr error) {
	defer func() {
		spec.PGID = 0
		for _, app := range spec.Apps {
			if app.Status() == infra.AppStatusRunning {
				app.SetStatus(infra.AppStatusStopped)
			}
			if err := spec.Save(); retErr == nil {
				retErr = err
			}
		}
	}()
	return target.Stop(ctx)
}

// Remove removes environment
func Remove(ctx context.Context, config infra.Config, target infra.Target) (retErr error) {
	if err := target.Remove(ctx); err != nil {
		return err
	}

	// It may happen that some files are flushed to disk even after processes are terminated
	// so let's try to delete dir a few times
	var err error
	for i := 0; i < 3; i++ {
		if err = os.RemoveAll(config.HomeDir); err == nil || errors.Is(err, os.ErrNotExist) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return errors.WithStack(err)
}

// Test runs integration tests
func Test(c *ioc.Container, configF *ConfigFactory) error {
	configF.TestingMode = true
	configF.ModeName = "test"
	var err error
	c.Call(func(ctx context.Context, config infra.Config, target infra.Target, appF *apps.Factory, spec *infra.Spec) (retErr error) {
		defer func() {
			if err := spec.Save(); retErr == nil {
				retErr = err
			}
		}()

		env, tests := tests.Tests(appF)
		return testing.Run(ctx, target, env, tests, config.TestFilters)
	}, &err)
	return err
}

// Spec print specification of running environment
func Spec(spec *infra.Spec) error {
	fmt.Println(spec)
	return nil
}

// PingPong connects to cored node and sends transactions back and forth from one account to another to generate
// transactions on the blockchain
func PingPong(ctx context.Context, mode infra.Mode) error {
	coredNode, err := coredNode(mode)
	if err != nil {
		return err
	}
	client := coredNode.Client()

	alice := cored.Wallet{Name: "alice", Key: cored.AlicePrivKey}
	bob := cored.Wallet{Name: "bob", Key: cored.BobPrivKey}
	charlie := cored.Wallet{Name: "charlie", Key: cored.CharliePrivKey}

	for {
		if err := sendTokens(ctx, client, alice, bob); err != nil {
			return err
		}
		if err := sendTokens(ctx, client, bob, charlie); err != nil {
			return err
		}
		if err := sendTokens(ctx, client, charlie, alice); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// Stress runs benchmark implemented by `corezstress` on top of network deployed by `coreznet`
func Stress(ctx context.Context, mode infra.Mode) error {
	coredNode, err := coredNode(mode)
	if err != nil {
		return err
	}
	return zstress.Stress(ctx, zstress.StressConfig{
		ChainID:           coredNode.ChainID(),
		NodeAddress:       coredNode.RPCAddress(),
		Accounts:          cored.RandomWallets,
		NumOfAccounts:     len(cored.RandomWallets),
		NumOfTransactions: 100,
	})
}

func coredNode(mode infra.Mode) (apps.Cored, error) {
	for _, app := range mode {
		if app.Type() == apps.CoredType && app.Status() == infra.AppStatusRunning {
			return app.(apps.Cored), nil
		}
	}
	return apps.Cored{}, errors.New("haven't found any running cored node")
}

func sendTokens(ctx context.Context, client cored.Client, from, to cored.Wallet) error {
	log := logger.Get(ctx)

	amount := cored.Balance{Amount: big.NewInt(1), Denom: "core"}
	txBytes, err := client.TxBankSend(from, to, amount)
	if err != nil {
		return err
	}
	txHash, err := client.Broadcast(txBytes)
	if err != nil {
		return err
	}

	log.Info("Sent tokens", zap.Stringer("from", from), zap.Stringer("to", to),
		zap.Stringer("amount", amount), zap.String("txHash", txHash))

	fromBalance, err := client.QBankBalances(ctx, from)
	if err != nil {
		return err
	}
	toBalance, err := client.QBankBalances(ctx, to)
	if err != nil {
		return err
	}

	log.Info("Current balance", zap.Stringer("wallet", from), zap.Stringer("balance", fromBalance["core"]))
	log.Info("Current balance", zap.Stringer("wallet", to), zap.Stringer("balance", toBalance["core"]))

	return nil
}