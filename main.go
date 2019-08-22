package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter The Folder Name [For eg. NGS_Run11]: ")
	folder, _ := reader.ReadString('\n')
	folder = strings.TrimSpace(folder)
	folderPath := "./" + folder + "/"
	copyFolderPath := "./Generated_vcfs/" + folder
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		panic(err)
	}

	fmt.Println("Looking for .vcf files in " + folder)
	err := filepath.Walk(folderPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// if the name contains genome, we'll skip it
			if strings.Contains(strings.ToLower(info.Name()), "genome") {
				return nil
			}
			// if the file is a .vcf.gz we'll unzip it in our copy folder
			if strings.HasSuffix(info.Name(), ".vcf.gz") {
				fmt.Println("Unzipping " + info.Name())
				compressedData, err := ioutil.ReadFile(path)
				check(err)
				uncompressedData, uncompressedDataErr := gUnzipData(compressedData)
				if uncompressedDataErr != nil {
					log.Fatal(uncompressedDataErr)
				}
				// let's create the copy folder if doesn't exist
				if _, err := os.Stat(copyFolderPath); os.IsNotExist(err) {
					os.Mkdir(copyFolderPath, 0777)
				}
				// removing the .gz from the end of the file
				fileName := info.Name()[:len(info.Name())-3]
				err = ioutil.WriteFile(copyFolderPath+"/"+fileName, uncompressedData, 0777)
				check(err)
				fmt.Println("Successful Unzipped " + info.Name())
			}
			// if the file is a .vcf let's copy it into our copy folder
			if filepath.Ext(path) == ".vcf" {
				fmt.Println("Found: " + info.Name())
				// if the folder where the files need to be copied does not exist
				// we will create a new folder with the same name
				if _, err := os.Stat(copyFolderPath); os.IsNotExist(err) {
					os.Mkdir(copyFolderPath, 0777)
				}
				// start copying files.
				err := CopyFile(path, copyFolderPath+"/"+info.Name())
				if err != nil {
					log.Println(err)
				}
			}
			return nil
		})
	if err != nil {
		panic(err)
	}

	fmt.Println("Done Copying the files.")

	// replaces the text
	err = filepath.Walk(copyFolderPath, visit)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done Replacing!")
}

// CopyFile copies files from src to dst
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func visit(path string, fi os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if !!fi.IsDir() {
		return nil //
	}

	matched, err := filepath.Match("*.vcf", fi.Name())

	if err != nil {
		panic(err)
		return err
	}

	if matched {
		read, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		//fmt.Println(string(read))
		fmt.Println("replacing: ", path)

		old := "##FORMAT=<ID=AD,Number=.,Type=Integer,Description=\"Allele Depth\">"
		new := "##FORMAT=<ID=AD,Number=A,Type=Integer,Description=\"Allelic depths for the ref and alt alleles in the order listed\"> "

		newContents := strings.Replace(string(read), old, new, -1)

		err = ioutil.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			panic(err)
		}

	}

	return nil
}

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
