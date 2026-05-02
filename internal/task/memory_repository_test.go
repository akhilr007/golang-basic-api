package task

import (
	"context"
	"testing"
	"time"

	"github.com/akhilr007/tasks/internal/utils"
)

func TestMemoryRepositoryGetAllPaginatesWithStableOrder(t *testing.T) {
	repo := NewMemoryRepository()
	createdAt := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	repo.tasks = map[int]Task{
		1: {ID: 1, UserID: 1, Title: "old", CreatedAt: createdAt},
		2: {ID: 2, UserID: 1, Title: "same time lower id", CreatedAt: createdAt.Add(time.Minute)},
		3: {ID: 3, UserID: 1, Title: "same time higher id", CreatedAt: createdAt.Add(time.Minute)},
		4: {ID: 4, UserID: 2, Title: "other user", CreatedAt: createdAt.Add(2 * time.Minute)},
	}

	tasks, hasMore, err := repo.GetAll(context.Background(), 1, utils.NewPagination(2, 0))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hasMore {
		t.Fatal("expected another page")
	}
	if got, want := taskIDs(tasks), []int{3, 2}; !sameInts(got, want) {
		t.Fatalf("expected first page IDs %v, got %v", want, got)
	}

	tasks, hasMore, err = repo.GetAll(context.Background(), 1, utils.NewPagination(2, 2))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hasMore {
		t.Fatal("expected final page")
	}
	if got, want := taskIDs(tasks), []int{1}; !sameInts(got, want) {
		t.Fatalf("expected second page IDs %v, got %v", want, got)
	}
}

func taskIDs(tasks []Task) []int {
	ids := make([]int, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
