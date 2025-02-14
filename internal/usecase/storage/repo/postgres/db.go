package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	//"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"merch/internal/entity"
	"merch/internal/usecase/storage"
)

var (
	ErrNotEnoughCoins = errors.New("not enough coins")
	ErrUserNotFound   = errors.New("user not found")
)

const BucketCount = 4

type (
	ShardMap map[int]*sql.DB
)
type PgRepo struct {
	ShardMap ShardMap
}

// TODO: add op
func (p *PgRepo) Transfer(ctx context.Context, fromUserID, toUserID, amount int) error {
	const op = "PgRepo - Transfer"

	shardFrom := p.getShardID(fromUserID)
	shardTo := p.getShardID(toUserID)

	fromDB := p.ShardMap[shardFrom]
	toDB := p.ShardMap[shardTo]

	if fromDB == toDB {
		tx, err := fromDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return fmt.Errorf("%s - failed to start transaction: %w", op, err)
		}

		err = p.transferInSameShard(ctx, tx, fromUserID, toUserID, amount)
		if err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit()
	}

	txFrom, err := fromDB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("%s - failed to BeginTx from_user: %w", op, err)
	}

	txTo, err := toDB.BeginTx(ctx, nil)
	if err != nil {
		txFrom.Rollback()
		return fmt.Errorf("%s - failed to BeginTx to_user: %w", op, err)
	}

	res, err := txFrom.ExecContext(ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2 AND coins >= $1", amount, fromUserID)
	if err != nil {
		txFrom.Rollback()
		txTo.Rollback()
		return fmt.Errorf("%s - failed to update coins from_user: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		txFrom.Rollback()
		txTo.Rollback()
		return fmt.Errorf("%s - failed to get rows affected: %w", op, err)
	}
	if rowsAffected == 0 {
		txFrom.Rollback()
		txTo.Rollback()
		return ErrNotEnoughCoins
	}

	_, err = txTo.ExecContext(ctx, "UPDATE users SET coins = coins + $1 WHERE id = $2", amount, toUserID)
	if err != nil {
		txFrom.Rollback()
		txTo.Rollback()
		return fmt.Errorf("%s - failed to update coins to_user: %w", op, err)
	}

	_, err = txFrom.ExecContext(ctx, "INSERT INTO coin_history (from_user, to_user, amount) VALUES ($1, $2, $3)", fromUserID, toUserID, amount)
	if err != nil {
		txFrom.Rollback()
		txTo.Rollback()
		return fmt.Errorf("%s - failed to insert in coin_history from_user: %w", op, err)
	}

	_, err = txTo.ExecContext(ctx, "INSERT INTO coin_history (from_user, to_user, amount) VALUES ($1, $2, $3)", fromUserID, toUserID, amount)
	if err != nil {
		txFrom.Rollback()
		txTo.Rollback()
		return fmt.Errorf("%s - failed to insert in coin_history to_user: %w", op, err)
	}

	if err = txFrom.Commit(); err != nil {
		txTo.Rollback()
		return err
	}

	if err = txTo.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *PgRepo) GetUserByUsername(ctx context.Context, username string) (entity.User, error) {
	for _, db := range p.ShardMap {
		var user entity.User

		query := "SELECT id, username, coins FROM users WHERE username = $1"

		err := db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.UserName, &user.Coins)
		if err == nil {
			return user, nil
		}
		if err != sql.ErrNoRows {
			return entity.User{}, err
		}
	}

	return entity.User{}, ErrUserNotFound
}

func (p *PgRepo) transferInSameShard(ctx context.Context, tx *sql.Tx, fromUserID, toUserID, amount int) error {
	const op = "PgRepo - transferInSameShard"

	query := "UPDATE users SET coins = coins - $1 " +
		"WHERE id = $2 AND coins >= $1"

	res, err := tx.ExecContext(ctx, query, amount, fromUserID)
	if err != nil {
		return fmt.Errorf("%s - failed to update coins on from_user: %w", op, err)
	}
	rowsAffecter, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s - failed to rowsAffected: %w", op, err)
	}
	if rowsAffecter == 0 {
		return ErrNotEnoughCoins
	}

	query = "UPDATE users SET coins = coins + $1 " +
		"WHERE id = $2"

	res, err = tx.ExecContext(ctx, query, amount, toUserID)
	if err != nil {
		return fmt.Errorf("%s - failed to update coins on to_user: %w", op, err)
	}

	query = "INSERT INTO coin_history (from_user, to_user, amount) " +
		"VALUES ($1, $2, $3)"
	_, err = tx.ExecContext(ctx, query, fromUserID, toUserID, amount)
	if err != nil {
		return fmt.Errorf("%s - failed to insert in coin_history: %w", op, err)
	}

	return nil
}

func (p *PgRepo) AddUser(ctx context.Context, user entity.User) (int, error) {
	query := "INSERT INTO users(id, username, coins) " +
		"VALUES ($1, $2, $3) " +
		"RETURNING id"

	shardNumber := p.getShardID(int(user.ID))

	db := p.ShardMap[shardNumber]

	var userID int

	err := db.QueryRowContext(ctx, query, user.ID, user.UserName, user.Coins).Scan(&userID)
	if err != nil || userID != int(user.ID) {
		return 0, fmt.Errorf("failed to add user: %w", err)
	}

	return userID, nil
}

func (p *PgRepo) Purchase(ctx context.Context, userID, value int, item string) error {
	query := "SELECT coins FROM users WHERE id = $1"

	var coins int

	shardNumber := p.getShardID(userID)

	db := p.ShardMap[shardNumber]

	err := db.QueryRowContext(ctx, query, userID).Scan(&coins)
	if err != nil {
		return fmt.Errorf("failed to get user coins: %w", err)
	}

	newCoins := coins - value

	if newCoins >= 0 {
		tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		defer tx.Rollback()

		query = "UPDATE users SET " +
			"coins = $1 " +
			"WHERE id = $2"

		_, err = tx.ExecContext(ctx, query, newCoins, userID)
		if err != nil {
			return fmt.Errorf("failed to update user coins: %w", err)
		}

		query = "INSERT INTO inventory (user_id, item, quantity) " +
			"VALUES ($1, $2, 1) " +
			"ON CONFLICT (user_id, item) " +
			"DO UPDATE SET quantity = inventory.quantity + 1"

		_, err = tx.ExecContext(ctx, query, userID, item)
		if err != nil {
			return fmt.Errorf("failed to update user inventory: %w", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	} else {
		return fmt.Errorf("failed to purchase item: %w", ErrNotEnoughCoins)
	}

	return nil
}

func (p *PgRepo) GetInfo(ctx context.Context, userID int) (entity.User, []entity.CoinHistory, error) {
	query := "SELECT coins FROM users WHERE users.id = $1"

	shardNumber := p.getShardID(userID)

	db := p.ShardMap[shardNumber]

	var coins int

	err := db.QueryRowContext(ctx, query, userID).Scan(&coins)
	if err != nil {
		return entity.User{}, nil, fmt.Errorf("failed to get user coins: %w", err)
	}

	query = "SELECT inv.item AS inventory_item, inv.quantity AS inventory_quantity " +
		"FROM users " +
		"LEFT JOIN inventory inv on users.id = inv.user_id " +
		"WHERE users.id = $1"

	rowsInv, err := db.QueryContext(ctx, query, userID)
	defer rowsInv.Close()

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, nil, fmt.Errorf("failed to get user inventory: %w", err)
	}

	var inventory []entity.Inventory

	for rowsInv.Next() {
		inv := entity.Inventory{}
		rowsInv.Scan(&inv.Item, &inv.Quantity)
		inventory = append(inventory, inv)
	}

	query = "SELECT to_user AS to_user, amount, 'sent' AS transaction_type " +
		"FROM coin_history " +
		"WHERE from_user = $1 " +
		"UNION " +
		"SELECT from_user AS from_user, amount, 'received' AS transaction_type " +
		"FROM coin_history " +
		"WHERE to_user = $1"

	rowsCoins, err := db.QueryContext(ctx, query, userID)
	defer rowsCoins.Close()

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, nil, fmt.Errorf("failed to get user history: %w", err)
	}

	var coinsHistory []entity.CoinHistory

	for rowsCoins.Next() {
		coinHistory := entity.CoinHistory{}
		rowsCoins.Scan(&coinHistory.ToUser, &coinHistory.Amount, &coinHistory.Type)
		if coinHistory.Type == "received" {
			coinHistory.ToUser, coinHistory.FromUser = coinHistory.FromUser, coinHistory.ToUser
		}

		coinsHistory = append(coinsHistory, coinHistory)
	}

	user := entity.User{
		ID:        int64(userID),
		Coins:     coins,
		Inventory: inventory,
	}

	return user, coinsHistory, nil
}

// check for implementation
var _ storage.UserRepo = (*PgRepo)(nil)

func NewRepo(dsns map[int]string) *PgRepo {
	return &PgRepo{
		ShardMap: initShardMap(dsns),
	}
}

func initShardMap(dsns map[int]string) ShardMap {
	m := make(ShardMap, len(dsns))
	for sh, dsn := range dsns {
		m[sh] = discoveryShard(dsn)
	}
	return m
}

// TODO: change to migrations
func (p *PgRepo) InitEntity(shardNum int) error {
	query := "CREATE TABLE IF NOT EXISTS users (id BIGINT PRIMARY KEY,username TEXT UNIQUE NOT NULL,coins INT)"

	rows, err := p.ShardMap[shardNum].Query(query)
	if err != nil {
		return fmt.Errorf("failed to migrate users :%w", err)
	}
	query = "CREATE TABLE IF NOT EXISTS inventory (id BIGSERIAL PRIMARY KEY, user_id BIGINT REFERENCES users(id), item TEXT NOT NULL, quantity INT)"

	rows, err = p.ShardMap[shardNum].Query(query)
	if err != nil {
		return fmt.Errorf("failed to migrate inventory :%w", err)
	}

	query = "ALTER TABLE inventory ADD CONSTRAINT inventory_unique_user_item UNIQUE (user_id, item)"

	rows, err = p.ShardMap[shardNum].Query(query)
	if err != nil {
		return fmt.Errorf("failed to migrate inventory :%w", err)
	}
	query = "CREATE TABLE IF NOT EXISTS coin_history (id BIGSERIAL PRIMARY KEY, from_user INT REFERENCES users(id), to_user INT REFERENCES users(id), amount INT, created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW())"

	rows, err = p.ShardMap[shardNum].Query(query)
	if err != nil {
		return fmt.Errorf("failed to migrate coin_history :%w", err)
	}

	_ = rows
	return nil
}

func discoveryShard(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

func (p *PgRepo) getShardID(ID int) int {
	return ID % BucketCount
}

func (p *PgRepo) CloseDB() {
	for _, db := range p.ShardMap {
		db.Close()
	}
}
