#include "generated_c.h"
#include "generated.hpp"
#include <cstring>

// Internal wrapper structs
struct MessageHandleImpl {
    std::vector<test::Plugin> items;  // Store full vector
    std::string error;
    std::vector<uint8_t> encoded_data;
};

// Helper to create error message
static char* make_error_msg(const std::string& msg) {
    char* error = new char[msg.size() + 1];
    std::strcpy(error, msg.c_str());
    return error;
}

extern "C" {

MessageHandle message_decode(const uint8_t* data, size_t len, char** error_msg) {
    if (!data || len == 0) {
        if (error_msg) *error_msg = make_error_msg("Invalid input data");
        return nullptr;
    }
    
    try {
        auto result = test::decode_plugin_message(data, len);
        
        if (result.empty()) {
            if (error_msg) *error_msg = make_error_msg("No items in message");
            return nullptr;
        }
        
        MessageHandleImpl* handle = new MessageHandleImpl;
        handle->items = std::move(result);
        return static_cast<MessageHandle>(handle);
    } catch (const std::exception& e) {
        if (error_msg) *error_msg = make_error_msg(e.what());
        return nullptr;
    }
}

size_t message_encode(MessageHandle handle, uint8_t** out_data, char** error_msg) {
    if (!handle) {
        if (error_msg) *error_msg = make_error_msg("Invalid handle");
        return 0;
    }
    
    try {
        MessageHandleImpl* impl = static_cast<MessageHandleImpl*>(handle);
        
        impl->encoded_data = test::encode_plugin_message(impl->items);
        
        // Allocate new buffer for caller
        *out_data = new uint8_t[impl->encoded_data.size()];
        std::memcpy(*out_data, impl->encoded_data.data(), impl->encoded_data.size());
        
        return impl->encoded_data.size();
    } catch (const std::exception& e) {
        if (error_msg) *error_msg = make_error_msg(e.what());
        return 0;
    }
}

void message_free(MessageHandle handle) {
    delete static_cast<MessageHandleImpl*>(handle);
}

void message_free_data(uint8_t* data) {
    delete[] data;
}

void message_free_error(char* error_msg) {
    delete[] error_msg;
}

// TODO: Implement message_get_name
// TODO: Implement message_get_manufacturerid
// TODO: Implement message_get_type
// TODO: Implement message_get_subtype
// TODO: Implement message_get_parameters

} // extern "C"
