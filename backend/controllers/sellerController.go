package controllers

import (
	"awesomeProject3/database"
	"awesomeProject3/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProductsForReceiving(c *gin.Context) {
	var products []models.Product
	if err := database.DB.Where("is_accepted = ?", "Принят Кладовщиком").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список товаров для приемки"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func ConfirmProductReception(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingProduct models.Product
	if err := database.DB.Where("id = ? AND is_accepted = ?", product.ID, "Принят Кладовщиком").First(&existingProduct).Error; err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден или уже принят продавцом"})
		return
	}

	if err := database.DB.Model(&product).Where("id = ?", product.ID).Update("is_accepted", "Принят Продавцом").Update("storage_location", "Магазин").Error; err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось подтвердить приемку товара"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар успешно принят"})
}

func CheckProductStock(c *gin.Context) {
	var products []models.Product
	if err := database.DB.Where("quantity < ?", 10).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось проверить наличие товаров"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func GetProductsForSale(c *gin.Context) {
	var products []models.Product
	if err := database.DB.Where("quantity > ?", 0).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список товаров для продажи"})
		return
	}
	c.JSON(http.StatusOK, products)
}

func ProcessSale(c *gin.Context) {
	var saleRequest models.Product
	if err := c.ShouldBindJSON(&saleRequest); err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var product models.Product
	if err := database.DB.Where("id = ? AND storage_location = ?", saleRequest.ID, "Магазин").First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		return
	}

	if product.Quantity < saleRequest.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недостаточно товара на складе"})
		return
	}

	product.Quantity -= saleRequest.Quantity
	if err := database.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при оформлении продажи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Продажа оформлена", "product": product})
}

func InspectProductQuality(c *gin.Context) {
	var product models.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingProduct models.Product
	if err := database.DB.First(&existingProduct, product.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		return
	}

	if product.Status == "дефектный" {
		existingProduct.IsAccepted = "Не принят"
		existingProduct.StorageLocation = "Поставщик"
	} else {
		existingProduct.IsAccepted = "Принят Продавцом"
		existingProduct.StorageLocation = "Магазин"
	}

	existingProduct.Status = product.Status

	if err := database.DB.Save(&existingProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при проверке качества товара"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Проверка качества завершена", "product": existingProduct})
}

func NotifyProductShortage(c *gin.Context) {
	var products []models.Product
	if err := database.DB.Where("quantity < ?", 10).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при проверке товаров на складе"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Уведомление о недостатке товаров", "products": products})
}

func AddProductToInvoice(c *gin.Context) { 
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		fmt.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Create(&product).Error; err != nil {
		fmt.Println("Error creating product:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении товара"})
		return
	}

	invoiceID := c.DefaultQuery("invoice_id", "")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID накладной"})
		return
	}

	var invoice models.Invoice
	if err := database.DB.First(&invoice, "id = ?", invoiceID).Error; err != nil {
		fmt.Println("Error finding invoice:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске накладной"})
		return
	}

	if err := database.DB.Model(&invoice).Association("Products").Append(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении товара в накладную"})
		return
	}

	invoice.Total += product.Price * float64(product.Quantity)

	if err := database.DB.Save(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении накладной"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Товар добавлен в накладную",
		"product": product,
		"invoice": invoice,
	})
}

func RemoveProductsFromInvoice(c *gin.Context) {
	var req struct {
		InvoiceID  uint   `json:"invoice_id"`
		ProductIDs []uint `json:"product_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	var invoice models.Invoice
	if err := database.DB.Preload("Products").First(&invoice, req.InvoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Накладная не найдена"})
		return
	}

	if invoice.Total == 0 {
		if err := database.DB.Where("invoice_id = ?", req.InvoiceID).Delete(&models.InvoiceProduct{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении всех товаров из накладной"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Все товары успешно удалены, так как общая стоимость равна 0",
		})
		return
	}

	for _, productID := range req.ProductIDs {
		var invoiceProduct models.InvoiceProduct

		if err := database.DB.Where("invoice_id = ? AND product_id = ?", req.InvoiceID, productID).First(&invoiceProduct).Error; err == nil {
			if invoiceProduct.Quantity > 1 {
				invoiceProduct.Quantity--
				if err := database.DB.Save(&invoiceProduct).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении количества товара"})
					return
				}
			} else {
				if err := database.DB.Delete(&invoiceProduct).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении товара"})
					return
				}
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Товар с ID %d не найден в накладной", productID)})
			return
		}
	}

	updateInvoiceTotal(&invoice)

	c.JSON(http.StatusOK, gin.H{
		"message": "Товары успешно удалены",
		"invoice": invoice,
	})
}

func updateInvoiceTotal(invoice *models.Invoice) {
	var products []models.InvoiceProduct
	database.DB.Where("invoice_id = ?", invoice.ID).Find(&products)

	total := 0.0
	for _, product := range products {
		var p models.Product
		if err := database.DB.First(&p, product.ProductID).Error; err == nil {
			total += float64(product.Quantity) * p.Price
		}
	}

	invoice.Total = total
	database.DB.Save(invoice)
}
func DeleteProductFromInvoice(c *gin.Context) {
	invoiceID := c.Param("invoice_id")
	productID := c.Param("product_id")
	var invoice models.Invoice
	if err := database.DB.Preload("Products").First(&invoice, invoiceID).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"message": "Invoice not found"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	if err := database.DB.Where("invoice_id = ? AND product_id = ?", invoiceID, productID).
		Delete(&models.InvoiceProduct{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting product from invoice"})
		return
	}
	if err := database.DB.Where("id = ?", productID).
		Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting product from invoice"})
		return
	}
}
