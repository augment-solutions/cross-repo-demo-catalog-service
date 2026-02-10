package unit

import (
	"testing"

	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCategory_Creation(t *testing.T) {
	category := &models.Category{
		ID:       uuid.New(),
		Name:     "Electronics",
		Slug:     "electronics",
		IsActive: true,
	}

	assert.NotEmpty(t, category.ID)
	assert.Equal(t, "Electronics", category.Name)
	assert.Equal(t, "electronics", category.Slug)
	assert.True(t, category.IsActive)
}

func TestCategory_Hierarchy(t *testing.T) {
	parent := &models.Category{
		ID:    uuid.New(),
		Name:  "Electronics",
		Slug:  "electronics",
		Level: 0,
	}

	child := &models.Category{
		ID:       uuid.New(),
		Name:     "Smartphones",
		Slug:     "smartphones",
		ParentID: &parent.ID,
		Level:    1,
	}

	assert.Equal(t, 0, parent.Level)
	assert.Equal(t, 1, child.Level)
	assert.NotNil(t, child.ParentID)
	assert.Equal(t, parent.ID, *child.ParentID)
}

func TestCategory_Tree(t *testing.T) {
	root := &models.CategoryTree{
		ID:   uuid.New(),
		Name: "Root",
		Slug: "root",
	}

	child1 := &models.CategoryTree{
		ID:   uuid.New(),
		Name: "Child 1",
		Slug: "child-1",
	}

	child2 := &models.CategoryTree{
		ID:   uuid.New(),
		Name: "Child 2",
		Slug: "child-2",
	}

	root.Children = append(root.Children, child1, child2)

	assert.Len(t, root.Children, 2)
	assert.Equal(t, "Child 1", root.Children[0].Name)
	assert.Equal(t, "Child 2", root.Children[1].Name)
}

func TestCategory_ProductCount(t *testing.T) {
	category := &models.Category{
		ID:           uuid.New(),
		Name:         "Test Category",
		ProductCount: 0,
	}

	assert.Equal(t, 0, category.ProductCount)

	category.ProductCount = 10
	assert.Equal(t, 10, category.ProductCount)
}

func TestCategory_Path(t *testing.T) {
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Smartphones",
		Path: "electronics/mobile/smartphones",
	}

	assert.Equal(t, "electronics/mobile/smartphones", category.Path)
}

func TestCategory_Slug_Generation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "Electronics", "electronics"},
		{"with spaces", "Smart Phones", "smart-phones"},
		{"with special chars", "TV & Audio", "tv-audio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug := generateSlug(tt.input)
			assert.Equal(t, tt.expected, slug)
		})
	}
}

func TestCategory_Featured(t *testing.T) {
	category := &models.Category{
		ID:         uuid.New(),
		Name:       "Featured Category",
		IsFeatured: true,
	}

	assert.True(t, category.IsFeatured)
}

func TestCategory_SortOrder(t *testing.T) {
	categories := []*models.Category{
		{ID: uuid.New(), Name: "C", SortOrder: 3},
		{ID: uuid.New(), Name: "A", SortOrder: 1},
		{ID: uuid.New(), Name: "B", SortOrder: 2},
	}

	// Sort by sort order
	for i := 0; i < len(categories)-1; i++ {
		for j := i + 1; j < len(categories); j++ {
			if categories[i].SortOrder > categories[j].SortOrder {
				categories[i], categories[j] = categories[j], categories[i]
			}
		}
	}

	assert.Equal(t, "A", categories[0].Name)
	assert.Equal(t, "B", categories[1].Name)
	assert.Equal(t, "C", categories[2].Name)
}

func generateSlug(name string) string {
	slug := ""
	for _, c := range name {
		if c >= 'A' && c <= 'Z' {
			slug += string(c + 32)
		} else if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			slug += string(c)
		} else if c == ' ' {
			slug += "-"
		}
	}
	return slug
}

