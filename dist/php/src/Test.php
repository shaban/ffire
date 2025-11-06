<?php

namespace Test;

use FFI;
use FFI\CData;
use Exception;

/**
 * FFire exception
 */
class FFireException extends Exception {}

/**
 * FFI Library loader
 */
class FFILibrary {
    private static ?FFI $ffi = null;

    public static function load(): FFI {
        if (self::$ffi !== null) {
            return self::$ffi;
        }

        // Determine library path based on OS
        $libName = match(PHP_OS_FAMILY) {
            'Darwin' => 'libffire.dylib',
            'Linux' => 'libffire.so',
            'Windows' => 'ffire.dll',
            default => 'libffire.so'
        };

        $libPath = __DIR__ . '/../lib/' . $libName;

        if (!file_exists($libPath)) {
            throw new FFireException("Library not found: {$libPath}");
        }

        $ffiCode = "
            void* message_decode(const uint8_t* data, size_t size, char** error);
            size_t message_encode(void* handle, uint8_t** data, char** error);
            void message_free(void* handle);
            void message_free_data(uint8_t* data);
            void message_free_error(char* error);
        ";

        self::$ffi = FFI::cdef($ffiCode, $libPath);
        return self::$ffi;
    }
}

/**
 * Message message type
 */
class Message {
    private CData $handle;
    private FFI $ffi;
    private bool $freed = false;

    private function __construct(CData $handle, FFI $ffi) {
        $this->handle = $handle;
        $this->ffi = $ffi;
    }

    public function __destruct() {
        $this->free();
    }

    /**
     * Decode a Message from binary data
     * @param string $data Binary data to decode
     * @return Message Decoded message object
     * @throws FFireException if decoding fails
     */
    public static function decode(string $data): self {
        $ffi = FFILibrary::load();
        $size = strlen($data);

        // Allocate memory for data
        $dataPtr = $ffi->new('uint8_t[' . $size . ']', false);
        FFI::memcpy($dataPtr, $data, $size);

        // Allocate error pointer
        $errorPtr = $ffi->new('char*');
        $errorPtr->cdata = null;

        $handle = $ffi->message_decode($dataPtr, $size, FFI::addr($errorPtr));

        if (FFI::isNull($handle)) {
            $error = 'Unknown error';
            if (!FFI::isNull($errorPtr->cdata)) {
                $error = FFI::string($errorPtr->cdata);
                $ffi->message_free_error($errorPtr->cdata);
            }
            throw new FFireException("Failed to decode Message: {$error}");
        }

        return new self($handle, $ffi);
    }

    /**
     * Encode this Message to binary data
     * @return string Encoded binary data
     * @throws FFireException if encoding fails
     */
    public function encode(): string {
        if ($this->freed) {
            throw new FFireException('Message already freed');
        }

        $dataPtr = $this->ffi->new('uint8_t*');
        $dataPtr->cdata = null;
        $errorPtr = $this->ffi->new('char*');
        $errorPtr->cdata = null;

        $size = $this->ffi->message_encode($this->handle, FFI::addr($dataPtr), FFI::addr($errorPtr));

        if ($size === 0) {
            $error = 'Unknown error';
            if (!FFI::isNull($errorPtr->cdata)) {
                $error = FFI::string($errorPtr->cdata);
                $this->ffi->message_free_error($errorPtr->cdata);
            }
            throw new FFireException("Failed to encode Message: {$error}");
        }

        $result = FFI::string($dataPtr->cdata, $size);
        $this->ffi->message_free_data($dataPtr->cdata);

        return $result;
    }

    /**
     * Free the native resources
     */
    public function free(): void {
        if (!$this->freed) {
            $this->ffi->message_free($this->handle);
            $this->freed = true;
        }
    }
}

