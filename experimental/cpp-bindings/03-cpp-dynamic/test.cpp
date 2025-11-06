#include "../common/generated_c.h"
#include <chrono>
#include <fstream>
#include <iostream>
#include <vector>
#include <cstdint>

std::vector<uint8_t> load_fixture(const char* path) {
    std::ifstream file(path, std::ios::binary | std::ios::ate);
    if (!file) {
        throw std::runtime_error("Cannot open fixture file");
    }
    size_t size = file.tellg();
    file.seekg(0);
    std::vector<uint8_t> buffer(size);
    file.read(reinterpret_cast<char*>(buffer.data()), size);
    return buffer;
}

int main() {
    try {
        auto data = load_fixture("../common/complex.bin");
        const int ITERATIONS = 100;
        
        // Warmup
        for (int i = 0; i < 10; i++) {
            char* error = nullptr;
            PluginHandle plugin = plugin_decode(data.data(), data.size(), &error);
            if (!plugin) {
                std::cerr << "Decode error: " << (error ? error : "unknown") << "\n";
                if (error) plugin_free_error(error);
                return 1;
            }
            plugin_free(plugin);
        }
        
        // Decode benchmark
        auto decode_start = std::chrono::high_resolution_clock::now();
        for (int i = 0; i < ITERATIONS; i++) {
            char* error = nullptr;
            PluginHandle plugin = plugin_decode(data.data(), data.size(), &error);
            plugin_free(plugin);
        }
        auto decode_end = std::chrono::high_resolution_clock::now();
        auto decode_us = std::chrono::duration_cast<std::chrono::microseconds>(
            decode_end - decode_start
        ).count() / ITERATIONS;
        
        // Get a plugin for encoding
        char* error = nullptr;
        PluginHandle plugin = plugin_decode(data.data(), data.size(), &error);
        
        // Encode benchmark
        auto encode_start = std::chrono::high_resolution_clock::now();
        uint8_t* encoded_data = nullptr;
        size_t encoded_size = 0;
        for (int i = 0; i < ITERATIONS; i++) {
            if (encoded_data) plugin_free_data(encoded_data);
            encoded_size = plugin_encode(plugin, &encoded_data, &error);
        }
        auto encode_end = std::chrono::high_resolution_clock::now();
        auto encode_us = std::chrono::duration_cast<std::chrono::microseconds>(
            encode_end - encode_start
        ).count() / ITERATIONS;
        
        // Output JSON for parsing
        std::cout << "{"
                  << "\"decode_us\":" << decode_us << ","
                  << "\"encode_us\":" << encode_us << ","
                  << "\"size_bytes\":" << encoded_size << ","
                  << "\"iterations\":" << ITERATIONS
                  << "}\n";
        
        plugin_free_data(encoded_data);
        plugin_free(plugin);
        
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << "\n";
        return 1;
    }
}
