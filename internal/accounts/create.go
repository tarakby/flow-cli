/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package accounts

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsCreate struct {
	Signer    string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Keys      []string `flag:"key" info:"Public keys to attach to account"`
	Weights   []int    `flag:"key-weight" info:"Weight for the key"`
	SigAlgo   []string `default:"ECDSA_P256" flag:"sig-algo" info:"Signature algorithm used to generate the keys"`
	HashAlgo  []string `default:"SHA3_256" flag:"hash-algo" info:"Hash used for the digest"`
	Contracts []string `flag:"contract" info:"Contract to be deployed during account creation. <name:filename>"`
	Include   []string `default:"" flag:"include" info:"Fields to include in the output"`
}

var createFlags = flagsCreate{}

var CreateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create",
		Short:   "Create a new account on network",
		Example: `flow accounts create --key d651f1931a2...8745`,
	},
	Flags: &createFlags,
	RunS:  create,
}

func create(
	_ []string,
	loader flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	// if user doesn't provide any flags go into interactive mode
	if len(createFlags.Keys) == 0 {
		account, err := createInteractive(state, loader)
		if err != nil {
			return nil, err
		}
		return &AccountResult{
			Account: account,
			include: createFlags.Include,
		}, nil
	}

	signer, err := state.Accounts().ByName(createFlags.Signer)
	if err != nil {
		return nil, err
	}

	if len(createFlags.SigAlgo) == 1 && len(createFlags.HashAlgo) == 1 {
		// Fill up depending on size of key input
		if len(createFlags.Keys) > 1 {
			for i := 1; i < len(createFlags.Keys); i++ {
				createFlags.SigAlgo = append(createFlags.SigAlgo, createFlags.SigAlgo[0])
				createFlags.HashAlgo = append(createFlags.HashAlgo, createFlags.HashAlgo[0])
			}
			// Deprecated usage message?
		}

	} else if len(createFlags.Keys) != len(createFlags.SigAlgo) || len(createFlags.SigAlgo) != len(createFlags.HashAlgo) { // double check matching array lengths on inputs
		return nil, fmt.Errorf("must provide a signature and hash algorithm for every key provided to --key: %d keys, %d signature algo, %d hash algo", len(createFlags.Keys), len(createFlags.SigAlgo), len(createFlags.HashAlgo))
	}

	// read all signature algorithms
	sigAlgos := make([]crypto.SignatureAlgorithm, 0, len(createFlags.SigAlgo))
	for _, sigAlgoStr := range createFlags.SigAlgo {
		sigAlgo := crypto.StringToSignatureAlgorithm(sigAlgoStr)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			return nil, fmt.Errorf("invalid signature algorithm: %s", createFlags.SigAlgo)
		}
		sigAlgos = append(sigAlgos, sigAlgo)
	}

	// read all hash algorithms
	hashAlgos := make([]crypto.HashAlgorithm, 0, len(createFlags.HashAlgo))
	for _, hashAlgoStr := range createFlags.HashAlgo {

		hashAlgo := crypto.StringToHashAlgorithm(hashAlgoStr)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			return nil, fmt.Errorf("invalid hash algorithm: %s", createFlags.HashAlgo)
		}
		hashAlgos = append(hashAlgos, hashAlgo)
	}

	keyWeights := createFlags.Weights

	// decode public keys
	pubKeys := make([]crypto.PublicKey, 0, len(createFlags.Keys))
	for i, k := range createFlags.Keys {
		k = strings.TrimPrefix(k, "0x") // clear possible prefix
		key, err := crypto.DecodePublicKeyHex(sigAlgos[i], k)
		if err != nil {
			return nil, fmt.Errorf("failed decoding public key: %s with error: %w", key, err)
		}
		pubKeys = append(pubKeys, key)
	}

	account, err := services.Accounts.Create(
		signer,
		pubKeys,
		keyWeights,
		sigAlgos,
		hashAlgos,
		createFlags.Contracts,
	)

	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account: account,
		include: createFlags.Include,
	}, nil
}

func createInteractive(state *flowkit.State, loader flowkit.ReaderWriter) (*flow.Account, error) {
	network := output.CreateAccountNetworkPrompt()
	log := output.NewStdoutLogger(output.InfoLog)

	// create new gateway based on chosen network
	gw, err := gateway.NewGrpcGateway(network.Host)
	if err != nil {
		return nil, err
	}
	service := services.NewServices(gw, state, output.NewStdoutLogger(output.NoneLog))
	log.Info("\nSTARTING ACCOUNT CREATION PROCESS: ")
	log.Info("\n------------------------")
	log.Info(fmt.Sprintf("%s (1/3) Generate public and private keys", output.GoEmoji()))
	log.Info(fmt.Sprintf("%s (2/3) Create account on %s with generated keys", output.GoEmoji(), network.Name))
	log.Info(fmt.Sprintf("%s (3/3) Save newly created account on flow.json", output.GoEmoji()))
	log.Info("\n------------------------")
	time.Sleep(time.Second * 2)

	log.Info(fmt.Sprintf("%s (1/3) GENERATING PUBLIC AND PRIVATE KEYS", output.WarningEmoji()))
	log.Info("\n------------------------")
	time.Sleep(time.Second * 3)
	key, err := service.Keys.Generate("", crypto.ECDSA_P256)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("%s (1/3) SUCCESSFULLY GENERATED PUBLIC AND PRIVATE KEYS", output.OkEmoji()))
	log.Info(fmt.Sprintf("\n PUBLIC KEY: %s", key.PublicKey().String()))
	log.Info(fmt.Sprintf("\n PRIVATE KEY: %s", key.String()))
	log.Info("\n------------------------")
	time.Sleep(time.Second * 2)

	log.Info(fmt.Sprintf("%s (2/3) CREATING ACCOUNT ON %s WITH GENERATED KEYS", output.WarningEmoji(),
		strings.ToUpper(network.Name)))
	log.Info("\n------------------------")
	time.Sleep(time.Second * 3)

	startHeight, err := service.Blocks.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	var address flow.Address

	if network == config.DefaultEmulatorNetwork() {
		signer, err := state.EmulatorServiceAccount()
		if err != nil {
			return nil, err
		}
		account, err := service.Accounts.Create(
			signer,
			[]crypto.PublicKey{key.PublicKey()},
			[]int{flow.AccountKeyWeightThreshold},
			[]crypto.SignatureAlgorithm{crypto.ECDSA_P256},
			[]crypto.HashAlgorithm{crypto.SHA3_256},
			nil,
		)
		if err != nil {
			return nil, err
		}

		log.Info("\nPlease note, that the newly created account will only be available until you keep the emulator service up and running, if you restart the emulator service all accounts will be reset. If you want to persist accounts between restarts you must use the '--persist' flag when starting the flow emulator.")

		address = account.Address
	} else {
		var link string
		switch network {
		case config.DefaultTestnetNetwork():
			log.Info("\nA testnet faucet website will open, follow the steps to create an account: \n 1. Fill in the captcha, \n 2. Click on 'Create Account' button.\n")
			link = util.TestnetFaucetURL(key.PublicKey().String(), crypto.ECDSA_P256)
		case config.DefaultMainnetNetwork():
			log.Info("\nA Flow Port website will open, follow the steps to create an account: \n 1. Click on 'Submit' button, \n 2. Connect existing Blocto or create a new account first, \n 3. Click on confirm, \n 4. Click on approve. \n")
			link = util.MainnetFlowPortURL(key.PublicKey().String())
		}

		time.Sleep(time.Second * 2)
		err := util.OpenBrowserWindow(link)
		if err != nil {
			return nil, err
		}

		log.StartProgress("Waiting for an account to be created, please finish all the steps in the browser...")

		addr, err := getAccountCreatedAddressWithPubKey(service, key.PublicKey(), startHeight)
		if err != nil {
			return nil, err
		}
		address = *addr

		log.StopProgress()
	}

	onChainAccount, err := service.Accounts.Get(address)
	if err != nil {
		return nil, err
	}

	name := output.NamePrompt()

	account, err := flowkit.NewAccountFromOnChainAccount(name, onChainAccount, key)
	if err != nil {
		return nil, err
	}
	log.Info("\n------------------------")
	log.Info(fmt.Sprintf("\n %s (2/3) SUCCESSFULLY CREATED ACCOUNT", output.OkEmoji()))
	log.Info(fmt.Sprintf("\n ACCOUNT NAME: %s", strings.ToUpper(name)))
	log.Info(fmt.Sprintf("\n ADDRESS: %s", fmt.Sprintf("0x%s", address.String())))
	log.Info("\n------------------------")
	time.Sleep(time.Second * 2)

	log.Info(fmt.Sprintf("\n %s (3/3) SAVING CREATED ACCOUNT IN FLOW.JSON", output.WarningEmoji()))
	time.Sleep(time.Second * 3)

	err = saveAccount(loader, state, account, network)
	if err != nil {
		return nil, err
	}
	log.Info("\n------------------------")
	log.Info(fmt.Sprintf("%s (3/3) SUCCESSFULLY SAVED ACCOUNT IN FLOW.JSON", output.OkEmoji()))
	if network != config.DefaultEmulatorNetwork() {
		fileName := strings.ToUpper(fmt.Sprintf("%s.private.json", name))
		log.Info(fmt.Sprintf("%s PRIVATE KEY FOR %s SUCCESSFULLY SAVED IN %s", output.OkEmoji(),
			strings.ToUpper(name), fileName))
		if output.AddToGitIgnorePrompt(fileName) {
			log.Info(fmt.Sprintf("%s %s ADDED TO .GITIGNORE", output.OkEmoji(), fileName))
		}
	}
	log.Info("\n------------------------")
	time.Sleep(time.Second * 2)

	log.Info(fmt.Sprintf("%s ACCOUNT CREATION PROCESS COMPLETED!", output.SuccessEmoji()))
	log.Info("\n------------------------")
	log.Info(fmt.Sprintf("\n %s (1/3) Generated public and private keys", output.OkEmoji()))
	log.Info(fmt.Sprintf("\n PUBLIC KEY: %s", key.PublicKey().String()))
	log.Info(fmt.Sprintf("\n PRIVATE KEY: %s", key.String()))
	log.Info(fmt.Sprintf("\n %s (2/3) Created account on %s with generated keys", output.OkEmoji(), network.Name))
	log.Info(fmt.Sprintf("\n ACCOUNT NAME: %s", strings.ToUpper(name)))
	log.Info(fmt.Sprintf("\n ADDRESS: %s", fmt.Sprintf("0x%s", address.String())))
	log.Info(fmt.Sprintf("\n %s (3/3) Saved newly created account on flow.json", output.OkEmoji()))
	log.Info("\n------------------------")
	time.Sleep(time.Second * 3)
	log.Info("\nACCOUNT CREATION SUMMARY")
	log.Info("\n------------------------")
	return onChainAccount, nil
}

func getAccountCreatedAddressWithPubKey(
	service *services.Services,
	pubKey crypto.PublicKey,
	startHeight uint64,
) (*flow.Address, error) {
	lastHeight, err := service.Blocks.GetLatestBlockHeight()
	if err != nil {
		return nil, err
	}

	flowEvents, err := service.Events.Get([]string{flow.EventAccountKeyAdded}, startHeight, lastHeight, 20, 1)
	if err != nil {
		return nil, err
	}

	var address *flow.Address
	for _, block := range flowEvents {
		events := flowkit.NewEvents(block.Events)
		address = events.GetAddressForKeyAdded(pubKey)
		if address != nil {
			break
		}
	}

	if address == nil {
		//TODO:sideninja 200 blocks might not be enough time for the user to sign into their wallet and create the account on mainnet
		if lastHeight-startHeight > 200 { // if something goes wrong don't keep waiting forever to avoid spamming network
			return nil, fmt.Errorf("failed to get the account address due to time out")
		}

		time.Sleep(time.Second * 2)
		address, err = getAccountCreatedAddressWithPubKey(service, pubKey, startHeight)
		if err != nil {
			return nil, err
		}

		return address, nil
	}

	return address, nil
}

func saveAccount(
	loader flowkit.ReaderWriter,
	state *flowkit.State,
	account *flowkit.Account,
	network config.Network,
) error {
	// If using emulator, save account private key to main flow.json configuration file
	if network == config.DefaultEmulatorNetwork() {
		return saveAccountToMainConfigFile(state, account)
	}

	// Otherwise, save to a separate {accountName}.private.json file.
	return saveAccountToPrivateConfigFile(loader, state, account)
}

func saveAccountToPrivateConfigFile(
	loader flowkit.ReaderWriter,
	state *flowkit.State,
	account *flowkit.Account,
) error {
	privateAccountFilename := fmt.Sprintf("%s.private.json", account.Name())

	// Step 1: save the private version of the account (incl. the private key)
	// to a separate JSON file.
	err := savePrivateAccount(loader, privateAccountFilename, account)
	if err != nil {
		return err
	}

	// Step 2: update the main configuration file to inlcude a reference
	// to the private account file.
	fromFileAccount := flowkit.
		NewAccount(account.Name()).
		SetFromFile(privateAccountFilename)

	state.Accounts().AddOrUpdate(fromFileAccount)

	err = state.SaveDefault()
	if err != nil {
		return err
	}

	return nil
}

func saveAccountToMainConfigFile(
	state *flowkit.State,
	account *flowkit.Account,
) error {
	state.Accounts().AddOrUpdate(account)
	err := state.SaveDefault()
	if err != nil {
		return err
	}

	return nil
}

func savePrivateAccount(
	loader flowkit.ReaderWriter,
	fileName string,
	account *flowkit.Account,
) error {
	privateState := flowkit.NewEmptyState(loader)
	privateState.Accounts().AddOrUpdate(account)

	err := privateState.Save(fileName)
	if err != nil {
		return err
	}
	if output.AddToGitIgnorePrompt(fileName) {
		addToGitIgnore(fileName)
	}

	return nil
}

func addToGitIgnore(
	filename string,
) error {
	currentWd, err := os.Getwd()
	if err != nil {
		return err
	}
	gitIgnoreDir := path.Join(currentWd, ".gitignore")

	//opens .gitignore so that we can add {name}.private.json to .gitignore, will create .gitignore if it doesn't exist
	file, err := os.OpenFile(gitIgnoreDir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString("\n" + filename)
	if err != nil {
		return err
	}
	return nil
}
