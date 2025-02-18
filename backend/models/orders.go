package models

type Order struct {
	ID          uint    `gorm:"primaryKey"`
	ProductID   uint    `json:"productID"`            
	Product     Product `gorm:"foreignKey:ProductID"` 
	Quantity    int     `json:"quantity"`             
	Status      string  `json:"status"`               
	Description string  `json:"description"`          
	InvoiceID   uint    `json:"invoiceID"`            
	Invoice     Invoice `gorm:"foreignKey:InvoiceID"` 
}
