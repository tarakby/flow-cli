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

package services

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/tests"
)

func TestEvents(t *testing.T) {
	mock := tests.DefaultMockGateway()
	readerWriter := tests.ReaderWriter()
	state, err := flowkit.Init(readerWriter, crypto.ECDSA_P256, crypto.SHA3_256)
	assert.NoError(t, err)

	events := NewEvents(mock, state, output.NewStdoutLogger(output.InfoLog))

	t.Run("Get Events", func(t *testing.T) {
		_, err := events.Get("flow.CreateAccount", "0", "1")

		assert.NoError(t, err)
		mock.AssertFuncsCalled(t, true, mock.GetEvents)
	})

	t.Run("Get Events Latest", func(t *testing.T) {
		_, err := events.Get("flow.CreateAccount", "0", "latest")

		assert.NoError(t, err)
		mock.AssertFuncsCalled(t, true, mock.GetEvents, mock.GetLatestBlock)
	})

	t.Run("Fails to get events without name", func(t *testing.T) {
		_, err := events.Get("", "0", "1")
		assert.Equal(t, err.Error(), "cannot use empty string as event name")
	})

	t.Run("Fails to get events with wrong height", func(t *testing.T) {
		_, err := events.Get("test", "-1", "1")
		assert.Equal(t, err.Error(), "failed to parse start height of block range: -1")
	})

	t.Run("Fails to get events with wrong end height", func(t *testing.T) {
		_, err := events.Get("test", "1", "-1")
		assert.Equal(t, err.Error(), "failed to parse end height of block range: -1")
	})

	t.Run("Fails to get events with wrong start height", func(t *testing.T) {
		_, err := events.Get("test", "10", "5")
		assert.Equal(t, err.Error(), "cannot have end height (5) of block range less that start height (10)")
	})
}
