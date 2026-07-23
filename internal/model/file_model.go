package model

type File struct {
	FilePath   string
	CipherData []byte
	MetaData   string
	FileName   string
	UserLogin  string
}
