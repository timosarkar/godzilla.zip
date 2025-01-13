#include <iostream>
#include <fstream>
#include <zlib.h>

bool createAndCompressFile(const std::string &filename, const std::string &zipFilename, size_t size) {
    // Step 1: Create a 1MB file with zeros
    std::ofstream file(filename, std::ios::binary);
    if (!file) {
        std::cerr << "Failed to create the file!" << std::endl;
        return false;
    }

    // Write 1MB of zeros to the file
    for (size_t i = 0; i < size; ++i) {
        file.put(0);  // Write a single byte with value 0
    }
    file.close();

    std::ifstream input(filename, std::ios::binary);
    if (!input.is_open()) {
        std::cerr << "Error opening input file!" << std::endl;
        return false;
    }

    std::ofstream output(zipFilename, std::ios::binary);
    if (!output.is_open()) {
        std::cerr << "Error opening output file!" << std::endl;
        return false;
    }

    const int bufferSize = 128 * 1024;  // Buffer size for compression
    char buffer[bufferSize];
    char outBuffer[bufferSize];
    z_stream strm = {0};

    // Initialize the compression stream
    if (deflateInit(&strm, Z_DEFAULT_COMPRESSION) != Z_OK) {
        std::cerr << "Error initializing zlib compression stream!" << std::endl;
        return false;
    }

    // Read input file and compress it
    int flush;
    do {
        input.read(buffer, bufferSize);
        std::streamsize bytesRead = input.gcount();

        strm.avail_in = static_cast<uInt>(bytesRead);
        strm.next_in = reinterpret_cast<Bytef*>(buffer);

        do {
            strm.avail_out = bufferSize;
            strm.next_out = reinterpret_cast<Bytef*>(outBuffer);
            flush = (input.eof()) ? Z_FINISH : Z_NO_FLUSH;

            if (deflate(&strm, flush) == Z_STREAM_ERROR) {
                std::cerr << "Error during compression!" << std::endl;
                return false;
            }

            std::streamsize bytesWritten = bufferSize - strm.avail_out;
            output.write(outBuffer, bytesWritten);
        } while (strm.avail_out == 0);

    } while (!input.eof());

    deflateEnd(&strm);
    output.close();
    input.close();
    return true;
}

int main() {
    const std::string filename = "dummy";
    const std::string zipFilename = "main.zip";
    const size_t size = 1024 * 1024;  // 1MB in bytes

    if (createAndCompressFile(filename, zipFilename, size)) {
        std::cout << "1MB file with zeros created and compressed: " << zipFilename << std::endl;
    } else {
        std::cerr << "Error during file creation and compression." << std::endl;
    }

    return 0;
}
