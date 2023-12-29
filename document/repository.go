package document

type Repository interface {
	FindById(id string) (Document, error)
	Save(document Document) (Document, error)
	Update(document Document) (Document, error)
	GetExpired() ([]Document, error)
	GetTotalUsage(client string) (int64, error)
}
