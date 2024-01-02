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

package dependencymanager

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
)

type addFlagsCollection struct {
	name string `default:"" flag:"name" info:"Name of the dependency"`
}

var addFlags = addFlagsCollection{}

var addCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "add",
		Short: "Add a single contract and its dependencies.",
		Args:  cobra.ExactArgs(1),
	},
	Flags: &addFlags,
	RunS:  add,
}

func add(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	logger.StartProgress(fmt.Sprintf("Installing dependencies for %s...", args[0]))
	defer logger.StopProgress()

	dep := args[0]

	installer := NewDepdencyInstaller(logger, state)
	if err := installer.add(dep, addFlags.name); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	logger.Info("✅  Dependencies installed. Check your flow.json")

	return nil, nil
}
