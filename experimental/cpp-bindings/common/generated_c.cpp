#include "generated_c.h"
#include "generated.hpp"
#include <cstring>

// Internal wrapper structs
struct PluginHandleImpl {
    std::vector<test::Plugin> plugins;  // Store full vector, not just first plugin
    std::string error;
    std::vector<uint8_t> encoded_data;
};

struct ParameterHandleImpl {
    const test::Parameter* param;
};

// Helper to create error message
static char* make_error_msg(const std::string& msg) {
    char* error = new char[msg.size() + 1];
    std::strcpy(error, msg.c_str());
    return error;
}

extern "C" {

PluginHandle plugin_decode(const uint8_t* data, size_t len, char** error_msg) {
    if (!data || len == 0) {
        if (error_msg) *error_msg = make_error_msg("Invalid input data");
        return nullptr;
    }
    
    try {
        // Decode the message which returns a vector of Plugin
        auto plugins = test::decode_plugin_message(data, len);
        
        if (plugins.empty()) {
            if (error_msg) *error_msg = make_error_msg("No plugins in message");
            return nullptr;
        }
        
        PluginHandleImpl* handle = new PluginHandleImpl;
        handle->plugins = std::move(plugins);  // Store full vector
        return static_cast<PluginHandle>(handle);
    } catch (const std::exception& e) {
        if (error_msg) *error_msg = make_error_msg(e.what());
        return nullptr;
    }
}

size_t plugin_encode(PluginHandle handle, uint8_t** out_data, char** error_msg) {
    if (!handle) {
        if (error_msg) *error_msg = make_error_msg("Invalid handle");
        return 0;
    }
    
    try {
        PluginHandleImpl* impl = static_cast<PluginHandleImpl*>(handle);
        
        // Encode the full vector (not just first plugin)
        impl->encoded_data = test::encode_plugin_message(impl->plugins);
        
        // Allocate new buffer for caller
        *out_data = new uint8_t[impl->encoded_data.size()];
        std::memcpy(*out_data, impl->encoded_data.data(), impl->encoded_data.size());
        
        return impl->encoded_data.size();
    } catch (const std::exception& e) {
        if (error_msg) *error_msg = make_error_msg(e.what());
        return 0;
    }
}

void plugin_free(PluginHandle handle) {
    delete static_cast<PluginHandleImpl*>(handle);
}

void plugin_free_data(uint8_t* data) {
    delete[] data;
}

void plugin_free_error(char* error_msg) {
    delete[] error_msg;
}

const char* plugin_get_name(PluginHandle handle) {
    if (!handle) return nullptr;
    PluginHandleImpl* impl = static_cast<PluginHandleImpl*>(handle);
    return impl->plugins[0].Name.c_str();
}

const char* plugin_get_manufacturer_id(PluginHandle handle) {
    if (!handle) return nullptr;
    PluginHandleImpl* impl = static_cast<PluginHandleImpl*>(handle);
    return impl->plugins[0].ManufacturerID.c_str();
}

const char* plugin_get_type(PluginHandle handle) {
    if (!handle) return nullptr;
    PluginHandleImpl* impl = static_cast<PluginHandleImpl*>(handle);
    return impl->plugins[0].Type.c_str();
}

const char* plugin_get_subtype(PluginHandle handle) {
    if (!handle) return nullptr;
    PluginHandleImpl* impl = static_cast<PluginHandleImpl*>(handle);
    return impl->plugins[0].Subtype.c_str();
}

size_t plugin_get_parameters_count(PluginHandle handle) {
    if (!handle) return 0;
    PluginHandleImpl* impl = static_cast<PluginHandleImpl*>(handle);
    return impl->plugins[0].Parameters.size();
}

ParameterHandle plugin_get_parameter(PluginHandle handle, size_t index) {
    if (!handle) return nullptr;
    PluginHandleImpl* impl = static_cast<PluginHandleImpl*>(handle);
    if (index >= impl->plugins[0].Parameters.size()) {
        return nullptr;
    }
    
    ParameterHandleImpl* param_handle = new ParameterHandleImpl;
    param_handle->param = &impl->plugins[0].Parameters[index];
    return static_cast<ParameterHandle>(param_handle);
}

const char* parameter_get_display_name(ParameterHandle handle) {
    if (!handle) return nullptr;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->DisplayName.c_str();
}

float parameter_get_default_value(ParameterHandle handle) {
    if (!handle) return 0.0f;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->DefaultValue;
}

float parameter_get_current_value(ParameterHandle handle) {
    if (!handle) return 0.0f;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->CurrentValue;
}

int32_t parameter_get_address(ParameterHandle handle) {
    if (!handle) return 0;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->Address;
}

float parameter_get_max_value(ParameterHandle handle) {
    if (!handle) return 0.0f;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->MaxValue;
}

float parameter_get_min_value(ParameterHandle handle) {
    if (!handle) return 0.0f;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->MinValue;
}

const char* parameter_get_unit(ParameterHandle handle) {
    if (!handle) return nullptr;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->Unit.c_str();
}

const char* parameter_get_identifier(ParameterHandle handle) {
    if (!handle) return nullptr;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->Identifier.c_str();
}

int parameter_get_can_ramp(ParameterHandle handle) {
    if (!handle) return 0;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->CanRamp ? 1 : 0;
}

int parameter_get_is_writable(ParameterHandle handle) {
    if (!handle) return 0;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->IsWritable ? 1 : 0;
}

int64_t parameter_get_raw_flags(ParameterHandle handle) {
    if (!handle) return 0;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    return impl->param->RawFlags;
}

size_t parameter_get_indexed_values_count(ParameterHandle handle) {
    if (!handle) return 0;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    if (!impl->param->IndexedValues.has_value()) {
        return 0;
    }
    return impl->param->IndexedValues->size();
}

const char* parameter_get_indexed_value(ParameterHandle handle, size_t index) {
    if (!handle) return nullptr;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    if (!impl->param->IndexedValues.has_value()) {
        return nullptr;
    }
    if (index >= impl->param->IndexedValues->size()) {
        return nullptr;
    }
    return (*impl->param->IndexedValues)[index].c_str();
}

const char* parameter_get_indexed_values_source(ParameterHandle handle) {
    if (!handle) return nullptr;
    ParameterHandleImpl* impl = static_cast<ParameterHandleImpl*>(handle);
    if (!impl->param->IndexedValuesSource.has_value()) {
        return nullptr;
    }
    return impl->param->IndexedValuesSource->c_str();
}

} // extern "C"
