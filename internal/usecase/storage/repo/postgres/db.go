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
	"merch/internal/usecase/storage"
	"time"
)

var (
	ErrNotEnoughCoins = errors.New("not enough coins")
	ErrUserNotFound   = errors.New("user not found")
	log               = slog.Default()
)

const BucketCount = 4

type (
	ShardMap map[int]*pgxpool.Pool
)
type PgRepo struct {
	ShardMap ShardMap
}

// TODO: add op
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

	query := "UPDATE users SET coins = coins - $1 WHERE id = $2"
	_, err = txFrom.Exec(ctx, query, amount, fromUserID)
	if err != nil {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - failed to update coins from_user: %w", op, err)
	}

	query = "UPDATE users SET coins = coins + $1 WHERE id = $2"
	_, err = txTo.Exec(ctx, query, amount, toUserID)
	if err != nil {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - failed to update coins to_user: %w", op, err)
	}

	query = "INSERT INTO coin_history (from_user, to_user, amount) VALUES ($1, $2, $3)"
	_, err = txFrom.Exec(ctx, query, fromUserName, toUserName, amount)
	if err != nil {
		_ = txFrom.Rollback(ctx)
		_ = txTo.Rollback(ctx)
		return fmt.Errorf("%s - failed to insert in coin_history from_user: %w", op, err)
	}

	query = "INSERT INTO coin_history (from_user, to_user, amount) VALUES ($1, $2, $3)"
	_, err = txTo.Exec(ctx, query, fromUserName, toUserName, amount)
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

func (p *PgRepo) GetUserByUsername(ctx context.Context, username string) (entity.User, error) {
	for _, db := range p.ShardMap {
		var user entity.User

		query := "SELECT id, username, password, coins FROM users WHERE username = $1"

		err := db.QueryRow(ctx, query, username).Scan(&user.ID, &user.UserName, &user.Password, &user.Coins)
		if err == nil {
			return user, nil
		}
		if err != sql.ErrNoRows {
			return entity.User{}, err
		}
	}

	return entity.User{}, ErrUserNotFound
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

func (p *PgRepo) AddUser(ctx context.Context, user entity.User) (int, error) {
	query := "INSERT INTO users(id, username, password, coins) " +
		"VALUES ($1, $2, $3, $4)"

	shardNumber := p.getShardID(int(user.ID))

	db := p.ShardMap[shardNumber]

	_, err := db.Exec(ctx, query, user.ID, user.UserName, user.Password, user.Coins)
	if err != nil {
		return 0, fmt.Errorf("failed to add user: %w", err)
	}

	return int(user.ID), nil
}

func (p *PgRepo) Purchase(ctx context.Context, user entity.User, value int, item string) error {
	query := "SELECT coins FROM users WHERE id = $1"

	userID := int(user.ID)

	var coins int

	shardNumber := p.getShardID(userID)

	db := p.ShardMap[shardNumber]

	err := db.QueryRow(ctx, query, userID).Scan(&coins)
	if err != nil {
		return fmt.Errorf("failed to get user coins: %w", err)
	}

	newCoins := coins - value

	if newCoins >= 0 {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		defer tx.Rollback(ctx)

		query = "UPDATE users SET " +
			"coins = $1 " +
			"WHERE id = $2"

		_, err = tx.Exec(ctx, query, newCoins, userID)
		if err != nil {
			return fmt.Errorf("failed to update user coins: %w", err)
		}

		query = "INSERT INTO inventory (user_id, item, quantity) " +
			"VALUES ($1, $2, 1) " +
			"ON CONFLICT (user_id, item) " +
			"DO UPDATE SET quantity = inventory.quantity + 1"

		_, err = tx.Exec(ctx, query, userID, item)
		if err != nil {
			return fmt.Errorf("failed to update user inventory: %w", err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	} else {
		return fmt.Errorf("failed to purchase item: %w", ErrNotEnoughCoins)
	}

	return nil
}

func (p *PgRepo) GetInfo(ctx context.Context, user entity.User) (entity.User, []entity.CoinHistory, error) {
	query := "SELECT coins FROM users WHERE users.id = $1"

	userID := int(user.ID)

	shardNumber := p.getShardID(userID)

	db := p.ShardMap[shardNumber]

	var coins int

	err := db.QueryRow(ctx, query, userID).Scan(&coins)
	if err != nil {
		return entity.User{}, nil, fmt.Errorf("failed to get user coins: %w", err)
	}

	query = "SELECT inv.item AS inventory_item, inv.quantity AS inventory_quantity " +
		"FROM users " +
		"LEFT JOIN inventory inv on users.id = inv.user_id " +
		"WHERE users.id = $1"

	rowsInv, err := db.Query(ctx, query, userID)
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
		"UNION ALL " +
		"SELECT from_user AS from_user, amount, 'received' AS transaction_type " +
		"FROM coin_history " +
		"WHERE to_user = $1"

	rowsCoins, err := db.Query(ctx, query, user.UserName)
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

	user = entity.User{
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
// TODO: add index on users(username)
func (p *PgRepo) InitEntity(shardNum int) error {
	query := "CREATE TABLE IF NOT EXISTS users (" +
		"id BIGINT PRIMARY KEY, " +
		"username TEXT UNIQUE NOT NULL, " +
		"password TEXT NOT NULL, " +
		"coins INT)"

	rows, err := p.ShardMap[shardNum].Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to migrate users :%w", err)
	}
	//CREATE INDEX ON users (id); -- Для шардирования
	query = "CREATE UNIQUE INDEX users_username_idx ON users (username);\n"
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
		"to_user TEXT NOT NULL, " +
		"amount INT NOT NULL, " +
		"created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW())"

	rows, err = p.ShardMap[shardNum].Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to migrate coin_history :%w", err)
	}

	_ = rows
	return nil
}

func discoveryShard(dsn string) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("failed to parse config", slog.Any("error", err))
	}

	config.MaxConns = 100
	config.MinConns = 10
	config.MaxConnLifetime = 5 * time.Minute

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
