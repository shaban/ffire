using System;
using System.Runtime.InteropServices;

namespace Test
{
    public class Message : IDisposable
    {
        private IntPtr handle;
        private bool disposed = false;

        private Message(IntPtr handle) { this.handle = handle; }

        public static Message Decode(byte[] data)
        {
            if (data == null) throw new ArgumentNullException(nameof(data));
            IntPtr handle = NativeLibrary.ffire_decode(data, data.Length);
            if (handle == IntPtr.Zero)
            {
                throw new TestException($"Decode failed: {NativeLibrary.GetLastError()}");
            }
            return new Message(handle);
        }

        public byte[] Encode()
        {
            if (disposed) throw new ObjectDisposedException(nameof(Message));
            IntPtr resultPtr = NativeLibrary.ffire_encode(handle);
            if (resultPtr == IntPtr.Zero)
            {
                throw new TestException($"Encode failed: {NativeLibrary.GetLastError()}");
            }
            int length = Marshal.ReadInt32(resultPtr);
            byte[] data = new byte[length];
            Marshal.Copy(resultPtr + 4, data, 0, length);
            return data;
        }

        public void Free() { Dispose(); }

        public void Dispose()
        {
            Dispose(true);
            GC.SuppressFinalize(this);
        }

        protected virtual void Dispose(bool disposing)
        {
            if (!disposed)
            {
                if (handle != IntPtr.Zero)
                {
                    NativeLibrary.ffire_free(handle);
                    handle = IntPtr.Zero;
                }
                disposed = true;
            }
        }

        ~Message() { Dispose(false); }
    }
}
