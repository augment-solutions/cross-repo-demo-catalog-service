package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ecomm-stack/catalog-service/internal/config"
	"github.com/ecomm-stack/catalog-service/internal/domain/models"
	"github.com/ecomm-stack/catalog-service/internal/infrastructure/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedProduct represents a product extracted from the Kaggle dataset
type SeedProduct struct {
	StockCode   string
	Description string
	UnitPrice   float64
	Country     string
}

func main() {
	csvPath := flag.String("csv", "", "Path to the Kaggle e-commerce CSV file")
	jsonPath := flag.String("json", "", "Path to a pre-processed JSON file")
	limit := flag.Int("limit", 0, "Limit number of products to seed (0 = all)")
	dryRun := flag.Bool("dry-run", false, "Parse data without inserting into database")
	flag.Parse()

	if *csvPath == "" && *jsonPath == "" {
		log.Fatal("Either -csv or -json flag is required")
	}

	var products []*models.Product
	var err error

	if *csvPath != "" {
		products, err = parseKaggleCSV(*csvPath, *limit)
	} else {
		products, err = loadFromJSON(*jsonPath, *limit)
	}

	if err != nil {
		log.Fatalf("Failed to parse data: %v", err)
	}

	log.Printf("Parsed %d unique products", len(products))

	if *dryRun {
		log.Println("Dry run mode - not inserting into database")
		printSampleProducts(products, 5)
		return
	}

	cfg := config.Load()
	database, err := db.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := seedProducts(context.Background(), database, products); err != nil {
		log.Fatalf("Failed to seed products: %v", err)
	}

	log.Printf("Successfully seeded %d products", len(products))
}

func parseKaggleCSV(path string, limit int) ([]*models.Product, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colIndex := mapColumns(header)
	productMap := make(map[string]*SeedProduct)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Warning: skipping malformed row: %v", err)
			continue
		}

		stockCode := getField(record, colIndex["StockCode"])
		description := getField(record, colIndex["Description"])
		priceStr := getField(record, colIndex["UnitPrice"])
		country := getField(record, colIndex["Country"])

		// Skip invalid entries
		if stockCode == "" || description == "" || priceStr == "" {
			continue
		}
		// Skip cancelled orders (StockCode starting with 'C')
		if strings.HasPrefix(strings.ToUpper(stockCode), "C") {
			continue
		}
		// Skip non-product codes (like POST, DOT, M, etc.)
		if isNonProductCode(stockCode) {
			continue
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || price <= 0 {
			continue
		}

		// Use StockCode as unique key, keep first occurrence
		if _, exists := productMap[stockCode]; !exists {
			productMap[stockCode] = &SeedProduct{
				StockCode:   stockCode,
				Description: cleanDescription(description),
				UnitPrice:   price,
				Country:     country,
			}
		}
	}

	return convertToProducts(productMap, limit), nil
}

func mapColumns(header []string) map[string]int {
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[strings.TrimSpace(col)] = i
	}
	return colIndex
}

func getField(record []string, index int) string {
	if index >= 0 && index < len(record) {
		return strings.TrimSpace(record[index])
	}
	return ""
}

func isNonProductCode(code string) bool {
	nonProductCodes := []string{"POST", "DOT", "M", "BANK CHARGES", "PADS", "CRUK", "D", "S", "AMAZONFEE"}
	upperCode := strings.ToUpper(code)
	for _, npc := range nonProductCodes {
		if upperCode == npc {
			return true
		}
	}
	return false
}

func cleanDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	// Title case the description
	words := strings.Fields(strings.ToLower(desc))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	// Replace non-alphanumeric with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

func convertToProducts(productMap map[string]*SeedProduct, limit int) []*models.Product {
	products := make([]*models.Product, 0, len(productMap))
	count := 0

	// Placeholder images for products (using picsum.photos for variety)
	placeholderImages := []string{
		"https://picsum.photos/seed/%s/400/400",
	}

	for _, sp := range productMap {
		if limit > 0 && count >= limit {
			break
		}

		productID := uuid.New()
		imageURL := fmt.Sprintf(placeholderImages[0], sp.StockCode)

		product := &models.Product{
			ID:               productID,
			SKU:              sp.StockCode,
			Name:             sp.Description,
			Slug:             generateSlug(sp.Description) + "-" + strings.ToLower(sp.StockCode),
			Description:      sp.Description,
			ShortDescription: truncateString(sp.Description, 100),
			Status:           models.ProductStatusActive,
			IsFeatured:       count < 10, // First 10 products are featured
			IsNew:            true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Variants: []models.Variant{
				{
					ID:        uuid.New(),
					SKU:       sp.StockCode + "-DEFAULT",
					Name:      "Default",
					Price:     sp.UnitPrice,
					IsDefault: true,
					IsActive:  true,
					ImageURL:  imageURL,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Media: []models.Media{
				{
					ID:        uuid.New(),
					ProductID: productID,
					URL:       imageURL,
					Type:      models.MediaTypeImage,
					AltText:   sp.Description,
					IsPrimary: true,
					SortOrder: 0,
					CreatedAt: time.Now(),
				},
			},
		}
		products = append(products, product)
		count++
	}

	return products
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func loadFromJSON(path string, limit int) ([]*models.Product, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer file.Close()

	var products []*models.Product
	if err := json.NewDecoder(file).Decode(&products); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	if limit > 0 && len(products) > limit {
		products = products[:limit]
	}

	return products, nil
}

func printSampleProducts(products []*models.Product, count int) {
	if len(products) < count {
		count = len(products)
	}
	log.Printf("Sample of %d products:", count)
	for i := 0; i < count; i++ {
		p := products[i]
		price := 0.0
		if len(p.Variants) > 0 {
			price = p.Variants[0].Price
		}
		log.Printf("  [%d] SKU: %s, Name: %s, Price: %.2f", i+1, p.SKU, p.Name, price)
	}
}

func seedProducts(ctx context.Context, database *gorm.DB, products []*models.Product) error {
	log.Printf("Starting to seed %d products...", len(products))

	// Create categories
	categories, err := ensureCategories(ctx, database)
	if err != nil {
		return fmt.Errorf("failed to create categories: %w", err)
	}
	log.Printf("Created/found %d categories", len(categories))

	// Create products in batches
	batchSize := 100
	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}
		batch := products[i:end]

		if err := database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			for _, product := range batch {
				// Check if product already exists
				var existing models.Product
				if err := tx.Where("sku = ?", product.SKU).First(&existing).Error; err == nil {
					// Product exists, skip
					continue
				}

				// Assign category based on product name
				category := assignCategory(product.Name, categories)
				product.CategoryID = &category.ID

				// Set variant product IDs
				for j := range product.Variants {
					product.Variants[j].ProductID = product.ID
				}

				if err := tx.Create(product).Error; err != nil {
					log.Printf("Warning: failed to create product %s: %v", product.SKU, err)
					continue
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to seed batch %d-%d: %w", i, end, err)
		}

		log.Printf("Seeded products %d-%d of %d", i+1, end, len(products))
	}

	return nil
}

// CategoryDef defines a category with keywords for matching
type CategoryDef struct {
	Name        string
	Slug        string
	Description string
	Keywords    []string
	ImageURL    string
}

var categoryDefs = []CategoryDef{
	{
		Name:        "Home & Kitchen",
		Slug:        "home-kitchen",
		Description: "Quality kitchen essentials and home accessories for everyday use.",
		Keywords:    []string{"tea towel", "coat rack", "kitchen", "recipe box", "teaspoons", "mug cosy", "hot water bottle", "hottie"},
		ImageURL:    "https://images.unsplash.com/photo-1556909114-f6e7ad7d3136?w=600&h=400&fit=crop",
	},
	{
		Name:        "Toys & Games",
		Slug:        "toys-games",
		Description: "Fun toys, games, and puzzles for all ages.",
		Keywords:    []string{"puzzle", "jigsaw", "playhouse", "building block", "card game", "doll", "nesting boxes", "babushka"},
		ImageURL:    "https://images.unsplash.com/photo-1558060370-d644479cb6f7?w=600&h=400&fit=crop",
	},
	{
		Name:        "Stationery & Craft",
		Slug:        "stationery-craft",
		Description: "Creative stationery, craft supplies, and gift wrapping essentials.",
		Keywords:    []string{"sticker", "tape", "paint set", "bag", "globe"},
		ImageURL:    "https://images.unsplash.com/photo-1513542789411-b6a5d4f31634?w=600&h=400&fit=crop",
	},
	{
		Name:        "Home Decor",
		Slug:        "home-decor",
		Description: "Beautiful decorative items to brighten up your home.",
		Keywords:    []string{"lantern", "t-light", "ornament", "knob", "night light", "clock", "heart", "hanger"},
		ImageURL:    "https://images.unsplash.com/photo-1586023492125-27b2c045efd7?w=600&h=400&fit=crop",
	},
	{
		Name:        "Lunch & Storage",
		Slug:        "lunch-storage",
		Description: "Practical lunch boxes and storage solutions for home and school.",
		Keywords:    []string{"lunch box", "snack box", "jam making"},
		ImageURL:    "https://images.unsplash.com/photo-1604467794349-0b74285de7e7?w=600&h=400&fit=crop",
	},
	{
		Name:        "Hand Warmers",
		Slug:        "hand-warmers",
		Description: "Cozy hand warmers to keep you warm in cold weather.",
		Keywords:    []string{"hand warmer"},
		ImageURL:    "https://images.unsplash.com/photo-1544027993-37dbfe43562a?w=600&h=400&fit=crop",
	},
}

func ensureCategories(ctx context.Context, database *gorm.DB) (map[string]*models.Category, error) {
	categories := make(map[string]*models.Category)

	for _, def := range categoryDefs {
		var category models.Category
		err := database.WithContext(ctx).Where("slug = ?", def.Slug).First(&category).Error
		if err == nil {
			categories[def.Slug] = &category
			continue
		}

		// Create category
		category = models.Category{
			ID:          uuid.New(),
			Name:        def.Name,
			Slug:        def.Slug,
			Description: def.Description,
			ImageURL:    def.ImageURL,
			IsActive:    true,
			IsFeatured:  true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := database.WithContext(ctx).Create(&category).Error; err != nil {
			return nil, err
		}
		categories[def.Slug] = &category
		log.Printf("Created category: %s", def.Name)
	}

	return categories, nil
}

func assignCategory(productName string, categories map[string]*models.Category) *models.Category {
	lowerName := strings.ToLower(productName)

	for _, def := range categoryDefs {
		for _, keyword := range def.Keywords {
			if strings.Contains(lowerName, keyword) {
				return categories[def.Slug]
			}
		}
	}

	// Default to Home Decor if no match
	if cat, ok := categories["home-decor"]; ok {
		return cat
	}

	// Fallback to first category
	for _, cat := range categories {
		return cat
	}

	return nil
}

