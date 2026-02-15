package testdata

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string
	IsActive  bool
	CreatedAt string // simplified type for testing
}
