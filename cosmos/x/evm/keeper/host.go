// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2023, Furychain Foundation. All rights reserved.
// Use of this software is govered by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package keeper

import (
	storetypes "cosmossdk.io/store/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"

	"github.com/gridironOne/polaris/cosmos/x/evm/plugins"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/block"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/configuration"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/gas"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/historical"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/precompile"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/precompile/log"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/state"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/txpool"
	"github.com/gridironOne/polaris/cosmos/x/evm/plugins/txpool/mempool"
	"github.com/gridironOne/polaris/eth/core"
	ethprecompile "github.com/gridironOne/polaris/eth/core/precompile"
	"github.com/gridironOne/polaris/lib/utils"
)

// Compile-time interface assertion.
var _ core.PolarisHostChain = (*host)(nil)

// Host is the interface that must be implemented by the host.
// It includes core.PolarisHostChain and functions that are called in other packages.
type Host interface {
	core.PolarisHostChain
	GetAllPlugins() []plugins.Base
	Setup(
		storetypes.StoreKey,
		storetypes.StoreKey,
		state.AccountKeeper,
		state.BankKeeper,
		func(height int64, prove bool) (sdk.Context, error),
	)
}

type host struct {
	// The various plugins that are are used to implement core.PolarisHostChain.
	bp  block.Plugin
	cp  configuration.Plugin
	gp  gas.Plugin
	hp  historical.Plugin
	pp  precompile.Plugin
	sp  state.Plugin
	txp txpool.Plugin

	pcs func() *ethprecompile.Injector
}

// Newhost creates new instances of the plugin host.
func NewHost(
	storeKey storetypes.StoreKey,
	ak state.AccountKeeper,
	bk state.BankKeeper,
	authority string,
	appOpts servertypes.AppOptions,
	ethTxMempool sdkmempool.Mempool,
	precompiles func() *ethprecompile.Injector,
) Host {
	// We setup the host with some Cosmos standard sauce.
	h := &host{}

	// Build the Plugins
	h.bp = block.NewPlugin(storeKey)
	h.cp = configuration.NewPlugin(storeKey)
	h.gp = gas.NewPlugin()
	h.txp = txpool.NewPlugin(h.cp, utils.MustGetAs[*mempool.EthTxPool](ethTxMempool))
	h.pcs = precompiles

	return h
}

// Setup sets up the precompile and state plugins with the given precompiles and keepers. It also
// sets the query context function for the block and state plugins (to support historical queries).
func (h *host) Setup(
	storeKey storetypes.StoreKey,
	offchainStoreKey storetypes.StoreKey,
	ak state.AccountKeeper,
	bk state.BankKeeper,
	qc func(height int64, prove bool) (sdk.Context, error),
) {
	// Setup the state, precompile, historical, and txpool plugins
	h.sp = state.NewPlugin(ak, bk, storeKey, h.cp, log.NewFactory(h.pcs().GetPrecompiles()))
	h.pp = precompile.NewPlugin(h.pcs().GetPrecompiles(), h.sp)
	h.hp = historical.NewPlugin(h.bp, offchainStoreKey, storeKey)
	h.txp.SetNonceRetriever(h.sp)

	// Set the query context function for the block and state plugins
	h.sp.SetQueryContextFn(qc)
	h.bp.SetQueryContextFn(qc)
}

// GetBlockPlugin returns the header plugin.
func (h *host) GetBlockPlugin() core.BlockPlugin {
	return h.bp
}

// GetConfigurationPlugin returns the configuration plugin.
func (h *host) GetConfigurationPlugin() core.ConfigurationPlugin {
	return h.cp
}

// GetGasPlugin returns the gas plugin.
func (h *host) GetGasPlugin() core.GasPlugin {
	return h.gp
}

func (h *host) GetHistoricalPlugin() core.HistoricalPlugin {
	return h.hp
}

// GetPrecompilePlugin returns the precompile plugin.
func (h *host) GetPrecompilePlugin() core.PrecompilePlugin {
	return h.pp
}

// GetStatePlugin returns the state plugin.
func (h *host) GetStatePlugin() core.StatePlugin {
	return h.sp
}

// GetTxPoolPlugin returns the txpool plugin.
func (h *host) GetTxPoolPlugin() core.TxPoolPlugin {
	return h.txp
}

// GetAllPlugins returns all the plugins.
func (h *host) GetAllPlugins() []plugins.Base {
	return []plugins.Base{h.bp, h.cp, h.gp, h.hp, h.pp, h.sp, h.txp}
}
