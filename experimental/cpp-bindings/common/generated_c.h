#ifndef GENERATED_C_H
#define GENERATED_C_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque handle for Plugin struct
typedef void* PluginHandle;

// Create a new Plugin from binary data
// Returns NULL on error
PluginHandle plugin_decode(const uint8_t* data, size_t len, char** error_msg);

// Encode a Plugin to binary data
// Returns the size of the encoded data, or 0 on error
size_t plugin_encode(PluginHandle handle, uint8_t** out_data, char** error_msg);

// Free a Plugin handle
void plugin_free(PluginHandle handle);

// Free encoded data returned by plugin_encode
void plugin_free_data(uint8_t* data);

// Free error message string
void plugin_free_error(char* error_msg);

// Getters for Plugin fields
const char* plugin_get_name(PluginHandle handle);
const char* plugin_get_manufacturer_id(PluginHandle handle);
const char* plugin_get_type(PluginHandle handle);
const char* plugin_get_subtype(PluginHandle handle);
size_t plugin_get_parameters_count(PluginHandle handle);

// Opaque handle for Parameter struct
typedef void* ParameterHandle;

// Get a parameter by index (does not need to be freed, valid while plugin handle is valid)
ParameterHandle plugin_get_parameter(PluginHandle handle, size_t index);

// Getters for Parameter fields
const char* parameter_get_display_name(ParameterHandle handle);
float parameter_get_default_value(ParameterHandle handle);
float parameter_get_current_value(ParameterHandle handle);
int32_t parameter_get_address(ParameterHandle handle);
float parameter_get_max_value(ParameterHandle handle);
float parameter_get_min_value(ParameterHandle handle);
const char* parameter_get_unit(ParameterHandle handle);
const char* parameter_get_identifier(ParameterHandle handle);
int parameter_get_can_ramp(ParameterHandle handle);  // Returns 1 for true, 0 for false
int parameter_get_is_writable(ParameterHandle handle);
int64_t parameter_get_raw_flags(ParameterHandle handle);

// Optional array getter (returns count, or 0 if not present)
size_t parameter_get_indexed_values_count(ParameterHandle handle);
const char* parameter_get_indexed_value(ParameterHandle handle, size_t index);

// Optional string getter (returns NULL if not present)
const char* parameter_get_indexed_values_source(ParameterHandle handle);

#ifdef __cplusplus
}
#endif

#endif // GENERATED_C_H
