#ifndef TEST_C_H
#define TEST_C_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>
#include <stdint.h>

// Opaque handle types
typedef void* MessageHandle;
typedef void* PluginHandle;
typedef void* ParameterHandle;

// Decode function
MessageHandle message_decode(const uint8_t* data, size_t len, char** error_msg);

// Encode function
size_t message_encode(MessageHandle handle, uint8_t** out_data, char** error_msg);

// Memory management functions
void message_free(MessageHandle handle);
void message_free_data(uint8_t* data);
void message_free_error(char* error_msg);

// Getter functions
size_t message_get_count(MessageHandle handle);
PluginHandle message_get_at(MessageHandle handle, size_t index);

#ifdef __cplusplus
}
#endif

#endif // TEST_C_H
