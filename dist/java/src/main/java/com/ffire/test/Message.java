package com.ffire.test;

/**
 * Message message type
 */
public class Message implements AutoCloseable {
    static {
        NativeLibrary.ensureLoaded();
    }

    private long handle;
    private boolean freed = false;

    private Message(long handle) {
        this.handle = handle;
    }

    // Native method declarations
    private static native long messageDecode(byte[] data, int size);
    private static native byte[] messageEncode(long handle);
    private static native void messageFree(long handle);
    private static native String messageGetError();

    /**
     * Decode a Message from binary data
     * @param data Binary data to decode
     * @return Decoded Message object
     * @throws FFireException if decoding fails
     */
    public static Message decode(byte[] data) throws FFireException {
        long handle = messageDecode(data, data.length);
        if (handle == 0) {
            String error = messageGetError();
            if (error == null) error = "Unknown error";
            throw new FFireException("Failed to decode Message: " + error);
        }
        return new Message(handle);
    }

    /**
     * Encode this Message to binary data
     * @return Encoded binary data
     * @throws FFireException if encoding fails
     */
    public byte[] encode() throws FFireException {
        if (freed) {
            throw new FFireException("Message already freed");
        }

        byte[] result = messageEncode(handle);
        if (result == null) {
            String error = messageGetError();
            if (error == null) error = "Unknown error";
            throw new FFireException("Failed to encode Message: " + error);
        }
        return result;
    }

    /**
     * Free the native resources
     */
    public void free() {
        if (!freed && handle != 0) {
            messageFree(handle);
            freed = true;
            handle = 0;
        }
    }

    @Override
    public void close() {
        free();
    }

    @Override
    protected void finalize() throws Throwable {
        try {
            free();
        } finally {
            super.finalize();
        }
    }
}
