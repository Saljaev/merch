package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"log/slog"
	"merch/internal/entity"
	"merch/internal/usecase/usecase"
	"time"
)

var (
	log = slog.Default()
)

const BucketCount = 4

type (
	ShardMap map[int]*pgxpool.Pool
)
type PgRepo struct {
	ShardMap ShardMap
}

func (p *PgRepo) Transfer(ctx context.Context, amount int, fromUser, toUser entity.User) error {
	const op = "PgRepo - Transfer"

	fromUserID := int(fromUser.ID)
	toUserID := int(toUser.ID)

	fromUserName := fromUser.UserName
	toUserName := toUser.UserName

	shardFrom := p.getShardID(fromUserID)
	shardTo := p.getShardID(toUserID)

	fromDB := p.ShardMap[shardFrom]
	toDB := p.ShardMap[shardTo]

	if fromDB == toDB {
		tx, err := fromDB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
		if err != nil {
			return fmt.Errorf("%s - failed to start transaction: %w", op, err)
		}

		err = p.transferInSameShard(ctx, tx, amount, fromUser, toUser)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}

		return tx.Commit(ctx)
	}

	txFrom, err := fromDB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("%s - failed to BeginTx from_user: %w", op, err)
	}

	txTo, err := toDB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		_ = txFrom.Rollback(ctx)
		return fmt.Errorf("%s - failed to BeginTx to_user: %w", op, err)
	}

	query := "UPDATE users SET coins = coins - $1 WHERE id = $2 AND coins >= $1"
	row, err := txFrom.Exec(ctx, query, amount, fromUserID)
	if err != nil {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - failed to update coins from_user: %w", op, err)
	}
	if row.RowsAffected() == 0 {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - %w", op, usecase.ErrNotEnoughCoin)
	}

	query = "UPDATE users SET coins = coins + $1 WHERE id = $2"
	_, err = txTo.Exec(ctx, query, amount, toUserID)
	if err != nil {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - failed to update coins to_user: %w", op, err)
	}

	query = "INSERT INTO coin_history (from_user, from_user_id, to_user, to_user_id, amount) VALUES ($1, $2, $3, $4, $5)"
	_, err = txFrom.Exec(ctx, query, fromUserName, fromUserID, toUserName, toUserID, amount)
	if err != nil {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - failed to insert in coin_history from_user: %w", op, err)
	}

	query = "INSERT INTO coin_history (from_user, from_user_id, to_user, to_user_id, amount) VALUES ($1, $2, $3, $4, $5)"
	_, err = txTo.Exec(ctx, query, fromUserName, fromUserID, toUserName, toUserID, amount)
	if err != nil {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - failed to insert in coin_history to_user: %w", op, err)
	}

	if err = txFrom.Commit(ctx); err != nil {
		_ = txTo.Rollback(ctx)
		return err
	}

	if err = txTo.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *PgRepo) GetUserByID(ctx context.Context, ID int) (entity.User, error) {
	const op = "PgRepo - GetUserByID"

	for _, db := range p.ShardMap {
		var user entity.User

		query := "SELECT id, username, password, coins FROM users WHERE id = $1"

		err := db.QueryRow(ctx, query, ID).Scan(&user.ID, &user.UserName, &user.Password, &user.Coins)
		if err == nil {
			return user, nil
		}
	}

	return entity.User{}, fmt.Errorf("%s - %w", op, usecase.ErrUserNotFound)
}

func (p *PgRepo) GetUserByUsername(ctx context.Context, username string) (entity.User, error) {
	const op = "PgRepo - GetUserByUsername"

	for _, db := range p.ShardMap {
		var user entity.User

		query := "SELECT id, username, password, coins FROM users WHERE username = $1"

		err := db.QueryRow(ctx, query, username).Scan(&user.ID, &user.UserName, &user.Password, &user.Coins)
		if err == nil {
			return user, nil
		}
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, fmt.Errorf("%s - db.QueryRow: %w", op, usecase.ErrUserNotFound)
		}
	}

	return entity.User{}, fmt.Errorf("%s - %w", op, usecase.ErrUserNotFound)
}

func (p *PgRepo) transferInSameShard(ctx context.Context, tx pgx.Tx, amount int, fromUser, toUser entity.User) error {
	const op = "PgRepo - transferInSameShard"

	fromUserID := int(fromUser.ID)
	toUserID := int(toUser.ID)

	fromUserName := fromUser.UserName
	toUserName := toUser.UserName

	query := "UPDATE users SET coins = coins - $1 " +
		"WHERE id = $2"

	_, err := tx.Exec(ctx, query, amount, fromUserID)
	if err != nil {
		return fmt.Errorf("%s - failed to update coins on from_user: %w", op, err)
	}

	query = "UPDATE users SET coins = coins + $1 " +
		"WHERE id = $2"

	_, err = tx.Exec(ctx, query, amount, toUserID)
	if err != nil {
		return fmt.Errorf("%s - failed to update coins on to_user: %w", op, err)
	}

	query = "INSERT INTO coin_history (from_user, to_user, amount) " +
		"VALUES ($1, $2, $3)"
	_, err = tx.Exec(ctx, query, fromUserName, toUserName, amount)
	if err != nil {
		return fmt.Errorf("%s - failed to insert in coin_history: %w", op, err)
	}

	return nil
}

func (p *PgRepo) AddUser(ctx context.Context, user entity.User) error {
	const op = "PgRepo - AddUser"

	query := "INSERT INTO users(id, username, password, coins) " +
		"VALUES ($1, $2, $3, $4)"

	shardNumber := p.getShardID(int(user.ID))

	db := p.ShardMap[shardNumber]

	_, err := db.Exec(ctx, query, user.ID, user.UserName, user.Password, user.Coins)
	if err != nil {
		return fmt.Errorf("%s - failed to add user: %w", op, err)
	}

	return nil
}
func (p *PgRepo) Purchase(ctx context.Context, userID int, coins, value int, item string) error {
	const op = "PgRepo - Purchase"

	//userID := int(user.ID)
	shardNumber := p.getShardID(userID)
	db := p.ShardMap[shardNumber]

	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("%s - failed to start transaction: %w", op, err)
	}

	defer tx.Rollback(ctx)

	query := "UPDATE users SET " +
		"coins = $1 " +
		"WHERE id = $2"

	_, err = tx.Exec(ctx, query, coins-value, userID)
	if err != nil {
		return fmt.Errorf("%s - failed to update user coins: %w", op, err)
	}

	query = "INSERT INTO inventory (user_id, item, quantity) " +
		"VALUES ($1, $2, 1) " +
		"ON CONFLICT (user_id, item) " +
		"DO UPDATE SET quantity = inventory.quantity + 1"

	_, err = tx.Exec(ctx, query, userID, item)
	if err != nil {
		return fmt.Errorf("%s - failed to update user inventory: %w", op, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s - failed to commit transaction: %w", op, err)
	}

	return nil
}

func (p *PgRepo) GetInfo(ctx context.Context, userID int) (entity.User, []entity.CoinHistory, error) {
	const op = "PgRepo - GetInfo"

	shardNumber := p.getShardID(userID)
	db := p.ShardMap[shardNumber]

	query := "SELECT coins FROM users WHERE users.id = $1"
	var coins int

	err := db.QueryRow(ctx, query, userID).Scan(&coins)
	if err != nil {
		return entity.User{}, nil, fmt.Errorf("%s - failed to get user coins: %w", op, err)
	}

	query = "SELECT inv.item AS inventory_item, inv.quantity AS inventory_quantity " +
		"FROM users " +
		"LEFT JOIN inventory inv on users.id = inv.user_id " +
		"WHERE users.id = $1"

	rowsInv, err := db.Query(ctx, query, userID)
	defer rowsInv.Close()

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, nil, fmt.Errorf("%s - failed to get user inventory: %w", op, err)
	}

	var inventory []entity.Inventory

	for rowsInv.Next() {
		inv := entity.Inventory{}
		rowsInv.Scan(&inv.Item, &inv.Quantity)
		inventory = append(inventory, inv)
	}

	user := entity.User{
		ID:        int64(userID),
		Coins:     coins,
		Inventory: inventory,
	}

	coinsHistory, err := p.GetTransaction(ctx, userID)
	if err != nil {
		return entity.User{}, nil, err
	}

	return user, coinsHistory, nil
}

func (p *PgRepo) GetTransaction(ctx context.Context, userID int) ([]entity.CoinHistory, error) {
	const op = "PgRepo - GetTransaction"

	shardNumber := p.getShardID(userID)

	db := p.ShardMap[shardNumber]

	query := "SELECT to_user AS to_user, amount, 'sent' AS transaction_type " +
		"FROM coin_history " +
		"WHERE from_user_id = $1 " +
		"UNION ALL " +
		"SELECT from_user AS from_user, amount, 'received' AS transaction_type " +
		"FROM coin_history " +
		"WHERE to_user_id = $1"

	rowsCoins, err := db.Query(ctx, query, userID)
	defer rowsCoins.Close()

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s - failed to get user history: %w", op, err)
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

	return coinsHistory, nil
}

// check for implementation
var _ usecase.UserRepo = (*PgRepo)(nil)

func NewRepo(dsns map[int]string, maxConn, minConn int, lifeConn time.Duration) *PgRepo {
	return &PgRepo{
		ShardMap: initShardMap(dsns, maxConn, minConn, lifeConn),
	}
}

func initShardMap(dsns map[int]string, maxConn, minConn int, lifeConn time.Duration) ShardMap {
	m := make(ShardMap, len(dsns))
	for sh, dsn := range dsns {
		m[sh] = discoveryShard(dsn, maxConn, minConn, lifeConn)
	}
	return m
}

// TODO: change to migrations
// TODO: add index on users(username)
func (p *PgRepo) InitEntity(shardNum int) error {
	query := "CREATE TABLE IF NOT EXISTS users (" +
		"id BIGINT PRIMARY KEY, " +
		"username TEXT UNIQUE NOT NULL, " +
		"password TEXT NOT NULL, " +
		"coins INT)"

	rows, err := p.ShardMap[shardNum].Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to migrate users: %w", err)
	}
	query = "CREATE UNIQUE INDEX users_username_idx ON users (username);"
	rows, err = p.ShardMap[shardNum].Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create users index:%w", err)
	}

	query = "CREATE TABLE IF NOT EXISTS inventory (id BIGSERIAL PRIMARY KEY, user_id BIGINT REFERENCES users(id), item TEXT NOT NULL, quantity INT)"

	rows, err = p.ShardMap[shardNum].Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to migrate inventory :%w", err)
	}

	query = "ALTER TABLE inventory ADD CONSTRAINT inventory_unique_user_item UNIQUE (user_id, item)"

	rows, err = p.ShardMap[shardNum].Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to migrate inventory :%w", err)
	}
	query = "CREATE TABLE IF NOT EXISTS coin_history (" +
		"id BIGSERIAL PRIMARY KEY, " +
		"from_user TEXT NOT NULL, " +
		"from_user_id BIGINT, " +
		"to_user TEXT NOT NULL, " +
		"to_user_id BIGINT, " +
		"amount INT NOT NULL, " +
		"created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW())"

	rows, err = p.ShardMap[shardNum].Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to migrate coin_history :%w", err)
	}

	_ = rows
	return nil
}

func discoveryShard(dsn string, maxConn, minConn int, lifeConn time.Duration) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("failed to parse config", slog.Any("error", err))
	}

	config.MaxConns = int32(maxConn)
	config.MinConns = int32(minConn)
	config.MaxConnLifetime = lifeConn

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Error("failed new pgxpool with config", slog.Any("error", err))
	}

	err = pool.Ping(context.Background())
	if err != nil {
		log.Error("failed ping shard", slog.Any("error", err))
	}

	return pool
}

func (p *PgRepo) getShardID(ID int) int {
	return ID % BucketCount
}

func (p *PgRepo) CloseDB() {
	for _, pool := range p.ShardMap {
		pool.Close()
	}
}
