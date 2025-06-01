package cache

type Item struct {
	Key   string `gorm:"primarykey"`
	Value []byte
}
