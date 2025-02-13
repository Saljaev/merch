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
)

type (
	ShardMap map[int]*sql.DB
)
type PgRepo struct {
	ShardMap ShardMap
}

func (p *PgRepo) Transfer(ctx context.Context, fromUserID, toUserID, amount int) error {
	//TODO implement me
	panic("implement me")
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

func (p *PgRepo) GetInfo(ctx context.Context, userID int) (*entity.User, []*entity.CoinHistory, error) {
	//TODO implement me
	panic("implement me")
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

	query = "CREATE TABLE IF NOT EXISTS coin_history (id BIGSERIAL PRIMARY KEY, user_id BIGINT REFERENCES users(id), from_user TEXT NOT NULL, to_user TEXT NOT NULL, amount INT, type TEXT CHECK (type IN ('received', 'sent')), created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW())"

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
	return ID % storage.BucketCount
}

func (p *PgRepo) CloseDB() {
	for _, db := range p.ShardMap {
		db.Close()
	}
}
