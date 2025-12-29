package database

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"time"

	"x-ui/config"
	"x-ui/database/model"
	"x-ui/util/crypto"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

const (
	defaultUsername = "admin"
	defaultPassword = "admin"
)

// fixClientTrafficsInboundId 为旧数据的 inbound_id 设置默认值
// 这个函数在 AutoMigrate 之前执行，确保所有记录都有有效的 inbound_id
func fixClientTrafficsInboundId() error {
	log.Println("Checking client_traffics table for inbound_id issues...")

	// 检查表是否存在
	var tableCount int64
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='client_traffics'").Scan(&tableCount).Error
	if err != nil {
		return err
	}

	// 如果表不存在，跳过
	if tableCount == 0 {
		log.Println("client_traffics table does not exist, will be created by AutoMigrate")
		return nil
	}

	log.Println("client_traffics table exists, checking for inbound_id column...")

	// 检查是否有 inbound_id 列
	var columnCount int64
	err = db.Raw("SELECT COUNT(*) FROM pragma_table_info('client_traffics') WHERE name='inbound_id'").Scan(&columnCount).Error
	if err != nil {
		return err
	}

	// 如果没有 inbound_id 列，说明是旧表，需要添加并设置默认值
	if columnCount == 0 {
		log.Println("inbound_id column does not exist, adding it with default value 1...")
		// 使用 DEFAULT 1 添加列，这样所有现有行都会自动获得值 1
		err = db.Exec("ALTER TABLE client_traffics ADD COLUMN inbound_id INTEGER NOT NULL DEFAULT 1").Error
		if err != nil {
			log.Printf("Failed to add inbound_id column: %v", err)
			return err
		}
		log.Println("inbound_id column added successfully with default value 1")

		// 验证添加是否成功
		var verifyCount int64
		db.Raw("SELECT COUNT(*) FROM client_traffics WHERE inbound_id = 1").Scan(&verifyCount)
		log.Printf("Verified: %d records now have inbound_id = 1", verifyCount)

		return nil
	}

	log.Println("inbound_id column exists, checking for NULL or 0 values...")

	// 如果列已存在，检查并更新 NULL 或 0 的值
	var nullCount int64
	err = db.Raw("SELECT COUNT(*) FROM client_traffics WHERE inbound_id IS NULL OR inbound_id = 0").Scan(&nullCount).Error
	if err != nil {
		return err
	}

	if nullCount > 0 {
		log.Printf("Found %d records with NULL or 0 inbound_id, updating to 1...", nullCount)
		result := db.Exec("UPDATE client_traffics SET inbound_id = 1 WHERE inbound_id IS NULL OR inbound_id = 0")
		if result.Error != nil {
			log.Printf("Failed to update inbound_id: %v", result.Error)
			return result.Error
		}
		log.Printf("Successfully updated %d records with default inbound_id = 1", result.RowsAffected)
	} else {
		log.Println("All records have valid inbound_id values")
	}

	return nil
}

// ensureClientTrafficsSchema 确保 client_traffics 表的结构和索引正确
func ensureClientTrafficsSchema() error {
	log.Println("Ensuring client_traffics table schema...")

	// 检查表是否存在
	var tableCount int64
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='client_traffics'").Scan(&tableCount).Error
	if err != nil {
		return err
	}

	// 如果表不存在，创建它
	if tableCount == 0 {
		log.Println("Creating client_traffics table...")
		err = db.Exec(`
			CREATE TABLE client_traffics (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				inbound_id INTEGER NOT NULL DEFAULT 1,
				enable INTEGER NOT NULL,
				email TEXT NOT NULL,
				up INTEGER NOT NULL,
				down INTEGER NOT NULL,
				all_time INTEGER NOT NULL,
				expiry_time INTEGER NOT NULL,
				total INTEGER NOT NULL,
				reset INTEGER NOT NULL DEFAULT 0,
				last_online INTEGER NOT NULL DEFAULT 0
			)
		`).Error
		if err != nil {
			log.Printf("Failed to create client_traffics table: %v", err)
			return err
		}
		log.Println("client_traffics table created successfully")
	} else {
		// 表已存在，检查是否需要重建（删除旧的 uni_client_traffics_email 约束）
		log.Println("Checking for old unique constraint on email column...")
		var constraintSQL string
		db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='client_traffics'").Scan(&constraintSQL)

		// 如果表定义中包含旧的 uni_client_traffics_email 约束，需要重建表
		if len(constraintSQL) > 0 && (db.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name='client_traffics' AND sql LIKE '%uni_client_traffics_email%'").Scan(&constraintSQL).Error == nil && constraintSQL != "") {
			log.Println("Found old uni_client_traffics_email constraint, rebuilding table...")

			// 开始事务
			tx := db.Begin()

			// 创建新表（没有旧约束）
			err = tx.Exec(`
				CREATE TABLE client_traffics_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					inbound_id INTEGER NOT NULL DEFAULT 1,
					enable INTEGER NOT NULL,
					email TEXT NOT NULL,
					up INTEGER NOT NULL,
					down INTEGER NOT NULL,
					all_time INTEGER NOT NULL,
					expiry_time INTEGER NOT NULL,
					total INTEGER NOT NULL,
					reset INTEGER NOT NULL DEFAULT 0,
					last_online INTEGER NOT NULL DEFAULT 0
				)
			`).Error
			if err != nil {
				tx.Rollback()
				log.Printf("Failed to create new table: %v", err)
				return err
			}

			// 复制数据
			err = tx.Exec("INSERT INTO client_traffics_new SELECT id, inbound_id, enable, email, up, down, all_time, expiry_time, total, reset, last_online FROM client_traffics").Error
			if err != nil {
				tx.Rollback()
				log.Printf("Failed to copy data: %v", err)
				return err
			}

			// 删除旧表
			err = tx.Exec("DROP TABLE client_traffics").Error
			if err != nil {
				tx.Rollback()
				log.Printf("Failed to drop old table: %v", err)
				return err
			}

			// 重命名新表
			err = tx.Exec("ALTER TABLE client_traffics_new RENAME TO client_traffics").Error
			if err != nil {
				tx.Rollback()
				log.Printf("Failed to rename table: %v", err)
				return err
			}

			// 提交事务
			if err = tx.Commit().Error; err != nil {
				log.Printf("Failed to commit transaction: %v", err)
				return err
			}

			log.Println("Table rebuilt successfully without old constraint")
		}
	}

	// 检查联合唯一索引是否存在
	var indexCount int64
	err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_inbound_email'").Scan(&indexCount).Error
	if err != nil {
		return err
	}

	// 如果索引不存在，创建它
	if indexCount == 0 {
		log.Println("Creating unique index idx_inbound_email on client_traffics...")
		err = db.Exec("CREATE UNIQUE INDEX idx_inbound_email ON client_traffics(inbound_id, email)").Error
		if err != nil {
			log.Printf("Failed to create unique index: %v", err)
			return err
		}
		log.Println("Unique index idx_inbound_email created successfully")
	} else {
		log.Println("Unique index idx_inbound_email already exists")
	}

	// 检查旧的 email 唯一索引是否存在，如果存在则删除
	var oldIndexCount int64
	err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE '%email%' AND name != 'idx_inbound_email' AND tbl_name='client_traffics'").Scan(&oldIndexCount).Error
	if err == nil && oldIndexCount > 0 {
		// 获取旧索引的名称
		var oldIndexName string
		db.Raw("SELECT name FROM sqlite_master WHERE type='index' AND name LIKE '%email%' AND name != 'idx_inbound_email' AND tbl_name='client_traffics' LIMIT 1").Scan(&oldIndexName)
		if oldIndexName != "" {
			log.Printf("Dropping old email unique index: %s", oldIndexName)
			db.Exec("DROP INDEX IF EXISTS " + oldIndexName)
		}
	}

	log.Println("client_traffics table schema is correct")
	return nil
}

func initModels() error {
	// 在 AutoMigrate 之前，确保旧数据有有效的 inbound_id
	if err := fixClientTrafficsInboundId(); err != nil {
		log.Printf("Error fixing client_traffics inbound_id: %v", err)
		return err
	}

	// 手动确保 client_traffics 表结构正确（包括索引）
	if err := ensureClientTrafficsSchema(); err != nil {
		log.Printf("Error ensuring client_traffics schema: %v", err)
		return err
	}

	// 对其他表使用 AutoMigrate
	models := []any{
		&model.User{},
		&model.Inbound{},
		&model.OutboundTraffics{},
		&model.Setting{},
		&model.InboundClientIps{},
		// &xray.ClientTraffic{}, // 手动处理，不使用 AutoMigrate
		&model.HistoryOfSeeders{},
		&LinkHistory{},      // 把 LinkHistory 表也迁移
		&model.LotteryWin{}, // 新增 抽奖游戏LotteryWin 数据模型
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Error auto migrating model: %v", err)
			return err
		}
	}
	return nil
}

func initUser() error {
	empty, err := isTableEmpty("users")
	if err != nil {
		log.Printf("Error checking if users table is empty: %v", err)
		return err
	}
	if empty {
		hashedPassword, err := crypto.HashPasswordAsBcrypt(defaultPassword)

		if err != nil {
			log.Printf("Error hashing default password: %v", err)
			return err
		}

		user := &model.User{
			Username: defaultUsername,
			Password: hashedPassword,
		}
		return db.Create(user).Error
	}
	return nil
}

func runSeeders(isUsersEmpty bool) error {
	empty, err := isTableEmpty("history_of_seeders")
	if err != nil {
		log.Printf("Error checking if users table is empty: %v", err)
		return err
	}

	if empty && isUsersEmpty {
		hashSeeder := &model.HistoryOfSeeders{
			SeederName: "UserPasswordHash",
		}
		return db.Create(hashSeeder).Error
	} else {
		var seedersHistory []string
		db.Model(&model.HistoryOfSeeders{}).Pluck("seeder_name", &seedersHistory)

		if !slices.Contains(seedersHistory, "UserPasswordHash") && !isUsersEmpty {
			var users []model.User
			db.Find(&users)

			for _, user := range users {
				hashedPassword, err := crypto.HashPasswordAsBcrypt(user.Password)
				if err != nil {
					log.Printf("Error hashing password for user '%s': %v", user.Username, err)
					return err
				}
				db.Model(&user).Update("password", hashedPassword)
			}

			hashSeeder := &model.HistoryOfSeeders{
				SeederName: "UserPasswordHash",
			}
			return db.Create(hashSeeder).Error
		}
	}

	return nil
}

func isTableEmpty(tableName string) (bool, error) {
	var count int64
	err := db.Table(tableName).Count(&count).Error
	return count == 0, err
}

func InitDB(dbPath string) error {
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}

	var gormLogger logger.Interface

	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}
	db, err = gorm.Open(sqlite.Open(dbPath), c)
	if err != nil {
		return err
	}

	if err := initModels(); err != nil {
		return err
	}

	isUsersEmpty, err := isTableEmpty("users")

	if err := initUser(); err != nil {
		return err
	}
	return runSeeders(isUsersEmpty)
}

func CloseDB() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func GetDB() *gorm.DB {
	return db
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func IsSQLiteDB(file io.ReaderAt) (bool, error) {
	signature := []byte("SQLite format 3\x00")
	buf := make([]byte, len(signature))
	_, err := file.ReadAt(buf, 0)
	if err != nil {
		return false, err
	}
	return bytes.Equal(buf, signature), nil
}

func Checkpoint() error {
	// Update WAL
	err := db.Exec("PRAGMA wal_checkpoint;").Error
	if err != nil {
		return err
	}
	return nil
}

// HasUserWonToday 检查指定用户今天是否已经中过奖
// 〔中文注释〕:【修正】将 gorm.DB() 替换为全局变量 db
func HasUserWonToday(userID int64) (bool, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var count int64
	// 在 lottery_wins 表中查找符合条件（用户ID匹配且中奖日期在今天之内）的记录数量
	err := db.Model(&model.LotteryWin{}).Where("user_id = ? AND win_date >= ? AND win_date < ?", userID, startOfDay, endOfDay).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RecordUserWin 记录用户的中奖信息
// 〔中文注释〕:【修正】将 gorm.DB() 替换为全局变量 db
func RecordUserWin(userID int64, prize string) error {
	winRecord := &model.LotteryWin{
		UserID:  userID,
		Prize:   prize,
		WinDate: time.Now(),
	}
	// 在 lottery_wins 表中创建一条新的记录
	return db.Create(winRecord).Error
}
