package psql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	pg "github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"github.com/qustavo/sqlhooks/v2"
)

type Client struct {
	db            *sqlx.DB
	connectionURI string
	driverName    string
	tracer        opentracing.Tracer
}

func connectPostgres(ctx context.Context, connectionStr string, databaseType Driver, tracing opentracing.Tracer) (client *Client, err error) {
	pool, err := pgxpool.New(ctx, connectionStr)
	if err != nil {
		return nil, err
	}
	var db *sqlx.DB
	var driverName Driver

	if tracing != nil {
		switch driverName {
		case Postgres:
			sql.Register(opentracing_postgres, sqlhooks.Wrap(&pg.Driver{}, NewTracingHook(tracing)))
		}
	} else {
		driverName = databaseType
	}

	db = sqlx.NewDb(stdlib.OpenDBFromPool(pool), string(driverName))
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)

	return &Client{
		db:            db,
		connectionURI: connectionStr,
		driverName:    string(driverName),
		tracer:        tracing,
	}, nil
}

func connect(ctx context.Context, connectionStr string, databaseType Driver, tracing opentracing.Tracer) (client *Client, err error) {
	if databaseType == Postgres {
		return connectPostgres(ctx, connectionStr, databaseType, tracing)
	}

	var driver Driver
	switch databaseType {
	case Mssql:
		driver = Mssql
	case Clickhouse:
		driver = Clickhouse
	}

	db, err := sqlx.Connect(string(driver), connectionStr)
	if err != nil {
		return nil, err
	}

	return &Client{
		db:            db,
		connectionURI: connectionStr,
		driverName:    string(driver),
		tracer:        tracing,
	}, nil
}

func NewConnection(connectionStr string, databaseType Driver) (*Client, error) {
	return connect(context.Background(), connectionStr, databaseType, nil)
}

func NewConnectionWithTracing(connectionStr string, databaseType Driver, tracing opentracing.Tracer) (client *Client, err error) {
	return connect(context.Background(), connectionStr, databaseType, tracing)
}

func (c *Client) GetClient() *sqlx.DB {
	return c.db
}

func (c *Client) GetConnectionURI() string {
	return c.connectionURI
}

func (c *Client) SetDB(db *sqlx.DB) {
	c.db = db
}

func (c *Client) IsConnect() bool {
	if err := c.db.Ping(); err == nil {
		return true
	}
	return false
}
func (c *Client) Close() error {
	return c.db.Close()
}
