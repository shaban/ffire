#import <Foundation/Foundation.h>
#include "../common/generated_c.h"
#include <chrono>
#include <vector>

std::vector<uint8_t> loadFixture(NSString* path) {
    NSData* data = [NSData dataWithContentsOfFile:path];
    if (!data) {
        @throw [NSException exceptionWithName:@"FileError" 
                                       reason:@"Cannot load fixture"
                                     userInfo:nil];
    }
    return std::vector<uint8_t>((const uint8_t*)data.bytes, 
                                (const uint8_t*)data.bytes + data.length);
}

int main(int, const char **) {
    @autoreleasepool {
        try {
            auto data = loadFixture(@"../common/complex.bin");
            const int ITERATIONS = 100;
            
            // Warmup
            for (int i = 0; i < 10; i++) {
                char* error = nullptr;
                PluginHandle plugin = plugin_decode(data.data(), data.size(), &error);
                if (!plugin) {
                    NSLog(@"Decode error: %s", error ? error : "unknown");
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
            printf("{\"decode_us\":%lld,\"encode_us\":%lld,\"size_bytes\":%zu,\"iterations\":%d}\n",
                   decode_us, encode_us, encoded_size, ITERATIONS);
            
            plugin_free_data(encoded_data);
            plugin_free(plugin);
            
            return 0;
        } catch (const std::exception& e) {
            NSLog(@"Error: %s", e.what());
            return 1;
        }
    }
}
