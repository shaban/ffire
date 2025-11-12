// BLUEPRINT: Clean reference implementation achieving 19.4k ns total
// This file shows the exact code patterns that achieve optimal performance
// Generator will follow these patterns exactly

using System;
using System.Buffers.Binary;
using System.Diagnostics;
using System.IO;
using System.Runtime.CompilerServices;
using System.Text;
using System.Text.Json;

namespace Blueprint
{
    // ============================================================================
    // GENERATOR: Data structures from schema
    // ============================================================================
    // Schema defines: struct Parameter with fields
    // Schema defines: struct Plugin with fields  
    // Schema defines: message PluginListMessage wrapping Plugin[]
    
    public struct Parameter
    {
        // GENERATOR: For each field, generate simple property (NO UTF-8 caching!)
        public string DisplayName { get; set; }
        public float DefaultValue { get; set; }
        public float CurrentValue { get; set; }
        public int Address { get; set; }
        public float MaxValue { get; set; }
        public float MinValue { get; set; }
        public string Unit { get; set; }
        public string Identifier { get; set; }
        public bool CanRamp { get; set; }
        public bool IsWritable { get; set; }
        public long RawFlags { get; set; }
        public string[] IndexedValues { get; set; }
        public string? IndexedValuesSource { get; set; }

        // GENERATOR: ComputeMaxSize - use string.Length * 3 for UTF-8 expansion
        // Pattern: For each string field: size += 2 + (field?.Length ?? 0) * 3
        // Pattern: For each primitive: size += sizeof(type)
        // Pattern: For each array of strings: foreach item, size += 2 + (item?.Length ?? 0) * 3
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public int ComputeMaxSize()
        {
            int size = 0;
            size += 2 + (DisplayName?.Length ?? 0) * 3;  // string: length prefix + max UTF-8 bytes
            size += 4;  // float32
            size += 4;  // float32
            size += 4;  // int32
            size += 4;  // float32
            size += 4;  // float32
            size += 2 + (Unit?.Length ?? 0) * 3;
            size += 2 + (Identifier?.Length ?? 0) * 3;
            size += 1;  // bool
            size += 1;  // bool
            size += 8;  // int64
            
            // Array: 1 byte presence + 2 bytes length + elements
            size += 1;
            if (IndexedValues != null)
            {
                size += 2;
                foreach (var item in IndexedValues)
                {
                    size += 2 + (item?.Length ?? 0) * 3;
                }
            }
            
            // Optional field: 1 byte presence + data
            size += 1;
            if (IndexedValuesSource != null)
            {
                size += 2 + IndexedValuesSource.Length * 3;
            }
            
            return size;
        }

        // GENERATOR: Encode method pattern
        // Pattern: Allocate max buffer, encode, trim if needed
        public byte[] Encode()
        {
            int maxSize = ComputeMaxSize();
            byte[] buffer = new byte[maxSize];
            int offset = 0;
            EncodeToInternal(buffer, ref offset);
            
            // GENERATOR: Always trim array if under-used
            if (offset < maxSize)
            {
                Array.Resize(ref buffer, offset);
            }
            return buffer;
        }

        // GENERATOR: EncodeToInternal - the core encoding logic
        // Pattern: Use inline bit-shifting for all primitives except i8/u8/bool
        // Pattern: Use single-pass Encoding.UTF8.GetBytes for strings
        // Pattern: Use unsafe pointer cast for floats/doubles
        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        internal unsafe void EncodeToInternal(byte[] buffer, ref int offset)
        {
            // GENERATOR: String encoding pattern (single-pass, writes length after encoding)
            int byteCount_displayname = Encoding.UTF8.GetBytes(DisplayName ?? "", buffer.AsSpan(offset + 2));
            buffer[offset++] = (byte)byteCount_displayname;
            buffer[offset++] = (byte)(byteCount_displayname >> 8);
            offset += byteCount_displayname;
            
            // GENERATOR: Float32 encoding pattern (unsafe cast + inline shifts)
            {
                float f = DefaultValue;
                uint v = *(uint*)&f;
                buffer[offset++] = (byte)v;
                buffer[offset++] = (byte)(v >> 8);
                buffer[offset++] = (byte)(v >> 16);
                buffer[offset++] = (byte)(v >> 24);
            }
            
            {
                float f = CurrentValue;
                uint v = *(uint*)&f;
                buffer[offset++] = (byte)v;
                buffer[offset++] = (byte)(v >> 8);
                buffer[offset++] = (byte)(v >> 16);
                buffer[offset++] = (byte)(v >> 24);
            }
            
            // GENERATOR: Int32 encoding pattern (inline shifts)
            {
                int v = Address;
                buffer[offset++] = (byte)v;
                buffer[offset++] = (byte)(v >> 8);
                buffer[offset++] = (byte)(v >> 16);
                buffer[offset++] = (byte)(v >> 24);
            }
            
            {
                float f = MaxValue;
                uint v = *(uint*)&f;
                buffer[offset++] = (byte)v;
                buffer[offset++] = (byte)(v >> 8);
                buffer[offset++] = (byte)(v >> 16);
                buffer[offset++] = (byte)(v >> 24);
            }
            
            {
                float f = MinValue;
                uint v = *(uint*)&f;
                buffer[offset++] = (byte)v;
                buffer[offset++] = (byte)(v >> 8);
                buffer[offset++] = (byte)(v >> 16);
                buffer[offset++] = (byte)(v >> 24);
            }
            
            // More strings
            int byteCount_unit = Encoding.UTF8.GetBytes(Unit ?? "", buffer.AsSpan(offset + 2));
            buffer[offset++] = (byte)byteCount_unit;
            buffer[offset++] = (byte)(byteCount_unit >> 8);
            offset += byteCount_unit;
            
            int byteCount_identifier = Encoding.UTF8.GetBytes(Identifier ?? "", buffer.AsSpan(offset + 2));
            buffer[offset++] = (byte)byteCount_identifier;
            buffer[offset++] = (byte)(byteCount_identifier >> 8);
            offset += byteCount_identifier;
            
            // GENERATOR: Bool encoding pattern (direct byte write)
            buffer[offset++] = (byte)(CanRamp ? 1 : 0);
            buffer[offset++] = (byte)(IsWritable ? 1 : 0);
            
            // GENERATOR: Int64 encoding pattern (inline shifts)
            {
                long v = RawFlags;
                buffer[offset++] = (byte)v;
                buffer[offset++] = (byte)(v >> 8);
                buffer[offset++] = (byte)(v >> 16);
                buffer[offset++] = (byte)(v >> 24);
                buffer[offset++] = (byte)(v >> 32);
                buffer[offset++] = (byte)(v >> 40);
                buffer[offset++] = (byte)(v >> 48);
                buffer[offset++] = (byte)(v >> 56);
            }
            
            // GENERATOR: Array encoding pattern (presence byte + length + elements)
            if (IndexedValues != null)
            {
                buffer[offset++] = 1;
                ushort arrayLen = (ushort)IndexedValues.Length;
                buffer[offset++] = (byte)arrayLen;
                buffer[offset++] = (byte)(arrayLen >> 8);
                
                foreach (var val in IndexedValues)
                {
                    int byteCount = Encoding.UTF8.GetBytes(val ?? "", buffer.AsSpan(offset + 2));
                    buffer[offset++] = (byte)byteCount;
                    buffer[offset++] = (byte)(byteCount >> 8);
                    offset += byteCount;
                }
            }
            else
            {
                buffer[offset++] = 0;
            }
            
            // GENERATOR: Optional field pattern
            if (IndexedValuesSource != null)
            {
                buffer[offset++] = 1;
                int byteCount = Encoding.UTF8.GetBytes(IndexedValuesSource, buffer.AsSpan(offset + 2));
                buffer[offset++] = (byte)byteCount;
                buffer[offset++] = (byte)(byteCount >> 8);
                offset += byteCount;
            }
            else
            {
                buffer[offset++] = 0;
            }
        }

        // GENERATOR: Decode pattern - use BinaryPrimitives (proven fast)
        // NOTE: We kept BinaryPrimitives for decode because it's already optimized
        public static Parameter Decode(byte[] data)
        {
            ReadOnlySpan<byte> buffer = data;
            int offset = 0;
            return DecodeFrom(buffer, ref offset);
        }

        internal static Parameter DecodeFrom(ReadOnlySpan<byte> buffer, ref int offset)
        {
            var obj = new Parameter();
            
            // GENERATOR: String decode pattern (read length, decode string)
            int len_displayname = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.DisplayName = len_displayname > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len_displayname)) : "";
            offset += len_displayname;
            
            // GENERATOR: Float32 decode pattern (BinaryPrimitives)
            obj.DefaultValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4));
            offset += 4;
            
            obj.CurrentValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4));
            offset += 4;
            
            // GENERATOR: Int32 decode pattern
            obj.Address = BinaryPrimitives.ReadInt32LittleEndian(buffer.Slice(offset, 4));
            offset += 4;
            
            obj.MaxValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4));
            offset += 4;
            
            obj.MinValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4));
            offset += 4;
            
            int len_unit = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.Unit = len_unit > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len_unit)) : "";
            offset += len_unit;
            
            int len_identifier = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.Identifier = len_identifier > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len_identifier)) : "";
            offset += len_identifier;
            
            // GENERATOR: Bool decode pattern
            obj.CanRamp = buffer[offset++] != 0;
            obj.IsWritable = buffer[offset++] != 0;
            
            // GENERATOR: Int64 decode pattern
            obj.RawFlags = BinaryPrimitives.ReadInt64LittleEndian(buffer.Slice(offset, 8));
            offset += 8;
            
            // GENERATOR: Array decode pattern
            if (buffer[offset++] == 1)
            {
                int arrayLen = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
                offset += 2;
                obj.IndexedValues = new string[arrayLen];
                for (int i = 0; i < arrayLen; i++)
                {
                    int itemLen = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
                    offset += 2;
                    obj.IndexedValues[i] = itemLen > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, itemLen)) : "";
                    offset += itemLen;
                }
            }
            
            // GENERATOR: Optional field decode pattern
            if (buffer[offset++] == 1)
            {
                int len = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
                offset += 2;
                obj.IndexedValuesSource = len > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len)) : "";
                offset += len;
            }
            
            return obj;
        }
    }

    public struct Plugin
    {
        public string Name { get; set; }
        public string ManufacturerID { get; set; }
        public string Type { get; set; }
        public string Subtype { get; set; }
        public Parameter[] Parameters { get; set; }

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        public int ComputeMaxSize()
        {
            int size = 0;
            size += 2 + (Name?.Length ?? 0) * 3;
            size += 2 + (ManufacturerID?.Length ?? 0) * 3;
            size += 2 + (Type?.Length ?? 0) * 3;
            size += 2 + (Subtype?.Length ?? 0) * 3;
            size += 2;  // array length
            
            if (Parameters != null)
            {
                foreach (var param in Parameters)
                {
                    size += param.ComputeMaxSize();
                }
            }
            
            return size;
        }

        public byte[] Encode()
        {
            int maxSize = ComputeMaxSize();
            byte[] buffer = new byte[maxSize];
            int offset = 0;
            EncodeToInternal(buffer, ref offset);
            if (offset < maxSize)
            {
                Array.Resize(ref buffer, offset);
            }
            return buffer;
        }

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        internal unsafe void EncodeToInternal(byte[] buffer, ref int offset)
        {
            int byteCount_name = Encoding.UTF8.GetBytes(Name ?? "", buffer.AsSpan(offset + 2));
            buffer[offset++] = (byte)byteCount_name;
            buffer[offset++] = (byte)(byteCount_name >> 8);
            offset += byteCount_name;
            
            int byteCount_manufacturerId = Encoding.UTF8.GetBytes(ManufacturerID ?? "", buffer.AsSpan(offset + 2));
            buffer[offset++] = (byte)byteCount_manufacturerId;
            buffer[offset++] = (byte)(byteCount_manufacturerId >> 8);
            offset += byteCount_manufacturerId;
            
            int byteCount_type = Encoding.UTF8.GetBytes(Type ?? "", buffer.AsSpan(offset + 2));
            buffer[offset++] = (byte)byteCount_type;
            buffer[offset++] = (byte)(byteCount_type >> 8);
            offset += byteCount_type;
            
            int byteCount_subtype = Encoding.UTF8.GetBytes(Subtype ?? "", buffer.AsSpan(offset + 2));
            buffer[offset++] = (byte)byteCount_subtype;
            buffer[offset++] = (byte)(byteCount_subtype >> 8);
            offset += byteCount_subtype;
            
            // GENERATOR: Array of structs pattern
            ushort paramLen = (ushort)(Parameters?.Length ?? 0);
            buffer[offset++] = (byte)paramLen;
            buffer[offset++] = (byte)(paramLen >> 8);
            
            if (Parameters != null)
            {
                foreach (var param in Parameters)
                {
                    param.EncodeToInternal(buffer, ref offset);
                }
            }
        }

        public static Plugin Decode(byte[] data)
        {
            ReadOnlySpan<byte> buffer = data;
            int offset = 0;
            return DecodeFrom(buffer, ref offset);
        }

        internal static Plugin DecodeFrom(ReadOnlySpan<byte> buffer, ref int offset)
        {
            var obj = new Plugin();
            
            int len_name = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.Name = len_name > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len_name)) : "";
            offset += len_name;
            
            int len_manufacturerId = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.ManufacturerID = len_manufacturerId > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len_manufacturerId)) : "";
            offset += len_manufacturerId;
            
            int len_type = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.Type = len_type > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len_type)) : "";
            offset += len_type;
            
            int len_subtype = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.Subtype = len_subtype > 0 ? Encoding.UTF8.GetString(buffer.Slice(offset, len_subtype)) : "";
            offset += len_subtype;
            
            int paramLen = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.Parameters = new Parameter[paramLen];
            
            for (int i = 0; i < paramLen; i++)
            {
                obj.Parameters[i] = Parameter.DecodeFrom(buffer, ref offset);
            }
            
            return obj;
        }
    }

    // ============================================================================
    // GENERATOR: Message wrapper (top-level array holder)
    // ============================================================================
    public struct PluginListMessage
    {
        public Plugin[] Items { get; set; }

        public int ComputeMaxSize()
        {
            int size = 2;  // array length
            if (Items != null)
            {
                foreach (var item in Items)
                {
                    size += item.ComputeMaxSize();
                }
            }
            return size;
        }

        public byte[] Encode()
        {
            int maxSize = ComputeMaxSize();
            byte[] buffer = new byte[maxSize];
            int offset = 0;
            EncodeToInternal(buffer, ref offset);
            
            // GENERATOR: Critical - always trim for messages too!
            if (offset < maxSize)
            {
                Array.Resize(ref buffer, offset);
            }
            return buffer;
        }

        [MethodImpl(MethodImplOptions.AggressiveInlining)]
        internal unsafe void EncodeToInternal(byte[] buffer, ref int offset)
        {
            ushort len = (ushort)(Items?.Length ?? 0);
            buffer[offset++] = (byte)len;
            buffer[offset++] = (byte)(len >> 8);
            
            if (Items != null)
            {
                foreach (var item in Items)
                {
                    item.EncodeToInternal(buffer, ref offset);
                }
            }
        }

        public static PluginListMessage Decode(byte[] data)
        {
            ReadOnlySpan<byte> buffer = data;
            int offset = 0;
            return DecodeFrom(buffer, ref offset);
        }

        internal static PluginListMessage DecodeFrom(ReadOnlySpan<byte> buffer, ref int offset)
        {
            var obj = new PluginListMessage();
            
            int len = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            obj.Items = new Plugin[len];
            
            for (int i = 0; i < len; i++)
            {
                obj.Items[i] = Plugin.DecodeFrom(buffer, ref offset);
            }
            
            return obj;
        }
    }

    // ============================================================================
    // Test harness matching Go benchmark structure (correct measurement!)
    // ============================================================================
    class Program
    {
        static void Main()
        {
            const int iterations = 10000;
            
            // Load real complex.json data
            string jsonPath = "/Users/shaban/Code/ffire/testdata/json/complex.json";
            string json = File.ReadAllText(jsonPath);
            var plugins = JsonSerializer.Deserialize<Plugin[]>(json, new JsonSerializerOptions 
            { 
                PropertyNameCaseInsensitive = true 
            });
            
            var msg = new PluginListMessage { Items = plugins };
            
            Console.WriteLine($"Loaded {plugins.Length} plugins\n");
            
            // Warmup
            byte[] encoded = null;
            PluginListMessage decoded = default;
            for (int i = 0; i < 100; i++)
            {
                encoded = msg.Encode();
                decoded = PluginListMessage.Decode(encoded);
            }
            
            // Benchmark PURE ENCODE (like Go does it!)
            var sw = Stopwatch.StartNew();
            for (int i = 0; i < iterations; i++)
            {
                encoded = msg.Encode();
            }
            sw.Stop();
            long encodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
            
            // Benchmark PURE DECODE
            sw.Restart();
            for (int i = 0; i < iterations; i++)
            {
                decoded = PluginListMessage.Decode(encoded);
            }
            sw.Stop();
            long decodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
            
            Console.WriteLine("=== BLUEPRINT PERFORMANCE ===");
            Console.WriteLine($"  Encode: {encodeNs:N0} ns (pure encode)");
            Console.WriteLine($"  Decode: {decodeNs:N0} ns (pure decode)");
            Console.WriteLine($"  Total:  {encodeNs + decodeNs:N0} ns");
            Console.WriteLine($"  Size:   {encoded.Length} bytes");
            Console.WriteLine();
            Console.WriteLine("Target: < 20,000 ns total");
        }
    }
}
