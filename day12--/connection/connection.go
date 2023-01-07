package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var Conn *pgx.Conn

func DatabaseConnect() {
	// databaseURL := "postgres://username:password@localhost:5432/database_name"

	databaseURL := "postgres://postgres:yulfaid4@localhost:5432/project_43"

	var err error
	Conn, err = pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully connect to database.")
	// defer Conn.Close(context.Background())
}
