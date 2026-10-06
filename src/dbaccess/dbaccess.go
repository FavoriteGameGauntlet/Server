package dbaccess

import (
	"FGG-Service/src/auth/types"
	"context"
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func Init() func() {
	var err error
	db, err = sql.Open("pgx", os.Getenv(DatabaseURLEnvVar))

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)
	}

	return func() {
		_ = db.Close()
	}
}

func Exec(ctx context.Context, q Query, args ...any) (sql.Result, error) {
	queryArgs, logArgs := getArgs(args...)

	if !q.IsSilent {
		slog.Info(q.Name, "args", logArgs)
	}

	result, err := db.ExecContext(ctx, q.SQL, queryArgs...)

	if err != nil {
		slog.Error(q.Name, "error", err)
	}

	return result, err
}

func QueryRows(ctx context.Context, q Query, args ...any) (*sql.Rows, error) {
	queryArgs, logArgs := getArgs(args...)

	if !q.IsSilent {
		slog.Info(q.Name, "args", logArgs)
	}

	rows, err := db.QueryContext(ctx, q.SQL, queryArgs...)

	if err != nil {
		slog.Error(q.Name, "error", err)
	}

	return rows, err
}

func QueryRow(ctx context.Context, q Query, args ...any) *sql.Row {
	queryArgs, logArgs := getArgs(args...)

	if !q.IsSilent {
		slog.Info(q.Name, "args", logArgs)
	}

	return db.QueryRowContext(ctx, q.SQL, queryArgs...)
}

func getArgs(args ...any) ([]any, []any) {
	logArgs := make([]any, len(args))
	queryArgs := make([]any, len(args))

	for i, arg := range args {
		if p, ok := arg.(typeauth.Password); ok {
			logArgs[i] = "***"
			queryArgs[i] = p.Value
		} else {
			logArgs[i] = arg
			queryArgs[i] = arg
		}
	}

	return queryArgs, logArgs
}
