package database

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Project{},
		&Agent{},
		&Pipeline{},
		&Trigger{},
		&DatabaseTarget{},
		&Run{},
		&JobRun{},
		&TaskLog{},
		&Backup{},
	)
}
