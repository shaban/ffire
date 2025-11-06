//
// FFire test Swift bindings
//
// This module provides Swift bindings to the FFire binary serialization library
// via a C ABI dynamic library.
//

import Foundation

// Load the C library
#if os(macOS)
private let libName = "libffire.dylib"
#elseif os(Linux)
private let libName = "libffire.so"
#elseif os(Windows)
private let libName = "ffire.dll"
#endif

// Library handle
private let libraryPath: String = {
    let bundlePath = Bundle.main.bundlePath
    return "\(bundlePath)/lib/\(libName)"
}()

/// FFire error type
public enum FFireError: Error {
    case decodeFailed(String)
    case encodeFailed(String)
    case libraryError(String)
}

// C function declarations for Message
@_silgen_name("message_decode")
private func c_message_decode(
    _ data: UnsafePointer<UInt8>,
    _ size: Int,
    _ error: UnsafeMutablePointer<UnsafePointer<CChar>?>
) -> OpaquePointer?

@_silgen_name("message_encode")
private func c_message_encode(
    _ handle: OpaquePointer,
    _ data: UnsafeMutablePointer<UnsafePointer<UInt8>?>,
    _ error: UnsafeMutablePointer<UnsafePointer<CChar>?>
) -> Int

@_silgen_name("message_free")
private func c_message_free(_ handle: OpaquePointer)

@_silgen_name("message_free_data")
private func c_message_free_data(_ data: UnsafePointer<UInt8>)

@_silgen_name("message_free_error")
private func c_message_free_error(_ error: UnsafePointer<CChar>)

/// Message message type
public class Message {
    private var handle: OpaquePointer
    private var freed = false

    private init(handle: OpaquePointer) {
        self.handle = handle
    }

    deinit {
        free()
    }

    /// Decode a Message from binary data
    /// - Parameter data: Binary data to decode
    /// - Returns: Decoded Message object
    /// - Throws: FFireError if decoding fails
    public static func decode(_ data: Data) throws -> Message {
        var errorPtr: UnsafePointer<CChar>? = nil
        
        let handle = data.withUnsafeBytes { (ptr: UnsafeRawBufferPointer) -> OpaquePointer? in
            guard let baseAddress = ptr.baseAddress?.assumingMemoryBound(to: UInt8.self) else {
                return nil
            }
            return c_message_decode(baseAddress, data.count, &errorPtr)
        }
        
        if let handle = handle {
            return Message(handle: handle)
        } else {
            let errorMsg: String
            if let errorPtr = errorPtr {
                errorMsg = String(cString: errorPtr)
                c_message_free_error(errorPtr)
            } else {
                errorMsg = "Unknown error"
            }
            throw FFireError.decodeFailed(errorMsg)
        }
    }
    
    /// Encode this Message to binary data
    /// - Returns: Encoded binary data
    /// - Throws: FFireError if encoding fails
    public func encode() throws -> Data {
        guard !freed else {
            throw FFireError.encodeFailed("Message already freed")
        }
        
        var dataPtr: UnsafePointer<UInt8>? = nil
        var errorPtr: UnsafePointer<CChar>? = nil
        
        let size = c_message_encode(handle, &dataPtr, &errorPtr)
        
        if size > 0, let dataPtr = dataPtr {
            let data = Data(bytes: dataPtr, count: size)
            c_message_free_data(dataPtr)
            return data
        } else {
            let errorMsg: String
            if let errorPtr = errorPtr {
                errorMsg = String(cString: errorPtr)
                c_message_free_error(errorPtr)
            } else {
                errorMsg = "Unknown error"
            }
            throw FFireError.encodeFailed(errorMsg)
        }
    }
    
    /// Free the native resources
    public func free() {
        if !freed {
            c_message_free(handle)
            freed = true
        }
    }
}

