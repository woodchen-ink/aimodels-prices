package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"aimodels-prices/config"
	"aimodels-prices/database"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 检查SQLite数据库文件是否存在
	if _, err := os.Stat(cfg.SQLitePath); os.IsNotExist(err) {
		log.Printf("SQLite database file not found at %s, skipping migration", cfg.SQLitePath)
		os.Exit(0)
	}

	// 连接SQLite数据库
	sqliteDB, err := database.InitSQLiteDB(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite database: %v", err)
	}
	defer sqliteDB.Close()

	// 初始化MySQL数据库
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize MySQL database: %v", err)
	}
	defer database.DB.Close()

	// 开始迁移数据
	if err := migrateData(sqliteDB, database.DB); err != nil {
		log.Fatalf("Failed to migrate data: %v", err)
	}

	log.Println("Data migration completed successfully!")
}

func migrateData(sqliteDB *sql.DB, mysqlDB *sql.DB) error {
	// 迁移用户数据
	if err := migrateUsers(sqliteDB, mysqlDB); err != nil {
		return fmt.Errorf("failed to migrate users: %v", err)
	}

	// 迁移会话数据
	if err := migrateSessions(sqliteDB, mysqlDB); err != nil {
		return fmt.Errorf("failed to migrate sessions: %v", err)
	}

	// 迁移提供商数据
	if err := migrateProviders(sqliteDB, mysqlDB); err != nil {
		return fmt.Errorf("failed to migrate providers: %v", err)
	}

	// 迁移价格数据
	if err := migratePrices(sqliteDB, mysqlDB); err != nil {
		return fmt.Errorf("failed to migrate prices: %v", err)
	}

	return nil
}

func migrateUsers(sqliteDB *sql.DB, mysqlDB *sql.DB) error {
	rows, err := sqliteDB.Query("SELECT id, username, email, role, created_at, updated_at, deleted_at FROM user")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        uint
			username  string
			email     string
			role      string
			createdAt string
			updatedAt string
			deletedAt sql.NullString
		)

		if err := rows.Scan(&id, &username, &email, &role, &createdAt, &updatedAt, &deletedAt); err != nil {
			return err
		}

		_, err = mysqlDB.Exec(
			"INSERT INTO user (id, username, email, role, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			id, username, email, role, createdAt, updatedAt, deletedAt.String,
		)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func migrateSessions(sqliteDB *sql.DB, mysqlDB *sql.DB) error {
	rows, err := sqliteDB.Query("SELECT id, user_id, expires_at, created_at, updated_at, deleted_at FROM session")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        string
			userID    uint
			expiresAt string
			createdAt string
			updatedAt string
			deletedAt sql.NullString
		)

		if err := rows.Scan(&id, &userID, &expiresAt, &createdAt, &updatedAt, &deletedAt); err != nil {
			return err
		}

		_, err = mysqlDB.Exec(
			"INSERT INTO session (id, user_id, expires_at, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?)",
			id, userID, expiresAt, createdAt, updatedAt, deletedAt.String,
		)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func migrateProviders(sqliteDB *sql.DB, mysqlDB *sql.DB) error {
	rows, err := sqliteDB.Query("SELECT id, name, icon, created_at, updated_at, created_by FROM provider")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        uint
			name      string
			icon      sql.NullString
			createdAt string
			updatedAt string
			createdBy string
		)

		if err := rows.Scan(&id, &name, &icon, &createdAt, &updatedAt, &createdBy); err != nil {
			return err
		}

		_, err = mysqlDB.Exec(
			"INSERT INTO provider (id, name, icon, created_at, updated_at, created_by) VALUES (?, ?, ?, ?, ?, ?)",
			id, name, icon.String, createdAt, updatedAt, createdBy,
		)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func migratePrices(sqliteDB *sql.DB, mysqlDB *sql.DB) error {
	rows, err := sqliteDB.Query(`
		SELECT id, model, model_type, billing_type, channel_type, currency, 
		input_price, output_price, price_source, status, created_at, updated_at,
		created_by, temp_model, temp_model_type, temp_billing_type, temp_channel_type,
		temp_currency, temp_input_price, temp_output_price, temp_price_source, updated_by
		FROM price
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id              uint
			model           string
			modelType       string
			billingType     string
			channelType     uint
			currency        string
			inputPrice      float64
			outputPrice     float64
			priceSource     string
			status          string
			createdAt       string
			updatedAt       string
			createdBy       string
			tempModel       sql.NullString
			tempModelType   sql.NullString
			tempBillingType sql.NullString
			tempChannelType sql.NullInt64
			tempCurrency    sql.NullString
			tempInputPrice  sql.NullFloat64
			tempOutputPrice sql.NullFloat64
			tempPriceSource sql.NullString
			updatedBy       sql.NullString
		)

		if err := rows.Scan(
			&id, &model, &modelType, &billingType, &channelType, &currency,
			&inputPrice, &outputPrice, &priceSource, &status, &createdAt, &updatedAt,
			&createdBy, &tempModel, &tempModelType, &tempBillingType, &tempChannelType,
			&tempCurrency, &tempInputPrice, &tempOutputPrice, &tempPriceSource, &updatedBy,
		); err != nil {
			return err
		}

		_, err = mysqlDB.Exec(`
			INSERT INTO price (
				id, model, model_type, billing_type, channel_type, currency,
				input_price, output_price, price_source, status, created_at, updated_at,
				created_by, temp_model, temp_model_type, temp_billing_type, temp_channel_type,
				temp_currency, temp_input_price, temp_output_price, temp_price_source, updated_by
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			id, model, modelType, billingType, channelType, currency,
			inputPrice, outputPrice, priceSource, status, createdAt, updatedAt,
			createdBy, tempModel.String, tempModelType.String, tempBillingType.String, tempChannelType.Int64,
			tempCurrency.String, tempInputPrice.Float64, tempOutputPrice.Float64, tempPriceSource.String, updatedBy.String,
		)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}
