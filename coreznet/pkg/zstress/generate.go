package zstress

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"

	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
)

// GenerateConfig contains config for generating the blockchain
type GenerateConfig struct {
	// ChainID is the ID of the chain to generate
	ChainID string

	// NumOfValidators is the number of validators present on the blockchain
	NumOfValidators int

	// NumOfInstances is the maximum number of application instances used in the future during benchmarking
	NumOfInstances int

	// NumOfAccountsPerInstance is the maximum number of funded accounts per each instance used in the future during benchmarking
	NumOfAccountsPerInstance int

	// OutDirectory is the path to the directory where generated files are stored
	OutDirectory string
}

// Generate generates all the files required to deploy blockchain used for benchmarking
func Generate(config GenerateConfig) error {
	coredPath, err := exec.LookPath("cored")
	if err != nil {
		return fmt.Errorf(`can't find cored binary, run "core build/cored" to build it: %w`, err)
	}
	coredPath = filepath.Dir(coredPath) + "/linux/cored"

	dir := config.OutDirectory + "/corezstress-deployment"
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	dockerDir := dir + "/docker"
	dockerDirBin := dockerDir + "/bin"
	must.OK(os.MkdirAll(dockerDirBin, 0o700))
	if err := os.Link(coredPath, dockerDirBin+"/cored"); err != nil {
		return fmt.Errorf(`can't find cored linux binary, run "core build/cored" to build it: %w`, err)
	}

	must.OK(ioutil.WriteFile(dockerDir+"/Dockerfile", []byte(`FROM scratch
COPY . .
ENTRYPOINT ["cored"]
`), 0o600))

	genesis := cored.NewGenesis(config.ChainID)
	nodeIDs := make([]string, 0, config.NumOfValidators)
	for i := 0; i < config.NumOfValidators; i++ {
		nodePublicKey, nodePrivateKey, err := ed25519.GenerateKey(rand.Reader)
		must.OK(err)
		nodeIDs = append(nodeIDs, cored.NodeID(nodePublicKey))
		validatorPublicKey, validatorPrivateKey, err := ed25519.GenerateKey(rand.Reader)
		must.OK(err)
		stakerPublicKey, stakerPrivateKey := cored.GenerateSecp256k1Key()

		valDir := fmt.Sprintf("%s/validators/%d", dir, i)

		cored.ValidatorConfig{
			Name:           fmt.Sprintf("validator-%d", i),
			IP:             net.IPv4zero,
			PrometheusPort: cored.DefaultPorts.Prometheus,
			NodeKey:        nodePrivateKey,
			ValidatorKey:   validatorPrivateKey,
		}.Save(valDir)

		genesis.AddWallet(stakerPublicKey, "100000000000000000000000core,10000000000000000000000000stake")
		genesis.AddValidator(validatorPublicKey, stakerPrivateKey, "100000000stake")
	}
	must.OK(ioutil.WriteFile(dir+"/validators/ids.json", must.Bytes(json.Marshal(nodeIDs)), 0o600))

	for i := 0; i < config.NumOfInstances; i++ {
		accounts := make([]cored.Secp256k1PrivateKey, 0, config.NumOfAccountsPerInstance)
		for j := 0; j < config.NumOfAccountsPerInstance; j++ {
			accountPublicKey, accountPrivateKey := cored.GenerateSecp256k1Key()
			accounts = append(accounts, accountPrivateKey)
			genesis.AddWallet(accountPublicKey, "10000000000000000000000000000core")
		}

		instanceDir := fmt.Sprintf("%s/instances/%d", dir, i)
		must.OK(os.MkdirAll(instanceDir, 0o700))
		must.OK(ioutil.WriteFile(instanceDir+"/accounts.json", must.Bytes(json.Marshal(accounts)), 0o600))
	}
	genesis.Save(dockerDir)
	return nil
}
