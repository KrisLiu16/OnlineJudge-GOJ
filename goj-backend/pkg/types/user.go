package types

import "time"

type User struct {
	ID               uint      `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Avatar           string    `json:"avatar,omitempty"`
	Bio              string    `json:"bio,omitempty"`
	Role             string    `json:"role"`
	Submissions      int64     `json:"submissions"`
	AcceptedProblems int64     `json:"acceptedProblems"`
	Rating           int64     `json:"rating"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
