#include <iostream>
#include <fstream>
#include <vector>
#include "cpp/include/generated_c.h"

int main() {
    // Load test fixture
    std::ifstream file("experimental/cpp-bindings/common/complex.bin", std::ios::binary);
    if (!file) {
        std::cerr << "Failed to open test fixture\n";
        return 1;
    }
    
    std::vector<uint8_t> data((std::istreambuf_iterator<char>(file)),
                               std::istreambuf_iterator<char>());
    file.close();
    
    std::cout << "Loaded " << data.size() << " bytes\n";
    
    // Test decode
    char* error = nullptr;
    MessageHandle msg = message_decode(data.data(), data.size(), &error);
    
    if (!msg) {
        std::cerr << "Decode failed: " << (error ? error : "unknown") << "\n";
        if (error) message_free_error(error);
        return 1;
    }
    
    std::cout << "✓ Decode successful\n";
    
    // Test encode
    uint8_t* encoded_data = nullptr;
    size_t encoded_size = message_encode(msg, &encoded_data, &error);
    
    if (encoded_size == 0) {
        std::cerr << "Encode failed: " << (error ? error : "unknown") << "\n";
        if (error) message_free_error(error);
        message_free(msg);
        return 1;
    }
    
    std::cout << "✓ Encode successful: " << encoded_size << " bytes\n";
    
    // Verify size matches
    if (encoded_size == data.size()) {
        std::cout << "✓ Round-trip size matches!\n";
    } else {
        std::cerr << "✗ Size mismatch: " << encoded_size << " vs " << data.size() << "\n";
    }
    
    // Cleanup
    message_free_data(encoded_data);
    message_free(msg);
    
    return 0;
}
