import Foundation

// Swift calling C ABI via dylib (like Python does)

// C ABI function declarations
typealias PluginHandle = OpaquePointer

@_silgen_name("plugin_decode")
func plugin_decode(_ data: UnsafePointer<UInt8>, _ size: Int, _ error: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?) -> PluginHandle?

@_silgen_name("plugin_encode")
func plugin_encode(_ handle: PluginHandle, _ outData: UnsafeMutablePointer<UnsafeMutablePointer<UInt8>?>, _ error: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?) -> Int

@_silgen_name("plugin_free")
func plugin_free(_ handle: PluginHandle)

@_silgen_name("plugin_free_data")
func plugin_free_data(_ buffer: UnsafeMutablePointer<UInt8>)

func loadFixture(_ path: String) -> [UInt8] {
    guard let data = try? Data(contentsOf: URL(fileURLWithPath: path)) else {
        fatalError("Cannot load fixture: \(path)")
    }
    return Array(data)
}

// Main test
let data = loadFixture("../common/complex.bin")
let iterations = 100

// Warmup
for _ in 0..<10 {
    var error: UnsafeMutablePointer<CChar>? = nil
    guard let handle = data.withUnsafeBytes({ bytes in
        plugin_decode(bytes.bindMemory(to: UInt8.self).baseAddress!, data.count, &error)
    }) else {
        fatalError("Decode failed")
    }
    plugin_free(handle)
}

// Decode benchmark
let decodeStart = Date()
for _ in 0..<iterations {
    var error: UnsafeMutablePointer<CChar>? = nil
    guard let handle = data.withUnsafeBytes({ bytes in
        plugin_decode(bytes.bindMemory(to: UInt8.self).baseAddress!, data.count, &error)
    }) else {
        fatalError("Decode failed")
    }
    plugin_free(handle)
}
let decodeEnd = Date()
let decodeUs = Int(decodeEnd.timeIntervalSince(decodeStart) * 1_000_000 / Double(iterations))

// Encode benchmark
var error: UnsafeMutablePointer<CChar>? = nil
guard let handle = data.withUnsafeBytes({ bytes in
    plugin_decode(bytes.bindMemory(to: UInt8.self).baseAddress!, data.count, &error)
}) else {
    fatalError("Decode failed")
}

let encodeStart = Date()
var encodedSize = 0
var buffer: UnsafeMutablePointer<UInt8>? = nil
for _ in 0..<iterations {
    if buffer != nil {
        plugin_free_data(buffer!)
    }
    var error: UnsafeMutablePointer<CChar>? = nil
    encodedSize = plugin_encode(handle, &buffer, &error)
    if encodedSize == 0 {
        fatalError("Encode failed")
    }
}
let encodeEnd = Date()
let encodeUs = Int(encodeEnd.timeIntervalSince(encodeStart) * 1_000_000 / Double(iterations))

if buffer != nil {
    plugin_free_data(buffer!)
}

plugin_free(handle)

// Output JSON
print("{\"decode_us\":\(decodeUs),\"encode_us\":\(encodeUs),\"size_bytes\":\(encodedSize),\"iterations\":\(iterations)}")
