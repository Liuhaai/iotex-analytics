// Copyright (c) 2019 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package xrc20

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/iotexproject/iotex-analytics/indexprotocol"
	"github.com/pkg/errors"
)

const (
	// transferSha3 is sha3 of xrc20's transfer event,keccak('Transfer(address,address,uint256)')
	transferSha3 = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	xrc20TransactionTableName = "xrc20_transactions"
	xrc20BalanceTableName     = "xrc20_balances"
)

var (
	createTransactionTable = "CREATE TABLE IF NOT EXISTS `" + xrc20TransactionTableName + "` (" +
		"`block_height` DECIMAL(65) UNSIGNED NULL," +
		"`action_hash` VARCHAR(40) NOT NULL," +
		"`idx` INT(5) UNSIGNED NOT NULL," +
		"`token` VARCHAR(41) NOT NULL," +
		"`sender` VARCHAR(41) NOT NULL," +
		"`recipient` VARCHAR(41) NOT NULL," +
		"`amount` DECIMAL(65) UNSIGNED NULL," +
		"PRIMARY KEY (`action_hash`, `idx`)," +
		"KEY `i_block_height` (`block_height`)," +
		"KEY `i_sender` (`sender`)," +
		"KEY `i_recipient` (`recipient`)," +
		"KEY `i_token` (`token`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;"

	createBalanceTable = "CREATE TABLE IF NOT EXISTS `" + xrc20BalanceTableName + "` (" +
		"`token` varchar(41) NOT NULL," +
		"`account` varchar(41) NOT NULL," +
		"`block_height` decimal(65,0) NOT NULL," +
		"`balance` decimal(65,0) unsigned NOT NULL," +
		"PRIMARY KEY (`token`,`account`,`block_height`)," +
		"KEY `i_block_height` (`block_height`)," +
		"KEY `i_account` (`account`)," +
		"KEY `i_token` (`token`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;"

	insertTransactionQuery = "INSERT IGNORE INTO `" + xrc20TransactionTableName + "` (block_height, action_hash, idx, token, sender, recipient, amount) SELECT ?,?,?,?,?,?,? FROM %s m WHERE m.token = ? AND m.account in ('*',?,?) LIMIT 1"
	/*
		updateBalances = "INSERT INTO `" + xrc20BalanceTableName + "` (token, address, block_height, income, expense) " +
			"SELECT t1.token, t1.address, %d, coalesce(SUM(t2.amount), 0), coalesce(SUM(t3.amount), 0) " +
			"FROM (SELECT `token`, `sender` address FROM `" + xrc20TransactionTableName + "` WHERE block_height = %d UNION SELECT `token`, `recipient` address FROM `" + xrc20TransactionTableName + "` WHERE block_height = %d) t1 " +
			"LEFT JOIN `" + xrc20TransactionTableName + "` t2 ON t1.address = t2.recipient AND t1.token = t2.token " +
			"LEFT JOIN `" + xrc20TransactionTableName + "` t3 ON t1.address = t3.sender AND t1.token = t3.token " +
			"GROUP BY t1.token, t1.address"
	*/

	updateBalancesQuery = "INSERT INTO `" + xrc20BalanceTableName + "` (`token`, `account`, `block_height`, `balance`) " +
		"SELECT delta.token, delta.account, ?, coalesce(curr.balance, 0) + coalesce(SUM(delta.balance), 0) " +
		"FROM (" +
		"    SELECT b1.token, b1.account, b1.balance " +
		"    FROM `" + xrc20BalanceTableName + "` b1 " +
		"    INNER JOIN ((" +
		"        SELECT t.token, t.account, MAX(t.block_height) max_height " +
		"        FROM `" + xrc20BalanceTableName + "` t " +
		"        INNER JOIN %s m ON t.token = m.token AND t.account = m.account " +
		"        GROUP BY t.token, t.account " +
		"    ) UNION (" +
		"        SELECT t.token, t.account, MAX(t.block_height) max_height " +
		"        FROM `" + xrc20BalanceTableName + "` t " +
		"        INNER JOIN %s m ON t.token = m.token WHERE m.account = '*' " +
		"        GROUP BY t.token, t.account " +
		"    )) h1 ON b1.token = h1.token AND b1.account = h1.account AND b1.block_height = h1.max_height" +
		") AS curr RIGHT JOIN ((" +
		"    SELECT t.token, t.sender account, SUM(-t.amount) balance " +
		"    FROM `" + xrc20TransactionTableName + "` t " +
		"    INNER JOIN %s m ON m.token = t.token AND m.account = t.sender AND t.block_height = ? " +
		"    GROUP BY t.token, t.sender " +
		") UNION (" +
		"    SELECT t.token, t.sender account, SUM(-t.amount) balance " +
		"    FROM `" + xrc20TransactionTableName + "` t " +
		"    INNER JOIN %s m ON m.token = t.token AND m.account = '*' AND t.block_height = ? " +
		"    GROUP BY t.token, t.sender " +
		") UNION (" +
		"    SELECT t.token, t.recipient account, SUM(t.amount) balance " +
		"    FROM `" + xrc20TransactionTableName + "` t " +
		"    INNER JOIN %s m ON m.token = t.token AND m.account = t.recipient AND t.block_height = ? " +
		"    GROUP BY t.token, t.recipient " +
		") UNION (" +
		"    SELECT t.token, t.recipient account, SUM(t.amount) balance " +
		"    FROM `" + xrc20TransactionTableName + "` t " +
		"    INNER JOIN %s m ON m.token = t.token AND m.account = '*' AND t.block_height = ? " +
		"    GROUP BY t.token, t.recipient " +
		")) AS `delta` ON delta.token = curr.token and delta.account = curr.account " +
		"GROUP BY delta.token, delta.account"

	// lookupLatestBalance = "SELECT balance FROM `" + xrc20BalanceTableName + "` WHERE token = ? AND account = ? ORDER BY block_height DESC LIMIT 1"
)

type (
	// Transaction defines an IOTX transaction
	Transaction struct {
		Height     uint64
		ActionHash string
		Index      uint16
		Xrc20      string
		Sender     string
		Recipient  string
		Amount     string
	}

	// Protocol defines an xrc 20 protocol
	Protocol struct {
		insertTransactionQuery string
		updateBalancesQuery    string
	}
)

// NewProtocolWithMonitorTokenAndAddresses creates a new xrc20 Protocol
func NewProtocolWithMonitorTokenAndAddresses(monitorTableName string) *Protocol {
	return &Protocol{
		insertTransactionQuery: fmt.Sprintf(insertTransactionQuery, monitorTableName),
		updateBalancesQuery:    fmt.Sprintf(updateBalancesQuery, monitorTableName, monitorTableName, monitorTableName, monitorTableName, monitorTableName, monitorTableName),
	}
}

// Initialize starts the xrc20 protocol
func (p *Protocol) Initialize(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.Exec(createTransactionTable); err != nil {
		return err
	}
	if _, err := tx.Exec(createBalanceTable); err != nil {
		return err
	}
	return nil
}

// HandleBlockData stores Xrc20 information into Xrc20 related tables
func (p *Protocol) HandleBlockData(
	ctx context.Context,
	tx *sql.Tx,
	data *indexprotocol.BlockData,
) error {
	count := 0
	for _, receipt := range data.Block.Receipts {
		if receipt.Status != uint64(1) {
			// TODO: handle the legacy status
			continue
		}
		for _, l := range receipt.Logs() {
			if len(l.Topics) != 3 {
				continue
			}
			if hex.EncodeToString(l.Topics[0][:]) != transferSha3 {
				continue
			}
			sender, err := indexprotocol.ConvertTopicToAddress(l.Topics[1])
			if err != nil {
				return err
			}
			recipient, err := indexprotocol.ConvertTopicToAddress(l.Topics[2])
			if err != nil {
				return err
			}
			amount := new(big.Int).SetBytes(l.Data)
			if amount.Cmp(big.NewInt(0)) == 0 {
				continue
			}
			if _, err := tx.Exec(
				p.insertTransactionQuery,
				l.BlockHeight,
				hex.EncodeToString(l.ActionHash[:]),
				l.Index,
				l.Address,
				sender.String(),
				recipient.String(),
				amount.String(),
				l.Address,          // token in WHERE clause
				sender.String(),    // sender in WHERE clause
				recipient.String(), // recipient in WHERE clause
			); err != nil {
				return errors.Wrap(err, "failed to index xrc20 log")
			}
			count++
		}
	}
	if count == 0 {
		return nil
	}
	height := data.Block.Height()
	if _, err := tx.Exec(p.updateBalancesQuery, height, height, height, height, height); err != nil {
		return errors.Wrap(err, "failed to update xrc20 balances")
	}

	return nil
}