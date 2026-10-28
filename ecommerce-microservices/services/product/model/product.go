package model

type Product struct {
	ID    int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name  string  `gorm:"size:100;not null" json:"name" binding:"required"`
	Price float64 `gorm:"type:numeric(10,2);not null" json:"price" binding:"required,gt=0"`
}

func (Product) TableName() string {
	return "products"
}
