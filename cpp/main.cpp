#include <iostream>
#include <fstream>
#include <cstring>
#include <vector>
#include <lz4.h>
#include <zip.h>
#include <gmp.h>
#include <filesystem>
#include <chrono>

void computeDecompressedSize(const std::string& zipFileName, mpz_t& totalSize) {
    int error = 0;
    zip_t* zip = zip_open(zipFileName.c_str(), 0, &error);
    if (!zip) {
        std::cerr << "Error opening ZIP file: " << zip_strerror(zip) << std::endl;
        return;
    }

    zip_uint64_t numEntries = zip_get_num_entries(zip, 0);
    for (zip_uint64_t i = 0; i < numEntries; ++i) {
        struct zip_stat st;
        zip_stat_init(&st);
        if (zip_stat_index(zip, i, 0, &st) != 0) {
            std::cerr << "Error getting file stats: " << zip_strerror(zip) << std::endl;
            continue;
        }

        // Add the size of regular files directly to the total size
        if (st.name[strlen(st.name) - 1] != '/') {
            mpz_add_ui(totalSize, totalSize, st.size);
        }

        // If the file is another ZIP archive, we don't extract it, but treat it as metadata
        if (st.name[strlen(st.name) - 1] == 'z' && strstr(st.name, ".zip")) {
            std::cout << "Found nested ZIP file: " << st.name << std::endl;
            // Do not decompress, only consider its metadata
            mpz_add_ui(totalSize, totalSize, st.size); // Add its size as if it was fully expanded
        }
    }

    zip_close(zip);
}


// Function to create a text file
void createTextFile(const std::string& fileName, const std::string& content) {
    std::ofstream outFile(fileName);
    outFile << content;
    outFile.close();
}

// Function to compress data using LZ4
std::vector<char> compressLZ4(const std::string& input) {
    int maxCompressedSize = LZ4_compressBound(input.size());
    std::vector<char> compressedData(maxCompressedSize);

    int compressedSize = LZ4_compress_default(
        input.c_str(),
        compressedData.data(),
        input.size(),
        maxCompressedSize
    );

    compressedData.resize(compressedSize);
    return compressedData;
}

// Function to create a ZIP file and add compressed data
void createZipFile(const std::string& zipFileName, const std::string& fileName, const std::vector<char>& compressedData) {
    int error = 0;
    zip_t* zip = zip_open(zipFileName.c_str(), ZIP_CREATE | ZIP_TRUNCATE, &error);
    zip_source_t* source = zip_source_buffer(zip, compressedData.data(), compressedData.size(), 0);

    if (zip_file_add(zip, fileName.c_str(), source, ZIP_FL_OVERWRITE) < 0) {
        std::cerr << "Error adding file to ZIP: " << zip_strerror(zip) << std::endl;
        zip_source_free(source);
        zip_close(zip);
        return;
    }

    zip_close(zip);
}

// Function to compress a ZIP file
void compressZipFile(const std::string& inputZipFileName, const std::string& outputZipFileName) {
    // Read the previous zip file
    std::ifstream inFile(inputZipFileName, std::ios::binary);
    std::vector<char> zipContent((std::istreambuf_iterator<char>(inFile)), std::istreambuf_iterator<char>());
    inFile.close();

    // Create a new zip file
    zip_t* zip = zip_open(outputZipFileName.c_str(), ZIP_CREATE | ZIP_TRUNCATE, nullptr);
    zip_source_t* source1 = zip_source_buffer(zip, zipContent.data(), zipContent.size(), 0);
    zip_source_t* source2 = zip_source_buffer(zip, zipContent.data(), zipContent.size(), 0);

    // Add two identical copies to the new zip archive
    zip_file_add(zip, "copy1.zip", source1, ZIP_FL_OVERWRITE);
    zip_file_add(zip, "copy2.zip", source2, ZIP_FL_OVERWRITE);

    zip_close(zip);
}

void handleMultipleLevels(int numLevels) {
    mpz_t decompressionSize;
    mpz_init(decompressionSize);
    mpz_set_ui(decompressionSize, 1024); // Initial size for lvl1.zip

    std::string finalDecompressionSize;

    for (int level = 2; level <= numLevels; ++level) {
        std::string inputZipFileName = "lvl" + std::to_string(level - 1) + ".zip";
        std::string outputZipFileName = "lvl" + std::to_string(level) + ".zip";

        compressZipFile(inputZipFileName, outputZipFileName);

        // Increase the decompression size by a factor of 10
        mpz_mul_ui(decompressionSize, decompressionSize, 2);

        // Store the size of the final level
        if (level == numLevels) {
            char* size_str = mpz_get_str(nullptr, 10, decompressionSize);
            finalDecompressionSize = size_str;
            free(size_str);
        }

        std::remove(inputZipFileName.c_str());
    }

    std::string finalZipFileName = "lvl" + std::to_string(numLevels) + ".zip";
    std::rename(finalZipFileName.c_str(), "final.zip");

    // Output the decompression size for the final level
    std::cout << "Final decompression size: " << finalDecompressionSize << " bytes" << std::endl;

    mpz_clear(decompressionSize);
}


void printRealSize(const std::string& zipFileName) {
    int error = 0;
    zip_t* zip = zip_open(zipFileName.c_str(), 0, &error);
    if (!zip) {
        std::cerr << "Error opening ZIP file: " << zip_strerror(zip) << std::endl;
        return;
    }

    zip_uint64_t totalSize = 0;
    zip_uint64_t numEntries = zip_get_num_entries(zip, 0);
    for (zip_uint64_t i = 0; i < numEntries; ++i) {
        struct zip_stat st;
        zip_stat_init(&st);
        zip_stat(zip, zip_get_name(zip, i, 0), 0, &st);
        totalSize += st.size; // Add the size of each file
    }


    std::cout << "Real size of " << zipFileName << ": " << totalSize << " bytes" << std::endl;

    zip_close(zip);
}


int main(int argc, char* argv[]) {
    try {
        std::chrono::steady_clock::time_point begin = std::chrono::steady_clock::now();

        // Step 1: Create a text file
        const std::string textFileName = "init";
        const std::string textContent(1024 * 1024, '0');
        createTextFile(textFileName, textContent);

        // Step 2: Read file content
        std::ifstream inFile(textFileName, std::ios::binary);
        std::string fileContent((std::istreambuf_iterator<char>(inFile)), std::istreambuf_iterator<char>());
        inFile.close();

        // Step 3: Compress the file content using LZ4
        auto compressedData = compressLZ4(fileContent);

        // Step 4: Create a ZIP file and add the compressed data
        const std::string zipFileName = "lvl1.zip";
        createZipFile(zipFileName, textFileName, compressedData);
        std::remove("init");

        // Step 5: Handle multiple levels of compression
        int numLevels = std::atoi(argv[1]); // Specify the number of levels
        handleMultipleLevels(numLevels);

        printRealSize("final.zip");


        // Step 7: Calculate and print elapsed time in seconds
        std::chrono::steady_clock::time_point end = std::chrono::steady_clock::now();
        double elapsedSeconds = std::chrono::duration_cast<std::chrono::duration<double> >(end - begin).count();
        std::cout << "Time elapsed: " << elapsedSeconds << " seconds" << std::endl;

    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }

    return 0;
}
