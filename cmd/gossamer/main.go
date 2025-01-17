// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"

	"github.com/ChainSafe/gossamer/dot"
	"github.com/ChainSafe/gossamer/lib/keystore"
	"github.com/ChainSafe/gossamer/lib/utils"

	log "github.com/ChainSafe/log15"
	"github.com/urfave/cli"
)

// app is the cli application
var app = cli.NewApp()
var logger = log.New("pkg", "cmd")

var (
	// exportCommand defines the "export" subcommand (ie, `gossamer export`)
	exportCommand = cli.Command{
		Action:    FixFlagOrder(exportAction),
		Name:      "export",
		Usage:     "Export configuration values to TOML configuration file",
		ArgsUsage: "",
		Flags:     ExportFlags,
		Category:  "EXPORT",
		Description: "The export command exports configuration values from the command flags to a TOML configuration file.\n" +
			"\tUsage: gossamer export --config chain/test/config.toml --basepath ~/.gossamer/test",
	}
	// initCommand defines the "init" subcommand (ie, `gossamer init`)
	initCommand = cli.Command{
		Action:    FixFlagOrder(initAction),
		Name:      "init",
		Usage:     "Initialize node databases and load genesis data to state",
		ArgsUsage: "",
		Flags:     InitFlags,
		Category:  "INIT",
		Description: "The init command initializes the node databases and loads the genesis data from the genesis configuration file to state.\n" +
			"\tUsage: gossamer init --genesis genesis.json",
	}
	// accountCommand defines the "account" subcommand (ie, `gossamer account`)
	accountCommand = cli.Command{
		Action:   FixFlagOrder(accountAction),
		Name:     "account",
		Usage:    "Create and manage node keystore accounts",
		Flags:    AccountFlags,
		Category: "ACCOUNT",
		Description: "The account command is used to manage the gossamer keystore.\n" +
			"\tTo generate a new sr25519 account: gossamer account --generate\n" +
			"\tTo generate a new ed25519 account: gossamer account --generate --ed25519\n" +
			"\tTo generate a new secp256k1 account: gossamer account --generate --secp256k1\n" +
			"\tTo import a keystore file: gossamer account --import=path/to/file\n" +
			"\tTo list keys: gossamer account --list",
	}
)

// init initializes the cli application
func init() {
	app.Action = gossamerAction
	app.Copyright = "Copyright 2019 ChainSafe Systems Authors"
	app.Name = "gossamer"
	app.Usage = "Official gossamer command-line interface"
	app.Author = "ChainSafe Systems 2019"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		exportCommand,
		initCommand,
		accountCommand,
	}
	app.Flags = RootFlags
}

// main runs the cli application
func main() {
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// gossamerAction is the root action for the gossamer command, creates a node
// configuration, loads the keystore, initializes the node if not initialized,
// then creates and starts the node and node services
func gossamerAction(ctx *cli.Context) error {
	// check for unknown command arguments
	if arguments := ctx.Args(); len(arguments) > 0 {
		return fmt.Errorf("failed to read command argument: %q", arguments[0])
	}

	// setup gossamer logger
	lvl, err := setupLogger(ctx)
	if err != nil {
		logger.Error("failed to setup logger", "error", err)
		return err
	}

	// create new dot configuration (the dot configuration is created within the
	// cli application from the flag values provided)
	cfg, err := createDotConfig(ctx)
	if err != nil {
		logger.Error("failed to create node configuration", "error", err)
		return err
	}

	cfg.Global.LogLevel = lvl.String()

	// expand data directory and update node configuration (performed separately
	// from createDotConfig because dot config should not include expanded path)
	cfg.Global.BasePath = utils.ExpandDir(cfg.Global.BasePath)

	// check if node has not been initialized (expected true - add warning log)
	if !dot.NodeInitialized(cfg.Global.BasePath, true) {

		// initialize node (initialize state database and load genesis data)
		err = dot.InitNode(cfg)
		if err != nil {
			logger.Error("failed to initialize node", "error", err)
			return err
		}
	}

	// ensure configuration matches genesis data stored during node initialization
	// but do not overwrite configuration if the corresponding flag value is set
	err = updateDotConfigFromGenesisData(ctx, cfg)
	if err != nil {
		logger.Error("failed to update config from genesis data", "error", err)
		return err
	}

	ks, err := keystore.LoadKeystore(cfg.Account.Key)
	if err != nil {
		logger.Error("failed to load keystore", "error", err)
		return err
	}

	err = unlockKeystore(ks, cfg.Global.BasePath, cfg.Account.Unlock, ctx.String(PasswordFlag.Name))
	if err != nil {
		logger.Error("failed to unlock keystore", "error", err)
		return err
	}

	node, err := dot.NewNode(cfg, ks)
	if err != nil {
		logger.Error("failed to create node services", "error", err)
		return err
	}

	logger.Info("starting node...", "name", node.Name)

	// start node
	err = node.Start()
	if err != nil {
		return err
	}

	return nil
}

// initAction is the action for the "init" subcommand, initializes the trie and
// state databases and loads initial state from the configured genesis file
func initAction(ctx *cli.Context) error {
	lvl, err := setupLogger(ctx)
	if err != nil {
		logger.Error("failed to setup logger", "error", err)
		return err
	}

	cfg, err := createInitConfig(ctx)
	if err != nil {
		logger.Error("failed to create node configuration", "error", err)
		return err
	}

	cfg.Global.LogLevel = lvl.String()

	// expand data directory and update node configuration (performed separately
	// from createDotConfig because dot config should not include expanded path)
	cfg.Global.BasePath = utils.ExpandDir(cfg.Global.BasePath)

	// check if node has been initialized (expected false - no warning log)
	if dot.NodeInitialized(cfg.Global.BasePath, false) {

		// use --force value to force initialize the node
		force := ctx.Bool(ForceFlag.Name)

		// prompt user to confirm reinitialization
		if force || confirmMessage("Are you sure you want to reinitialize the node? [Y/n]") {
			logger.Info(
				"reinitializing node...",
				"basepath", cfg.Global.BasePath,
			)
		} else {
			logger.Warn(
				"exiting without reinitializing the node",
				"basepath", cfg.Global.BasePath,
			)
			return nil // exit if reinitialization is not confirmed
		}
	}

	// initialize node (initialize state database and load genesis data)
	err = dot.InitNode(cfg)
	if err != nil {
		logger.Error("failed to initialize node", "error", err)
		return err
	}

	return nil
}
