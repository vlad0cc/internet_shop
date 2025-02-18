package models

type Invoice struct {
	ID       uint      `json:"InvoiceId" gorm:"primaryKey"`
	UserID   uint      `json:"user_id"`                                                
	Products []Product `gorm:"many2many:invoice_products;constraint:OnDelete:CASCADE"` 
	Total    float64   `json:"total"`                                                  
}
