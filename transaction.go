package database

import "context"

type txKeyType struct{}

var txKey txKeyType

func WithTx(ctx context.Context, tx Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func TxFromCtx(ctx context.Context) (Tx, bool) {
	tx, ok := ctx.Value(txKey).(Tx)
	return tx, ok
}

func RunInTx(ctx context.Context, db DB, fn func(ctx context.Context) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
			panic(p)
		}
	}()

	fnCtx := WithTx(ctx, tx)
	if err := fn(fnCtx); err != nil {
		_ = tx.Rollback(context.Background())
		return err
	}

	return tx.Commit(ctx)
}
