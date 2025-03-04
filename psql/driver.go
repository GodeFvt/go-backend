package psql

type Driver string

const opentracing_postgres = "ot_pg"
const Mssql Driver = "sqlserver" // sqlserver
const Postgres Driver = "pgx"    // postgresql
const Clickhouse Driver = "clickhouse"
