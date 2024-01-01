package document

import (
	"gorm.io/gorm"
	"os"
	"time"
)

func CleanUp(db *gorm.DB) {
	repo := NewRepositoryImp(db)
	documents, err := repo.GetExpired()
	if err != nil {
		panic(err)
	}
	for _, d := range documents {
		now := time.Now()
		d.Status = Expired
		d.UpdatedAt = &now
		err = repo.Update(&d)
		if err != nil {
			panic(err)
		}
		err = os.Remove(dataFolder + d.ID)
		if err != nil {
			panic(err)
		}
	}
}
