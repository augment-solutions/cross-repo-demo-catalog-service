package db

// SQL queries for complex operations
const (
	// GetProductWithRelations gets a product with all relations
	GetProductWithRelations = `
		SELECT p.*, c.name as category_name, b.name as brand_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	// SearchProducts searches products with full-text search
	SearchProducts = `
		SELECT p.*
		FROM products p
		WHERE p.deleted_at IS NULL
		  AND (
			p.name ILIKE $1
			OR p.description ILIKE $1
			OR p.sku ILIKE $1
		  )
		ORDER BY
			CASE WHEN p.name ILIKE $1 THEN 0
				 WHEN p.sku ILIKE $1 THEN 1
				 ELSE 2
			END,
			p.name ASC
		LIMIT $2 OFFSET $3
	`

	// GetCategoryTree gets the full category tree
	GetCategoryTree = `
		WITH RECURSIVE category_tree AS (
			SELECT id, name, slug, parent_id, level, path, 0 as depth
			FROM categories
			WHERE parent_id IS NULL AND deleted_at IS NULL
			
			UNION ALL
			
			SELECT c.id, c.name, c.slug, c.parent_id, c.level, c.path, ct.depth + 1
			FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
			WHERE c.deleted_at IS NULL
		)
		SELECT * FROM category_tree ORDER BY depth, sort_order
	`

	// GetProductsByCategory gets products by category including subcategories
	GetProductsByCategory = `
		WITH RECURSIVE category_ids AS (
			SELECT id FROM categories WHERE id = $1
			UNION ALL
			SELECT c.id FROM categories c
			INNER JOIN category_ids ci ON c.parent_id = ci.id
		)
		SELECT p.*
		FROM products p
		WHERE p.category_id IN (SELECT id FROM category_ids)
		  AND p.deleted_at IS NULL
		  AND p.status = 'active'
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	// GetFeaturedProducts gets featured products
	GetFeaturedProducts = `
		SELECT p.*, 
			   (SELECT url FROM product_media WHERE product_id = p.id AND is_primary = true LIMIT 1) as image_url
		FROM products p
		WHERE p.is_featured = true
		  AND p.status = 'active'
		  AND p.deleted_at IS NULL
		ORDER BY p.updated_at DESC
		LIMIT $1
	`

	// GetNewArrivals gets recently added products
	GetNewArrivals = `
		SELECT p.*
		FROM products p
		WHERE p.is_new = true
		  AND p.status = 'active'
		  AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $1
	`

	// UpdateCategoryProductCount updates product count for a category
	UpdateCategoryProductCount = `
		UPDATE categories
		SET product_count = (
			SELECT COUNT(*) FROM products
			WHERE category_id = $1 AND deleted_at IS NULL AND status = 'active'
		)
		WHERE id = $1
	`

	// GetBrandProductCounts gets product counts for all brands
	GetBrandProductCounts = `
		SELECT b.id, b.name, COUNT(p.id) as product_count
		FROM brands b
		LEFT JOIN products p ON p.brand_id = b.id AND p.deleted_at IS NULL AND p.status = 'active'
		WHERE b.deleted_at IS NULL
		GROUP BY b.id, b.name
		ORDER BY product_count DESC
	`
)

