package grepbook

type BookReview struct {
	UID   string `json:"uid"`
	Title string `json:"title"`
	HTML  []byte `json:"html"`
	Delta string `json:"delta"`
}

func (db *DB) CreateBookReview(uid, title, delta string, HTML []byte) (*BookReview, error) {
	return &BookReview{}, nil
}

func (db *DB) GetBookReview(uid string) (*BookReview, error) {
	return &BookReview{}, nil
}

func (db *DB) DeleteBookReview(uid string) error {
	return nil
}

type BookReviewDB interface {
	CreateBookReview(string, string, string, []byte) (*BookReview, error)
	GetBookReview(string) (*BookReview, error)
	DeleteBookReview(string) error
}

func (br *BookReview) Save(db *DB) error {
	return nil
}
