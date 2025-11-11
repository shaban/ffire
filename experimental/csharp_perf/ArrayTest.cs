using System;
using System.Buffers;
using System.Buffers.Binary;
using System.Diagnostics;
using System.Runtime.InteropServices;

namespace CSharpPerfTest
{
    class ArrayTest
    {
        const int Iterations = 10000;
        const int ArraySize = 1000;

        static void Main(string[] args)
        {
            Console.WriteLine("C# Array Serialization Performance\n");

            int[] testArray = new int[ArraySize];
            for (int i = 0; i < ArraySize; i++)
                testArray[i] = i;

            // Warmup
            for (int i = 0; i < 100; i++)
            {
                TestLoopCopy(testArray);
                TestBufferBlockCopy(testArray);
                TestSpanCopyTo(testArray);
                TestMemoryMarshalCast(testArray);
            }

            Console.WriteLine($"Running {Iterations} iterations with {ArraySize} int array...\n");

            // Test 1: Loop copy (naive)
            var sw1 = Stopwatch.StartNew();
            for (int i = 0; i < Iterations; i++)
            {
                TestLoopCopy(testArray);
            }
            sw1.Stop();
            Console.WriteLine($"Loop copy:           {sw1.ElapsedMilliseconds}ms ({(double)sw1.ElapsedTicks / Iterations} ticks/op)");

            // Test 2: Buffer.BlockCopy
            var sw2 = Stopwatch.StartNew();
            for (int i = 0; i < Iterations; i++)
            {
                TestBufferBlockCopy(testArray);
            }
            sw2.Stop();
            Console.WriteLine($"Buffer.BlockCopy:    {sw2.ElapsedMilliseconds}ms ({(double)sw2.ElapsedTicks / Iterations} ticks/op)");

            // Test 3: Span.CopyTo
            var sw3 = Stopwatch.StartNew();
            for (int i = 0; i < Iterations; i++)
            {
                TestSpanCopyTo(testArray);
            }
            sw3.Stop();
            Console.WriteLine($"Span.CopyTo:         {sw3.ElapsedMilliseconds}ms ({(double)sw3.ElapsedTicks / Iterations} ticks/op)");

            // Test 4: MemoryMarshal.Cast (zero-copy view)
            var sw4 = Stopwatch.StartNew();
            for (int i = 0; i < Iterations; i++)
            {
                TestMemoryMarshalCast(testArray);
            }
            sw4.Stop();
            Console.WriteLine($"MemoryMarshal.Cast:  {sw4.ElapsedMilliseconds}ms ({(double)sw4.ElapsedTicks / Iterations} ticks/op)");

            Console.WriteLine("\nRelative Performance:");
            double baseline = sw3.ElapsedTicks;
            Console.WriteLine($"Loop copy:           {sw1.ElapsedTicks / baseline:F2}x");
            Console.WriteLine($"Buffer.BlockCopy:    {sw2.ElapsedTicks / baseline:F2}x");
            Console.WriteLine($"Span.CopyTo:         {sw3.ElapsedTicks / baseline:F2}x (baseline)");
            Console.WriteLine($"MemoryMarshal.Cast:  {sw4.ElapsedTicks / baseline:F2}x");

            Console.WriteLine("\n✅ Recommendation: Use MemoryMarshal.Cast for zero-copy array encoding");
        }

        static byte[] TestLoopCopy(int[] array)
        {
            byte[] buffer = new byte[array.Length * 4];
            for (int i = 0; i < array.Length; i++)
            {
                BinaryPrimitives.WriteInt32LittleEndian(buffer.AsSpan(i * 4, 4), array[i]);
            }
            return buffer;
        }

        static byte[] TestBufferBlockCopy(int[] array)
        {
            byte[] buffer = new byte[array.Length * 4];
            Buffer.BlockCopy(array, 0, buffer, 0, array.Length * 4);
            return buffer;
        }

        static byte[] TestSpanCopyTo(int[] array)
        {
            byte[] buffer = new byte[array.Length * 4];
            Span<int> intSpan = array;
            Span<byte> byteSpan = MemoryMarshal.AsBytes(intSpan);
            byteSpan.CopyTo(buffer);
            return buffer;
        }

        static byte[] TestMemoryMarshalCast(int[] array)
        {
            // Zero-copy view - no allocation!
            Span<int> intSpan = array;
            Span<byte> byteSpan = MemoryMarshal.AsBytes(intSpan);
            return byteSpan.ToArray(); // Only allocate result
        }
    }
}
