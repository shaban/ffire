using System;
using System.Buffers;
using System.Buffers.Binary;
using System.Diagnostics;
using System.IO;
using System.Runtime.InteropServices;

namespace CSharpPerfTest
{
    // Test struct: simple struct with mixed types
    public struct TestMessage
    {
        public bool BoolVal;
        public int IntVal;
        public long LongVal;
        public float FloatVal;
        public double DoubleVal;
    }

    class Program
    {
        const int Iterations = 100000;

        static void Main(string[] args)
        {
            Console.WriteLine("C# Serialization Performance Comparison\n");

            var msg = new TestMessage
            {
                BoolVal = true,
                IntVal = 42,
                LongVal = 123456789L,
                FloatVal = 3.14f,
                DoubleVal = 2.718281828
            };

            // Warmup
            for (int i = 0; i < 1000; i++)
            {
                TestBinaryWriter(msg);
                TestSpanModern(msg);
                TestUnsafe(msg);
            }

            Console.WriteLine($"Running {Iterations} iterations...\n");

            // Test 1: BinaryWriter (traditional)
            var sw1 = Stopwatch.StartNew();
            for (int i = 0; i < Iterations; i++)
            {
                TestBinaryWriter(msg);
            }
            sw1.Stop();
            Console.WriteLine($"BinaryWriter:     {sw1.ElapsedMilliseconds}ms ({(double)sw1.ElapsedTicks / Iterations} ticks/op)");

            // Test 2: Span<byte> + BinaryPrimitives (modern)
            var sw2 = Stopwatch.StartNew();
            for (int i = 0; i < Iterations; i++)
            {
                TestSpanModern(msg);
            }
            sw2.Stop();
            Console.WriteLine($"Span (modern):    {sw2.ElapsedMilliseconds}ms ({(double)sw2.ElapsedTicks / Iterations} ticks/op)");

            // Test 3: Unsafe pointers (fastest)
            var sw3 = Stopwatch.StartNew();
            for (int i = 0; i < Iterations; i++)
            {
                TestUnsafe(msg);
            }
            sw3.Stop();
            Console.WriteLine($"Unsafe pointers:  {sw3.ElapsedMilliseconds}ms ({(double)sw3.ElapsedTicks / Iterations} ticks/op)");

            Console.WriteLine("\nRelative Performance:");
            double baseline = sw2.ElapsedTicks;
            Console.WriteLine($"BinaryWriter:    {sw1.ElapsedTicks / baseline:F2}x");
            Console.WriteLine($"Span (modern):   {sw2.ElapsedTicks / baseline:F2}x (baseline)");
            Console.WriteLine($"Unsafe:          {sw3.ElapsedTicks / baseline:F2}x");

            Console.WriteLine("\nRecommendation:");
            if (sw2.ElapsedTicks < sw3.ElapsedTicks * 1.2)
            {
                Console.WriteLine("Use Span<byte> - modern, safe, nearly as fast as unsafe");
            }
            else
            {
                Console.WriteLine($"Unsafe is {baseline / sw3.ElapsedTicks:F2}x faster - consider for hot paths");
            }
        }

        // Approach 1: BinaryWriter (traditional)
        static byte[] TestBinaryWriter(TestMessage msg)
        {
            using var stream = new MemoryStream(32);
            using var writer = new BinaryWriter(stream);
            
            writer.Write(msg.BoolVal);
            writer.Write(msg.IntVal);
            writer.Write(msg.LongVal);
            writer.Write(msg.FloatVal);
            writer.Write(msg.DoubleVal);
            
            return stream.ToArray();
        }

        // Approach 2: Span<byte> + BinaryPrimitives (modern, recommended)
        static byte[] TestSpanModern(TestMessage msg)
        {
            byte[] buffer = new byte[25]; // 1 + 4 + 8 + 4 + 8
            Span<byte> span = buffer;
            int offset = 0;

            span[offset++] = msg.BoolVal ? (byte)1 : (byte)0;
            
            BinaryPrimitives.WriteInt32LittleEndian(span.Slice(offset, 4), msg.IntVal);
            offset += 4;
            
            BinaryPrimitives.WriteInt64LittleEndian(span.Slice(offset, 8), msg.LongVal);
            offset += 8;
            
            BinaryPrimitives.WriteSingleLittleEndian(span.Slice(offset, 4), msg.FloatVal);
            offset += 4;
            
            BinaryPrimitives.WriteDoubleLittleEndian(span.Slice(offset, 8), msg.DoubleVal);
            
            return buffer;
        }

        // Approach 3: Unsafe pointers (fastest, but unsafe)
        static unsafe byte[] TestUnsafe(TestMessage msg)
        {
            byte[] buffer = new byte[25];
            fixed (byte* ptr = buffer)
            {
                int offset = 0;
                
                ptr[offset++] = msg.BoolVal ? (byte)1 : (byte)0;
                
                *(int*)(ptr + offset) = msg.IntVal;
                offset += 4;
                
                *(long*)(ptr + offset) = msg.LongVal;
                offset += 8;
                
                *(float*)(ptr + offset) = msg.FloatVal;
                offset += 4;
                
                *(double*)(ptr + offset) = msg.DoubleVal;
            }
            return buffer;
        }
    }
}
