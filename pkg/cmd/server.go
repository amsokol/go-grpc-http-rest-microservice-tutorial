package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/amsokol/go-grpc-http-rest-microservice-tutorial/pkg/protocol/grpc"
	"github.com/amsokol/go-grpc-http-rest-microservice-tutorial/pkg/service/v1"
)

// Config is configuration for Server
type Config struct {
	// gRPC server start parameters section
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	// MySQL Datastore parameters section
	// DatastoreMySQLHost is host of MySQL database
	DatastoreMySQLHost string
	// DatastoreMySQLUser is username to connect to MySQL database
	DatastoreMySQLUser string
	// DatastoreMySQLPassword password to connect to MySQL database
	DatastoreMySQLPassword string
	// DatastoreMySQLSchema is schema of MySQL database
	DatastoreMySQLSchema string
	// DatastoreMySQLParams are parameters to connect to MySQL database
	DatastoreMySQLParams string
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	// get configuration
	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.DatastoreMySQLHost, "mysql-host", "", "MySQL database host")
	flag.StringVar(&cfg.DatastoreMySQLUser, "mysql-user", "", "MySQL database user")
	flag.StringVar(&cfg.DatastoreMySQLPassword, "mysql-password", "", "MySQL database password")
	flag.StringVar(&cfg.DatastoreMySQLSchema, "mysql-schema", "", "MySQL database schema")
	flag.StringVar(&cfg.DatastoreMySQLParams, "mysql-params", "", "MySQL database connection parameters")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	// add MySQL driver specific parameter to parse date/time
	if len(cfg.DatastoreMySQLParams) > 0 {
		cfg.DatastoreMySQLParams += "&"
	}
	cfg.DatastoreMySQLParams += "parseTime=true"

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.DatastoreMySQLUser,
		cfg.DatastoreMySQLPassword,
		cfg.DatastoreMySQLHost,
		cfg.DatastoreMySQLSchema,
		cfg.DatastoreMySQLParams)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	v1API := v1.NewToDoServiceServer(db)

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort)
}
