package cache

type item struct {
	ID    uint `gorm:"primarykey"`
	Key   string
	Value []byte
	Tags  []tag
}

type tag struct {
	ID     uint `gorm:"primarykey"`
	Name   string
	ItemID uint
}
