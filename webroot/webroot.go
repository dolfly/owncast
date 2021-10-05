package webroot

import (
	"embed"
	"io/ioutil"
	"log"
	"os"
)

//go:embed *.html *.ico *.jpg *.json offline.ts js img styles
var FS embed.FS

func Copy(path string, destination string) error {
	input, err := FS.ReadFile(path)
	if err != nil {
		log.Println(err)
		return err
	}
	return ioutil.WriteFile(destination, input, 0600)
}

// DoesFileExists checks if the file exists.
func DoesFileExists(name string) bool {
	if fh, err := FS.Open(name); err == nil {
		fh.Close()
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return false
	}
}
