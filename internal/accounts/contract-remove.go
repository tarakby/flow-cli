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
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

type flagsRemoveContract struct {
	Signer  string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var flagsRemove = flagsRemoveContract{}

var removeCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "remove-contract <name>",
		Short:   "Remove a contract deployed to an account",
		Example: `flow accounts remove-contract FungibleToken`,
		Args:    cobra.ExactArgs(1),
	},
	Flags: &flagsRemove,
	RunS:  removeContract,
}

func removeContract(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	contractName := args[0]

	from, err := state.Accounts().ByName(flagsRemove.Signer)
	if err != nil {
		return nil, err
	}

	_, err = flow.RemoveContract(context.Background(), from, contractName)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf(
		"Contract %s removed from account %s.",
		contractName,
		from.Address,
	))

	account, err := flow.GetAccount(context.Background(), from.Address)
	if err != nil {
		return nil, err
	}
	return &accountResult{
		Account: account,
		include: flagsRemove.Include,
	}, nil
}
