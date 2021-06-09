/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-cli/pkg/flowkit/config"

	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsUpdateContract struct {
	Signer  string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Results bool     `default:"false" flag:"results" info:"⚠️  Deprecated: results are provided by default"`
	Include []string `default:"" flag:"include" info:"Fields to include in the output"`
}

var updateFlags = flagsUpdateContract{}

var UpdateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "update-contract <name> <filename>",
		Short:   "Update a contract deployed to an account",
		Example: `flow accounts update-contract FungibleToken ./FungibleToken.cdc`,
		Args:    cobra.ExactArgs(2),
	},
	Flags: &updateFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
		proj *flowkit.Project,
	) (command.Result, error) {
		if updateFlags.Results {
			fmt.Println("⚠️ DEPRECATION WARNING: results flag is deprecated, results are by default included in all executions")
		}

		name := args[0]
		filename := args[1]

		code, err := util.LoadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error loading contract file: %w", err)
		}

		if proj == nil {
			return nil, config.ErrDoesNotExist
		}
		to := proj.AccountByName(addContractFlags.Signer)

		account, err := services.Accounts.AddContract(to, name, code, false)
		if err != nil {
			return nil, err
		}

		return &AccountResult{
			Account:  account,
			showCode: false,
			include:  updateFlags.Include,
		}, nil
	},
}
