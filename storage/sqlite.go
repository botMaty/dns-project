package storage

import (
	"dns-server/types"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteStorage struct {
	db *gorm.DB
}

type DBRecord struct {
	ID        uint `gorm:"primarykey"`
	Name      string
	Type      uint16
	Value     string
	TTL       uint32
	ExpiresAt time.Time
}

func (DBRecord) TableName() string {
	return "records"
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&DBRecord{}); err != nil {
		return nil, err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_name_type ON records(name, type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_expires ON records(expires_at)")

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Get(q types.DNSQuestion) ([]types.DNSRecord, bool) {
	var dbRecs []DBRecord
	now := time.Now()

	result := s.db.Where("name = ? AND type = ? AND expires_at > ?",
		q.Name, uint16(q.Type), now).Find(&dbRecs)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, false
	}

	recs := make([]types.DNSRecord, len(dbRecs))
	for i, dbRec := range dbRecs {
		recs[i] = types.DNSRecord{
			Name:      dbRec.Name,
			Type:      types.RecordType(dbRec.Type),
			Value:     dbRec.Value,
			TTL:       dbRec.TTL,
			ExpiresAt: dbRec.ExpiresAt,
		}
	}

	return recs, true
}

func (s *SQLiteStorage) Set(r types.DNSRecord) {
	r.ExpiresAt = time.Now().Add(time.Duration(r.TTL) * time.Second)
	dbRec := DBRecord{
		Name:      r.Name,
		Type:      uint16(r.Type),
		Value:     r.Value,
		TTL:       r.TTL,
		ExpiresAt: r.ExpiresAt,
	}

	var existing DBRecord
	result := s.db.Where("name = ? AND type = ? AND value = ?",
		r.Name, uint16(r.Type), r.Value).First(&existing)

	if result.Error == nil {
		existing.TTL = r.TTL
		existing.ExpiresAt = r.ExpiresAt
		s.db.Save(&existing)
	} else {
		s.db.Create(&dbRec)
	}
}

func (s *SQLiteStorage) Delete(name string, rtype types.RecordType, value string) {
	if value == "" {
		// delete all
		s.db.Where("name = ? AND type = ?", name, uint16(rtype)).Delete(&DBRecord{})
	} else {
		s.db.Where("name = ? AND type = ? AND value = ?",
			name, uint16(rtype), value).Delete(&DBRecord{})
	}
}

func (s *SQLiteStorage) List() []types.DNSRecord {
	var dbRecs []DBRecord
	now := time.Now()

	s.db.Where("expires_at > ?", now).Find(&dbRecs)

	recs := make([]types.DNSRecord, len(dbRecs))
	for i, dbRec := range dbRecs {
		recs[i] = types.DNSRecord{
			Name:      dbRec.Name,
			Type:      types.RecordType(dbRec.Type),
			Value:     dbRec.Value,
			TTL:       dbRec.TTL,
			ExpiresAt: dbRec.ExpiresAt,
		}
	}

	return recs
}

func (s *SQLiteStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *SQLiteStorage) CleanupExpired() error {
	return s.db.Where("expires_at <= ?", time.Now()).Delete(&DBRecord{}).Error
}
