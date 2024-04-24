package utils

import (
"os"
)

func TorrentFile(filePath string) string{
	fileData,err:=os.ReadFile(filePath)
	if err!=nil{
		panic(err)
	}
	return string(fileData)
}