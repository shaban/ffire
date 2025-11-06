package com.ffire.test;

/**
 * Native library loader
 */
class NativeLibrary {
    private static boolean loaded = false;

    static {
        try {
            // Try to load from java.library.path
            System.loadLibrary("ffire");
            loaded = true;
        } catch (UnsatisfiedLinkError e) {
            // Library not in java.library.path
            loaded = false;
        }
    }

    static void ensureLoaded() {
        if (!loaded) {
            throw new UnsatisfiedLinkError(
                "Native library 'ffire' not found. " +
                "Add library path with -Djava.library.path=<path>"
            );
        }
    }
}
