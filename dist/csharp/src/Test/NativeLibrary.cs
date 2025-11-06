using System;
using System.Runtime.InteropServices;

namespace Test
{
    internal static class NativeLibrary
    {
        private const string LibName = "ffire";

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        internal static extern IntPtr ffire_decode(byte[] data, int len);

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        internal static extern IntPtr ffire_encode(IntPtr msg);

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        internal static extern void ffire_free(IntPtr msg);

        [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
        internal static extern IntPtr ffire_get_error();

        internal static string GetLastError()
        {
            IntPtr errPtr = ffire_get_error();
            if (errPtr == IntPtr.Zero) return "Unknown error";
            return Marshal.PtrToStringAnsi(errPtr) ?? "Unknown error";
        }
    }
}
