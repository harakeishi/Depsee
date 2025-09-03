package storage

// Reader はデータ読み取りインターフェース
type Reader interface {
	Read(key string) (string, error)
	ReadAll() (map[string]string, error)
}

// Writer はデータ書き込みインターフェース
type Writer interface {
	Write(key, value string) error
	Delete(key string) error
}

// Storage は読み書き両方のインターフェース
type Storage interface {
	Reader
	Writer
}

// Logger はログ出力インターフェース
type Logger interface {
	Log(message string)
	Error(message string)
}
