package users

import "time"

type User struct {
	Id        int64     `db:"id,omitempty"`
	UserId    string    `db:"-"`
	Mobile    string    `db:"mobile,uni"`
	Username  string    `db:"username"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`
}
