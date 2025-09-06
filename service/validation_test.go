package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pyqol/pyqol/domain"
)

func TestCloneService_ComputeSimilarity_Validation(t *testing.T) {
	service := NewCloneService()
	ctx := context.Background()

	t.Run("Empty fragments", func(t *testing.T) {
		_, err := service.ComputeSimilarity(ctx, "", "print('hello')")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragments cannot be empty")

		_, err = service.ComputeSimilarity(ctx, "print('hello')", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragments cannot be empty")
	})

	t.Run("Nil context", func(t *testing.T) {
		//nolint:staticcheck // Testing nil context behavior
		_, err := service.ComputeSimilarity(nil, "print('hello')", "print('world')")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("Large fragments", func(t *testing.T) {
		largeFragment := make([]byte, 2*1024*1024) // 2MB
		for i := range largeFragment {
			largeFragment[i] = 'a'
		}

		_, err := service.ComputeSimilarity(ctx, string(largeFragment), "print('hello')")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragment size exceeds maximum allowed size")
	})

	t.Run("Valid fragments", func(t *testing.T) {
		fragment1 := "print('hello world')"
		fragment2 := "print('hello world')"

		similarity, err := service.ComputeSimilarity(ctx, fragment1, fragment2)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, similarity, 0.0)
		assert.LessOrEqual(t, similarity, 1.0)
	})
}

func TestCloneService_DetectClones_Validation(t *testing.T) {
	service := NewCloneService()
	ctx := context.Background()

	t.Run("Nil context", func(t *testing.T) {
		req := &domain.CloneRequest{}
		//nolint:staticcheck // Testing nil context behavior
		_, err := service.DetectClones(nil, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("Nil request", func(t *testing.T) {
		_, err := service.DetectClones(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clone request cannot be nil")
	})
}

func TestCloneService_DetectClonesInFiles_Validation(t *testing.T) {
	service := NewCloneService()
	ctx := context.Background()

	t.Run("Nil context", func(t *testing.T) {
		req := &domain.CloneRequest{}
		//nolint:staticcheck // Testing nil context behavior
		_, err := service.DetectClonesInFiles(nil, []string{"test.py"}, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("Nil request", func(t *testing.T) {
		_, err := service.DetectClonesInFiles(ctx, []string{"test.py"}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clone request cannot be nil")
	})

	t.Run("Empty file paths", func(t *testing.T) {
		req := &domain.CloneRequest{}
		_, err := service.DetectClonesInFiles(ctx, []string{}, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file paths cannot be empty")
	})
}
