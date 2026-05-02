package task

import "time"

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	UserID    int       `json:"user_id"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
}

type UpdateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}
