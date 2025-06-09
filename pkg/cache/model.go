package cache

type item struct {
	Key   string `gorm:"primarykey"`
	Value []byte
}
