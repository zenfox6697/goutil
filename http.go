package goutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func NewFileRecvHandler(sizeMB int64, dir string, formFile string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("File Upload Endpoint Hit")

		// Parse our multipart form, sizeMB << 20 specifies a maximum
		// upload of sizeMB MB files.
		r.ParseMultipartForm(sizeMB << 20)
		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, handler, err := r.FormFile(formFile)
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)

		// Create a temporary file within our temp-images directory that follows
		// a particular naming pattern
		// tempFile, err := ioutil.TempFile("temp-files", "upload-*.tmp")
		tempFile, err := ioutil.TempFile(dir, "*-"+handler.Filename)
		if err != nil {
			fmt.Println(err)
		}
		defer tempFile.Close()

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)
		// return that we have successfully uploaded our file!
		fmt.Fprintf(w, "Successfully Uploaded File\n")
	}
}

func NewFileServerHandler(dir, stripPrefix string) http.Handler {
	return http.StripPrefix(stripPrefix, http.FileServer(http.Dir(dir)))
}
