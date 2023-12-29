package document

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestRepositoryImp_FindById(t *testing.T) {
	tests := []struct {
		name string
		id   string
		// add other test case parameters as needed
	}{
		{"valid_id", "valid_id"},
		// add other test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer mockDb.Close()

			mock.ExpectQuery("SELECT (.+) FROM \"documents\" WHERE id = (.+) ORDER BY \"documents\".\"id\" LIMIT 1").
				WithArgs("valid_id").
				WillReturnRows(sqlmock.NewRows([]string{"id", "client", "status", "created_at", "updated_at", "uploaded_at"}).AddRow("valid_id", "valid_client", 1, "2021-01-01 00:00:00", nil, time.Time{})) // add rows expected to be returned

			dialector := postgres.New(postgres.Config{
				Conn:       mockDb,
				DriverName: "postgres",
			})
			db, _ := gorm.Open(dialector, &gorm.Config{})

			r := NewRepositoryImp(db)
			d, err := r.FindById(tt.id)
			if err != nil {
				t.Fatal("FindById() should not return error.")
			}

			// add assertion here
			fmt.Println(d)
		})
	}
}
