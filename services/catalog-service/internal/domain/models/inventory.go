package models

import (
	"time"

	"github.com/google/uuid"
)

// InventoryStatus represents inventory status
type InventoryStatus string

const (
	InventoryStatusInStock     InventoryStatus = "in_stock"
	InventoryStatusLowStock    InventoryStatus = "low_stock"
	InventoryStatusOutOfStock  InventoryStatus = "out_of_stock"
	InventoryStatusBackorder   InventoryStatus = "backorder"
)

// InventoryInfo represents inventory information for a variant
type InventoryInfo struct {
	VariantID         uuid.UUID       `json:"variant_id"`
	SKU               string          `json:"sku"`
	Quantity          int             `json:"quantity"`
	ReservedQuantity  int             `json:"reserved_quantity"`
	AvailableQuantity int             `json:"available_quantity"`
	Status            InventoryStatus `json:"status"`
	WarehouseID       *uuid.UUID      `json:"warehouse_id,omitempty"`
	LowStockThreshold int             `json:"low_stock_threshold"`
	BackorderAllowed  bool            `json:"backorder_allowed"`
	LeadTimeDays      int             `json:"lead_time_days"`
	LastUpdated       time.Time       `json:"last_updated"`
}

// GetStatus calculates the inventory status
func (i *InventoryInfo) GetStatus() InventoryStatus {
	if i.AvailableQuantity <= 0 {
		if i.BackorderAllowed {
			return InventoryStatusBackorder
		}
		return InventoryStatusOutOfStock
	}
	if i.AvailableQuantity <= i.LowStockThreshold {
		return InventoryStatusLowStock
	}
	return InventoryStatusInStock
}

// InventorySync represents sync status with inventory service
type InventorySync struct {
	VariantID      uuid.UUID `json:"variant_id"`
	LastSyncedAt   time.Time `json:"last_synced_at"`
	SyncStatus     string    `json:"sync_status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// StockLevel represents a simplified stock level
type StockLevel struct {
	SKU       string          `json:"sku"`
	Quantity  int             `json:"quantity"`
	Status    InventoryStatus `json:"status"`
	UpdatedAt time.Time       `json:"updated_at"`
}

