#ifndef TEST_H
#define TEST_H

#include <cstdint>
#include <cstring>
#include <string>
#include <vector>
#include <optional>
#include <stdexcept>

namespace test {

struct Plugin;
struct Parameter;

struct Parameter {
    std::string DisplayName;
    float DefaultValue;
    float CurrentValue;
    int32_t Address;
    float MaxValue;
    float MinValue;
    std::string Unit;
    std::string Identifier;
    bool CanRamp;
    bool IsWritable;
    int64_t RawFlags;
    std::optional<std::vector<std::string>> IndexedValues;
    std::optional<std::string> IndexedValuesSource;
};

struct Plugin {
    std::string Name;
    std::string ManufacturerID;
    std::string Type;
    std::string Subtype;
    std::vector<Parameter> Parameters;
};

// Binary encoder for wire format
class Encoder {
public:
    std::vector<uint8_t> buffer;

    void write_byte(uint8_t b) { buffer.push_back(b); }

    void write_bool(bool v) { buffer.push_back(v ? 0x01 : 0x00); }

    void write_int8(int8_t v) { buffer.push_back(static_cast<uint8_t>(v)); }

    void write_int16(int16_t v) {
        uint16_t u = static_cast<uint16_t>(v);
        buffer.push_back(static_cast<uint8_t>(u));
        buffer.push_back(static_cast<uint8_t>(u >> 8));
    }

    void write_int32(int32_t v) {
        uint32_t u = static_cast<uint32_t>(v);
        buffer.push_back(static_cast<uint8_t>(u));
        buffer.push_back(static_cast<uint8_t>(u >> 8));
        buffer.push_back(static_cast<uint8_t>(u >> 16));
        buffer.push_back(static_cast<uint8_t>(u >> 24));
    }

    void write_int64(int64_t v) {
        uint64_t u = static_cast<uint64_t>(v);
        buffer.push_back(static_cast<uint8_t>(u));
        buffer.push_back(static_cast<uint8_t>(u >> 8));
        buffer.push_back(static_cast<uint8_t>(u >> 16));
        buffer.push_back(static_cast<uint8_t>(u >> 24));
        buffer.push_back(static_cast<uint8_t>(u >> 32));
        buffer.push_back(static_cast<uint8_t>(u >> 40));
        buffer.push_back(static_cast<uint8_t>(u >> 48));
        buffer.push_back(static_cast<uint8_t>(u >> 56));
    }

    void write_float32(float v) {
        uint32_t u;
        std::memcpy(&u, &v, sizeof(float));
        write_int32(static_cast<int32_t>(u));
    }

    void write_float64(double v) {
        uint64_t u;
        std::memcpy(&u, &v, sizeof(double));
        write_int64(static_cast<int64_t>(u));
    }

    void write_string(const std::string& s) {
        uint16_t len = static_cast<uint16_t>(s.size());
        buffer.push_back(static_cast<uint8_t>(len));
        buffer.push_back(static_cast<uint8_t>(len >> 8));
        buffer.insert(buffer.end(), s.begin(), s.end());
    }
};

// Binary decoder for wire format
class Decoder {
public:
    const uint8_t* data;
    size_t size;
    size_t pos = 0;

    Decoder(const uint8_t* d, size_t s) : data(d), size(s) {}
    Decoder(const std::vector<uint8_t>& v) : data(v.data()), size(v.size()) {}

    void check_remaining(size_t needed) {
        if (pos + needed > size) {
            throw std::runtime_error("insufficient data for decode");
        }
    }

    bool read_bool() {
        check_remaining(1);
        return data[pos++] != 0x00;
    }

    int8_t read_int8() {
        check_remaining(1);
        return static_cast<int8_t>(data[pos++]);
    }

    int16_t read_int16() {
        check_remaining(2);
        uint16_t u = static_cast<uint16_t>(data[pos]) |
                     (static_cast<uint16_t>(data[pos + 1]) << 8);
        pos += 2;
        return static_cast<int16_t>(u);
    }

    int32_t read_int32() {
        check_remaining(4);
        uint32_t u = static_cast<uint32_t>(data[pos]) |
                     (static_cast<uint32_t>(data[pos + 1]) << 8) |
                     (static_cast<uint32_t>(data[pos + 2]) << 16) |
                     (static_cast<uint32_t>(data[pos + 3]) << 24);
        pos += 4;
        return static_cast<int32_t>(u);
    }

    int64_t read_int64() {
        check_remaining(8);
        uint64_t u = static_cast<uint64_t>(data[pos]) |
                     (static_cast<uint64_t>(data[pos + 1]) << 8) |
                     (static_cast<uint64_t>(data[pos + 2]) << 16) |
                     (static_cast<uint64_t>(data[pos + 3]) << 24) |
                     (static_cast<uint64_t>(data[pos + 4]) << 32) |
                     (static_cast<uint64_t>(data[pos + 5]) << 40) |
                     (static_cast<uint64_t>(data[pos + 6]) << 48) |
                     (static_cast<uint64_t>(data[pos + 7]) << 56);
        pos += 8;
        return static_cast<int64_t>(u);
    }

    float read_float32() {
        uint32_t u = static_cast<uint32_t>(read_int32());
        float f;
        std::memcpy(&f, &u, sizeof(float));
        return f;
    }

    double read_float64() {
        uint64_t u = static_cast<uint64_t>(read_int64());
        double d;
        std::memcpy(&d, &u, sizeof(double));
        return d;
    }

    std::string read_string() {
        check_remaining(2);
        uint16_t len = static_cast<uint16_t>(data[pos]) |
                       (static_cast<uint16_t>(data[pos + 1]) << 8);
        pos += 2;
        check_remaining(len);
        std::string s(reinterpret_cast<const char*>(data + pos), len);
        pos += len;
        return s;
    }

    uint16_t read_array_length() {
        check_remaining(2);
        uint16_t len = static_cast<uint16_t>(data[pos]) |
                       (static_cast<uint16_t>(data[pos + 1]) << 8);
        pos += 2;
        return len;
    }
};

// Encode Message to binary wire format
inline std::vector<uint8_t> encode_plugin_message(const std::vector<Plugin>& value) {
    Encoder enc;
    {
        uint16_t len = static_cast<uint16_t>(value.size());
        enc.write_byte(static_cast<uint8_t>(len));
        enc.write_byte(static_cast<uint8_t>(len >> 8));
    }
    for (const auto& elem : value) {
        enc.write_string(elem.Name);
        enc.write_string(elem.ManufacturerID);
        enc.write_string(elem.Type);
        enc.write_string(elem.Subtype);
        {
            uint16_t len = static_cast<uint16_t>(elem.Parameters.size());
            enc.write_byte(static_cast<uint8_t>(len));
            enc.write_byte(static_cast<uint8_t>(len >> 8));
        }
        for (const auto& elem : elem.Parameters) {
            enc.write_string(elem.DisplayName);
            enc.write_float32(elem.DefaultValue);
            enc.write_float32(elem.CurrentValue);
            enc.write_int32(elem.Address);
            enc.write_float32(elem.MaxValue);
            enc.write_float32(elem.MinValue);
            enc.write_string(elem.Unit);
            enc.write_string(elem.Identifier);
            enc.write_bool(elem.CanRamp);
            enc.write_bool(elem.IsWritable);
            enc.write_int64(elem.RawFlags);
            if (elem.IndexedValues.has_value()) {
                enc.write_byte(0x01);
                {
                    uint16_t len = static_cast<uint16_t>(elem.IndexedValues.value().size());
                    enc.write_byte(static_cast<uint8_t>(len));
                    enc.write_byte(static_cast<uint8_t>(len >> 8));
                }
                for (const auto& elem : elem.IndexedValues.value()) {
                    enc.write_string(elem);
                }
            } else {
                enc.write_byte(0x00);
            }
            if (elem.IndexedValuesSource.has_value()) {
                enc.write_byte(0x01);
                enc.write_string(elem.IndexedValuesSource.value());
            } else {
                enc.write_byte(0x00);
            }
        }
    }
    return enc.buffer;
}

// Decode Message from binary wire format
inline std::vector<Plugin> decode_plugin_message(const uint8_t* data, size_t size) {
    Decoder dec(data, size);
    std::vector<Plugin> result;
    {
        uint16_t len = dec.read_array_length();
        result.reserve(len);
        for (uint16_t i = 0; i < len; ++i) {
            Plugin elem0;
            elem0.Name = dec.read_string();
            elem0.ManufacturerID = dec.read_string();
            elem0.Type = dec.read_string();
            elem0.Subtype = dec.read_string();
            {
                uint16_t len = dec.read_array_length();
                elem0.Parameters.reserve(len);
                for (uint16_t i = 0; i < len; ++i) {
                    Parameter elem1;
                    elem1.DisplayName = dec.read_string();
                    elem1.DefaultValue = dec.read_float32();
                    elem1.CurrentValue = dec.read_float32();
                    elem1.Address = dec.read_int32();
                    elem1.MaxValue = dec.read_float32();
                    elem1.MinValue = dec.read_float32();
                    elem1.Unit = dec.read_string();
                    elem1.Identifier = dec.read_string();
                    elem1.CanRamp = dec.read_bool();
                    elem1.IsWritable = dec.read_bool();
                    elem1.RawFlags = dec.read_int64();
                    if (dec.read_bool()) {
                        std::vector<std::string> tmp;
                        {
                            uint16_t len = dec.read_array_length();
                            tmp.reserve(len);
                            for (uint16_t i = 0; i < len; ++i) {
                                std::string elem2;
                                elem2 = dec.read_string();
                                tmp.push_back(std::move(elem2));
                            }
                        }
                        elem1.IndexedValues = std::move(tmp);
                    }
                    if (dec.read_bool()) {
                        elem1.IndexedValuesSource = dec.read_string();
                    }
                    elem0.Parameters.push_back(std::move(elem1));
                }
            }
            result.push_back(std::move(elem0));
        }
    }
    return result;
}

inline std::vector<Plugin> decode_plugin_message(const std::vector<uint8_t>& data) {
    return decode_plugin_message(data.data(), data.size());
}

} // namespace test

#endif // TEST_H
