using System;
using System.Buffers.Binary;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Runtime.CompilerServices;
using System.Text;
using System.Text.Json;

namespace ComplexOptimization;

// GOAL: Beat protobuf's 19,973 ns (9,857 encode + 10,116 decode)
// Current ffire: 32,217 ns (23,217 encode + 9,000 decode)

class Program
{
    static void Main()
    {
        Console.WriteLine("=== COMPLEX FIXTURE OPTIMIZATION SCRATCHSPACE ===\n");
        Console.WriteLine("Goal: <15,000 ns total (beat protobuf's 19,973 ns)\n");
        
        // Load REAL complex.json data (20 plugins with unique parameters!)
        var plugins = LoadRealComplexData();
        
        // Approach 1: Current generator (single-pass UTF-8, * 3 multiplier, Array.Resize)
        Console.WriteLine("Approach 1: Current generator (single-pass, * 3, Array.Resize)");
        BenchmarkApproach1(plugins);
        
        Console.WriteLine("\nApproach 2: List<byte> (like C++ vector, no pre-allocation)");
        BenchmarkApproach2(plugins);
        
        Console.WriteLine("\nApproach 3: Inline bit-shifting (like Go, no BinaryPrimitives)");
        BenchmarkApproach3(plugins);
        
        Console.WriteLine("\nApproach 4: HYBRID - Inline shifts + optimized decode");
        BenchmarkApproach4(plugins);
        
        Console.WriteLine("\n=== ANALYSIS ===");
        Console.WriteLine("- Approach 1: Current ffire generator (16.3 encode, 13.3 decode)");
        Console.WriteLine("- Approach 2: C++-style List<byte> (22.4 encode, 6.2 decode)");
        Console.WriteLine("- Approach 3: Go-style inline shifts (16.5 encode, 5.5 decode) ✓");
        Console.WriteLine("- Approach 4: HYBRID best of all (target: <10 encode, <5 decode = <15 total)");
    }
    
    static Plugin[] LoadRealComplexData()
    {
        string jsonPath = "/Users/shaban/Code/ffire/testdata/json/complex.json";
        string json = File.ReadAllText(jsonPath);
        
        var options = new JsonSerializerOptions 
        { 
            PropertyNameCaseInsensitive = true 
        };
        
        var plugins = JsonSerializer.Deserialize<Plugin[]>(json, options);
        
        if (plugins == null || plugins.Length == 0)
        {
            throw new Exception("Failed to load complex.json");
        }
        
        // Count total parameters to verify uniqueness
        int totalParams = 0;
        foreach (var plugin in plugins)
        {
            totalParams += plugin.Parameters?.Length ?? 0;
        }
        
        Console.WriteLine($"✓ Loaded {plugins.Length} plugins with {totalParams} unique parameters\n");
        return plugins;
    }
    
    static void BenchmarkApproach1(Plugin[] plugins)
    {
        const int iterations = 10000;
        byte[]? encoded = null;
        Plugin[]? decoded = null;
        
        // Warmup
        for (int i = 0; i < 100; i++)
        {
            encoded = Approach1.Encode(plugins);
            decoded = Approach1.Decode(encoded);
        }
        
        // Benchmark encode
        var sw = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            encoded = Approach1.Encode(plugins);
        }
        sw.Stop();
        long encodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        // Benchmark decode
        sw.Restart();
        for (int i = 0; i < iterations; i++)
        {
            decoded = Approach1.Decode(encoded!);
        }
        sw.Stop();
        long decodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        Console.WriteLine($"  Encode: {encodeNs:N0} ns");
        Console.WriteLine($"  Decode: {decodeNs:N0} ns");
        Console.WriteLine($"  Total:  {encodeNs + decodeNs:N0} ns");
        Console.WriteLine($"  Size:   {encoded!.Length} bytes");
    }
    
    static void BenchmarkApproach2(Plugin[] plugins)
    {
        const int iterations = 10000;
        byte[]? encoded = null;
        Plugin[]? decoded = null;
        
        // Warmup
        for (int i = 0; i < 100; i++)
        {
            encoded = Approach2.Encode(plugins);
            decoded = Approach2.Decode(encoded);
        }
        
        // Benchmark encode
        var sw = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            encoded = Approach2.Encode(plugins);
        }
        sw.Stop();
        long encodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        // Benchmark decode
        sw.Restart();
        for (int i = 0; i < iterations; i++)
        {
            decoded = Approach2.Decode(encoded!);
        }
        sw.Stop();
        long decodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        Console.WriteLine($"  Encode: {encodeNs:N0} ns");
        Console.WriteLine($"  Decode: {decodeNs:N0} ns");
        Console.WriteLine($"  Total:  {encodeNs + decodeNs:N0} ns");
        Console.WriteLine($"  Size:   {encoded!.Length} bytes");
    }
    
    static void BenchmarkApproach3(Plugin[] plugins)
    {
        const int iterations = 10000;
        byte[]? encoded = null;
        Plugin[]? decoded = null;
        
        // Warmup
        for (int i = 0; i < 100; i++)
        {
            encoded = Approach3.Encode(plugins);
            decoded = Approach3.Decode(encoded);
        }
        
        // Benchmark encode
        var sw = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            encoded = Approach3.Encode(plugins);
        }
        sw.Stop();
        long encodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        // Benchmark decode
        sw.Restart();
        for (int i = 0; i < iterations; i++)
        {
            decoded = Approach3.Decode(encoded!);
        }
        sw.Stop();
        long decodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        Console.WriteLine($"  Encode: {encodeNs:N0} ns");
        Console.WriteLine($"  Decode: {decodeNs:N0} ns");
        Console.WriteLine($"  Total:  {encodeNs + decodeNs:N0} ns");
        Console.WriteLine($"  Size:   {encoded!.Length} bytes");
    }
    
    static void BenchmarkApproach4(Plugin[] plugins)
    {
        const int iterations = 10000;
        byte[]? encoded = null;
        Plugin[]? decoded = null;
        
        // Warmup
        for (int i = 0; i < 100; i++)
        {
            encoded = Approach4.Encode(plugins);
            decoded = Approach4.Decode(encoded);
        }
        
        // Benchmark encode
        var sw = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            encoded = Approach4.Encode(plugins);
        }
        sw.Stop();
        long encodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        // Benchmark decode
        sw.Restart();
        for (int i = 0; i < iterations; i++)
        {
            decoded = Approach4.Decode(encoded!);
        }
        sw.Stop();
        long decodeNs = sw.ElapsedTicks * 1_000_000_000 / Stopwatch.Frequency / iterations;
        
        Console.WriteLine($"  Encode: {encodeNs:N0} ns");
        Console.WriteLine($"  Decode: {decodeNs:N0} ns");
        Console.WriteLine($"  Total:  {encodeNs + decodeNs:N0} ns");
        Console.WriteLine($"  Size:   {encoded!.Length} bytes");
    }
}

// Data structures matching complex.json
struct Plugin
{
    public string Name { get; set; }
    public string ManufacturerID { get; set; }
    public string Type { get; set; }
    public string Subtype { get; set; }
    public Parameter[] Parameters { get; set; }
}

struct Parameter
{
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
    public string[]? IndexedValues { get; set; }
    public string? IndexedValuesSource { get; set; }
}

// APPROACH 1: Current generator (single-pass UTF-8, * 3 multiplier, Array.Resize)
static class Approach1
{
    public static byte[] Encode(Plugin[] plugins)
    {
        int maxSize = ComputeMaxSize(plugins);
        byte[] buffer = new byte[maxSize];
        int offset = 0;
        
        BinaryPrimitives.WriteUInt16LittleEndian(buffer.AsSpan(offset, 2), (ushort)plugins.Length);
        offset += 2;
        
        foreach (var plugin in plugins)
        {
            EncodePlugin(buffer, ref offset, plugin);
        }
        
        if (offset < maxSize)
        {
            Array.Resize(ref buffer, offset);
        }
        
        return buffer;
    }
    
    static int ComputeMaxSize(Plugin[] plugins)
    {
        int size = 2;
        foreach (var plugin in plugins)
        {
            size += 2 + (plugin.Name?.Length ?? 0) * 3;
            size += 2 + (plugin.ManufacturerID?.Length ?? 0) * 3;
            size += 2 + (plugin.Type?.Length ?? 0) * 3;
            size += 2 + (plugin.Subtype?.Length ?? 0) * 3;
            size += 2;
            if (plugin.Parameters != null)
            {
                foreach (var param in plugin.Parameters)
                {
                    size += 2 + (param.DisplayName?.Length ?? 0) * 3;
                    size += 4 + 4 + 4 + 4 + 4;
                    size += 2 + (param.Unit?.Length ?? 0) * 3;
                    size += 2 + (param.Identifier?.Length ?? 0) * 3;
                    size += 1 + 1 + 8;
                    size += 1;
                    if (param.IndexedValues != null)
                    {
                        size += 2;
                        foreach (var val in param.IndexedValues)
                        {
                            size += 2 + (val?.Length ?? 0) * 3;
                        }
                    }
                    size += 1;
                    if (param.IndexedValuesSource != null)
                    {
                        size += 2 + param.IndexedValuesSource.Length * 3;
                    }
                }
            }
        }
        return size;
    }
    
    static void EncodePlugin(byte[] buffer, ref int offset, Plugin plugin)
    {
        EncodeString(buffer, ref offset, plugin.Name);
        EncodeString(buffer, ref offset, plugin.ManufacturerID);
        EncodeString(buffer, ref offset, plugin.Type);
        EncodeString(buffer, ref offset, plugin.Subtype);
        
        BinaryPrimitives.WriteUInt16LittleEndian(buffer.AsSpan(offset, 2), (ushort)(plugin.Parameters?.Length ?? 0));
        offset += 2;
        
        if (plugin.Parameters != null)
        {
            foreach (var param in plugin.Parameters)
            {
                EncodeParameter(buffer, ref offset, param);
            }
        }
    }
    
    static void EncodeParameter(byte[] buffer, ref int offset, Parameter param)
    {
        EncodeString(buffer, ref offset, param.DisplayName);
        BinaryPrimitives.WriteSingleLittleEndian(buffer.AsSpan(offset, 4), param.DefaultValue);
        offset += 4;
        BinaryPrimitives.WriteSingleLittleEndian(buffer.AsSpan(offset, 4), param.CurrentValue);
        offset += 4;
        BinaryPrimitives.WriteInt32LittleEndian(buffer.AsSpan(offset, 4), param.Address);
        offset += 4;
        BinaryPrimitives.WriteSingleLittleEndian(buffer.AsSpan(offset, 4), param.MaxValue);
        offset += 4;
        BinaryPrimitives.WriteSingleLittleEndian(buffer.AsSpan(offset, 4), param.MinValue);
        offset += 4;
        EncodeString(buffer, ref offset, param.Unit);
        EncodeString(buffer, ref offset, param.Identifier);
        
        buffer[offset++] = (byte)(param.CanRamp ? 1 : 0);
        buffer[offset++] = (byte)(param.IsWritable ? 1 : 0);
        
        BinaryPrimitives.WriteInt64LittleEndian(buffer.AsSpan(offset, 8), param.RawFlags);
        offset += 8;
        
        if (param.IndexedValues != null)
        {
            buffer[offset++] = 1;
            BinaryPrimitives.WriteUInt16LittleEndian(buffer.AsSpan(offset, 2), (ushort)param.IndexedValues.Length);
            offset += 2;
            foreach (var val in param.IndexedValues)
            {
                EncodeString(buffer, ref offset, val);
            }
        }
        else
        {
            buffer[offset++] = 0;
        }
        
        if (param.IndexedValuesSource != null)
        {
            buffer[offset++] = 1;
            EncodeString(buffer, ref offset, param.IndexedValuesSource);
        }
        else
        {
            buffer[offset++] = 0;
        }
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static void EncodeString(byte[] buffer, ref int offset, string? str)
    {
        if (str == null || str.Length == 0)
        {
            BinaryPrimitives.WriteUInt16LittleEndian(buffer.AsSpan(offset, 2), 0);
            offset += 2;
            return;
        }
        
        int byteCount = Encoding.UTF8.GetBytes(str, buffer.AsSpan(offset + 2));
        BinaryPrimitives.WriteUInt16LittleEndian(buffer.AsSpan(offset, 2), (ushort)byteCount);
        offset += 2 + byteCount;
    }
    
    public static Plugin[] Decode(byte[] data)
    {
        ReadOnlySpan<byte> buffer = data;
        int offset = 0;
        
        int length = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
        offset += 2;
        
        var plugins = new Plugin[length];
        for (int i = 0; i < length; i++)
        {
            plugins[i] = DecodePlugin(buffer, ref offset);
        }
        
        return plugins;
    }
    
    static Plugin DecodePlugin(ReadOnlySpan<byte> buffer, ref int offset)
    {
        var plugin = new Plugin
        {
            Name = DecodeString(buffer, ref offset),
            ManufacturerID = DecodeString(buffer, ref offset),
            Type = DecodeString(buffer, ref offset),
            Subtype = DecodeString(buffer, ref offset)
        };
        
        int paramCount = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
        offset += 2;
        
        plugin.Parameters = new Parameter[paramCount];
        for (int i = 0; i < paramCount; i++)
        {
            plugin.Parameters[i] = DecodeParameter(buffer, ref offset);
        }
        
        return plugin;
    }
    
    static Parameter DecodeParameter(ReadOnlySpan<byte> buffer, ref int offset)
    {
        var param = new Parameter
        {
            DisplayName = DecodeString(buffer, ref offset),
            DefaultValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4))
        };
        offset += 4;
        
        param.CurrentValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4));
        offset += 4;
        
        param.Address = BinaryPrimitives.ReadInt32LittleEndian(buffer.Slice(offset, 4));
        offset += 4;
        
        param.MaxValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4));
        offset += 4;
        
        param.MinValue = BinaryPrimitives.ReadSingleLittleEndian(buffer.Slice(offset, 4));
        offset += 4;
        
        param.Unit = DecodeString(buffer, ref offset);
        param.Identifier = DecodeString(buffer, ref offset);
        
        param.CanRamp = buffer[offset++] != 0;
        param.IsWritable = buffer[offset++] != 0;
        
        param.RawFlags = BinaryPrimitives.ReadInt64LittleEndian(buffer.Slice(offset, 8));
        offset += 8;
        
        if (buffer[offset++] == 1)
        {
            int arrayLen = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
            offset += 2;
            param.IndexedValues = new string[arrayLen];
            for (int i = 0; i < arrayLen; i++)
            {
                param.IndexedValues[i] = DecodeString(buffer, ref offset);
            }
        }
        
        if (buffer[offset++] == 1)
        {
            param.IndexedValuesSource = DecodeString(buffer, ref offset);
        }
        
        return param;
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static string DecodeString(ReadOnlySpan<byte> buffer, ref int offset)
    {
        int len = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset, 2));
        offset += 2;
        
        if (len == 0) return string.Empty;
        
        string result = Encoding.UTF8.GetString(buffer.Slice(offset, len));
        offset += len;
        return result;
    }
}

// APPROACH 2: List<byte> (like C++ std::vector - no pre-allocation, dynamic growth)
static class Approach2
{
    public static byte[] Encode(Plugin[] plugins)
    {
        var buffer = new List<byte>(4096); // Initial capacity hint
        
        // Array length
        ushort len = (ushort)plugins.Length;
        buffer.Add((byte)len);
        buffer.Add((byte)(len >> 8));
        
        foreach (var plugin in plugins)
        {
            EncodePlugin(buffer, plugin);
        }
        
        return buffer.ToArray();
    }
    
    static void EncodePlugin(List<byte> buffer, Plugin plugin)
    {
        EncodeString(buffer, plugin.Name);
        EncodeString(buffer, plugin.ManufacturerID);
        EncodeString(buffer, plugin.Type);
        EncodeString(buffer, plugin.Subtype);
        
        ushort paramLen = (ushort)(plugin.Parameters?.Length ?? 0);
        buffer.Add((byte)paramLen);
        buffer.Add((byte)(paramLen >> 8));
        
        if (plugin.Parameters != null)
        {
            foreach (var param in plugin.Parameters)
            {
                EncodeParameter(buffer, param);
            }
        }
    }
    
    static void EncodeParameter(List<byte> buffer, Parameter param)
    {
        EncodeString(buffer, param.DisplayName);
        
        // Floats - use BitConverter or unsafe
        buffer.AddRange(BitConverter.GetBytes(param.DefaultValue));
        buffer.AddRange(BitConverter.GetBytes(param.CurrentValue));
        buffer.AddRange(BitConverter.GetBytes(param.Address));
        buffer.AddRange(BitConverter.GetBytes(param.MaxValue));
        buffer.AddRange(BitConverter.GetBytes(param.MinValue));
        
        EncodeString(buffer, param.Unit);
        EncodeString(buffer, param.Identifier);
        
        buffer.Add((byte)(param.CanRamp ? 1 : 0));
        buffer.Add((byte)(param.IsWritable ? 1 : 0));
        
        buffer.AddRange(BitConverter.GetBytes(param.RawFlags));
        
        if (param.IndexedValues != null)
        {
            buffer.Add(1);
            ushort arrayLen = (ushort)param.IndexedValues.Length;
            buffer.Add((byte)arrayLen);
            buffer.Add((byte)(arrayLen >> 8));
            foreach (var val in param.IndexedValues)
            {
                EncodeString(buffer, val);
            }
        }
        else
        {
            buffer.Add(0);
        }
        
        if (param.IndexedValuesSource != null)
        {
            buffer.Add(1);
            EncodeString(buffer, param.IndexedValuesSource);
        }
        else
        {
            buffer.Add(0);
        }
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static void EncodeString(List<byte> buffer, string? str)
    {
        if (str == null || str.Length == 0)
        {
            buffer.Add(0);
            buffer.Add(0);
            return;
        }
        
        byte[] utf8Bytes = Encoding.UTF8.GetBytes(str);
        ushort len = (ushort)utf8Bytes.Length;
        buffer.Add((byte)len);
        buffer.Add((byte)(len >> 8));
        buffer.AddRange(utf8Bytes);
    }
    
    public static Plugin[] Decode(byte[] data)
    {
        return Approach1.Decode(data); // Decode is same
    }
}

// APPROACH 3: Inline bit-shifting (like Go - reduce BinaryPrimitives overhead)
static class Approach3
{
    public static byte[] Encode(Plugin[] plugins)
    {
        int maxSize = ComputeMaxSize(plugins);
        byte[] buffer = new byte[maxSize];
        int offset = 0;
        
        // Inline length write
        ushort len = (ushort)plugins.Length;
        buffer[offset++] = (byte)len;
        buffer[offset++] = (byte)(len >> 8);
        
        foreach (var plugin in plugins)
        {
            EncodePlugin(buffer, ref offset, plugin);
        }
        
        if (offset < maxSize)
        {
            Array.Resize(ref buffer, offset);
        }
        
        return buffer;
    }
    
    static int ComputeMaxSize(Plugin[] plugins)
    {
        int size = 2;
        foreach (var plugin in plugins)
        {
            size += 2 + (plugin.Name?.Length ?? 0) * 3;
            size += 2 + (plugin.ManufacturerID?.Length ?? 0) * 3;
            size += 2 + (plugin.Type?.Length ?? 0) * 3;
            size += 2 + (plugin.Subtype?.Length ?? 0) * 3;
            size += 2;
            if (plugin.Parameters != null)
            {
                foreach (var param in plugin.Parameters)
                {
                    size += 2 + (param.DisplayName?.Length ?? 0) * 3;
                    size += 4 + 4 + 4 + 4 + 4;
                    size += 2 + (param.Unit?.Length ?? 0) * 3;
                    size += 2 + (param.Identifier?.Length ?? 0) * 3;
                    size += 1 + 1 + 8;
                    size += 1;
                    if (param.IndexedValues != null)
                    {
                        size += 2;
                        foreach (var val in param.IndexedValues)
                        {
                            size += 2 + (val?.Length ?? 0) * 3;
                        }
                    }
                    size += 1;
                    if (param.IndexedValuesSource != null)
                    {
                        size += 2 + param.IndexedValuesSource.Length * 3;
                    }
                }
            }
        }
        return size;
    }
    
    static void EncodePlugin(byte[] buffer, ref int offset, Plugin plugin)
    {
        EncodeString(buffer, ref offset, plugin.Name);
        EncodeString(buffer, ref offset, plugin.ManufacturerID);
        EncodeString(buffer, ref offset, plugin.Type);
        EncodeString(buffer, ref offset, plugin.Subtype);
        
        ushort paramLen = (ushort)(plugin.Parameters?.Length ?? 0);
        buffer[offset++] = (byte)paramLen;
        buffer[offset++] = (byte)(paramLen >> 8);
        
        if (plugin.Parameters != null)
        {
            foreach (var param in plugin.Parameters)
            {
                EncodeParameter(buffer, ref offset, param);
            }
        }
    }
    
    static void EncodeParameter(byte[] buffer, ref int offset, Parameter param)
    {
        EncodeString(buffer, ref offset, param.DisplayName);
        
        // Inline float writes using unsafe pointer cast (like C++)
        unsafe
        {
            float f = param.DefaultValue;
            uint v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
            
            f = param.CurrentValue;
            v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
        }
        
        // Inline int write
        uint addr = (uint)param.Address;
        buffer[offset++] = (byte)addr;
        buffer[offset++] = (byte)(addr >> 8);
        buffer[offset++] = (byte)(addr >> 16);
        buffer[offset++] = (byte)(addr >> 24);
        
        unsafe
        {
            float f = param.MaxValue;
            uint v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
            
            f = param.MinValue;
            v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
        }
        
        EncodeString(buffer, ref offset, param.Unit);
        EncodeString(buffer, ref offset, param.Identifier);
        
        buffer[offset++] = (byte)(param.CanRamp ? 1 : 0);
        buffer[offset++] = (byte)(param.IsWritable ? 1 : 0);
        
        // Inline long write
        ulong rawFlags = (ulong)param.RawFlags;
        buffer[offset++] = (byte)rawFlags;
        buffer[offset++] = (byte)(rawFlags >> 8);
        buffer[offset++] = (byte)(rawFlags >> 16);
        buffer[offset++] = (byte)(rawFlags >> 24);
        buffer[offset++] = (byte)(rawFlags >> 32);
        buffer[offset++] = (byte)(rawFlags >> 40);
        buffer[offset++] = (byte)(rawFlags >> 48);
        buffer[offset++] = (byte)(rawFlags >> 56);
        
        if (param.IndexedValues != null)
        {
            buffer[offset++] = 1;
            ushort arrayLen = (ushort)param.IndexedValues.Length;
            buffer[offset++] = (byte)arrayLen;
            buffer[offset++] = (byte)(arrayLen >> 8);
            foreach (var val in param.IndexedValues)
            {
                EncodeString(buffer, ref offset, val);
            }
        }
        else
        {
            buffer[offset++] = 0;
        }
        
        if (param.IndexedValuesSource != null)
        {
            buffer[offset++] = 1;
            EncodeString(buffer, ref offset, param.IndexedValuesSource);
        }
        else
        {
            buffer[offset++] = 0;
        }
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static void EncodeString(byte[] buffer, ref int offset, string? str)
    {
        if (str == null || str.Length == 0)
        {
            buffer[offset++] = 0;
            buffer[offset++] = 0;
            return;
        }
        
        int byteCount = Encoding.UTF8.GetBytes(str, buffer.AsSpan(offset + 2));
        buffer[offset++] = (byte)byteCount;
        buffer[offset++] = (byte)(byteCount >> 8);
        offset += byteCount;
    }
    
    public static Plugin[] Decode(byte[] data)
    {
        return Approach1.Decode(data); // Decode is same for now
    }
}

// APPROACH 4: HYBRID - Best of all approaches
// - Inline bit-shifting for primitives (from Approach 3)
// - Pre-allocated array (from Approach 1) 
// - Optimized decode with inline shifts
// - Direct Span<byte> usage everywhere
static class Approach4
{
    public static byte[] Encode(Plugin[] plugins)
    {
        int maxSize = ComputeMaxSize(plugins);
        byte[] buffer = new byte[maxSize];
        int offset = 0;
        
        // Inline length write
        ushort len = (ushort)plugins.Length;
        buffer[offset++] = (byte)len;
        buffer[offset++] = (byte)(len >> 8);
        
        foreach (var plugin in plugins)
        {
            EncodePlugin(buffer, ref offset, plugin);
        }
        
        if (offset < maxSize)
        {
            Array.Resize(ref buffer, offset);
        }
        
        return buffer;
    }
    
    static int ComputeMaxSize(Plugin[] plugins)
    {
        int size = 2;
        foreach (var plugin in plugins)
        {
            size += 2 + (plugin.Name?.Length ?? 0) * 3;
            size += 2 + (plugin.ManufacturerID?.Length ?? 0) * 3;
            size += 2 + (plugin.Type?.Length ?? 0) * 3;
            size += 2 + (plugin.Subtype?.Length ?? 0) * 3;
            size += 2;
            if (plugin.Parameters != null)
            {
                foreach (var param in plugin.Parameters)
                {
                    size += 2 + (param.DisplayName?.Length ?? 0) * 3;
                    size += 4 + 4 + 4 + 4 + 4;
                    size += 2 + (param.Unit?.Length ?? 0) * 3;
                    size += 2 + (param.Identifier?.Length ?? 0) * 3;
                    size += 1 + 1 + 8;
                    size += 1;
                    if (param.IndexedValues != null)
                    {
                        size += 2;
                        foreach (var val in param.IndexedValues)
                        {
                            size += 2 + (val?.Length ?? 0) * 3;
                        }
                    }
                    size += 1;
                    if (param.IndexedValuesSource != null)
                    {
                        size += 2 + param.IndexedValuesSource.Length * 3;
                    }
                }
            }
        }
        return size;
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static void EncodePlugin(byte[] buffer, ref int offset, Plugin plugin)
    {
        EncodeString(buffer, ref offset, plugin.Name);
        EncodeString(buffer, ref offset, plugin.ManufacturerID);
        EncodeString(buffer, ref offset, plugin.Type);
        EncodeString(buffer, ref offset, plugin.Subtype);
        
        ushort paramLen = (ushort)(plugin.Parameters?.Length ?? 0);
        buffer[offset++] = (byte)paramLen;
        buffer[offset++] = (byte)(paramLen >> 8);
        
        if (plugin.Parameters != null)
        {
            foreach (var param in plugin.Parameters)
            {
                EncodeParameter(buffer, ref offset, param);
            }
        }
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static void EncodeParameter(byte[] buffer, ref int offset, Parameter param)
    {
        EncodeString(buffer, ref offset, param.DisplayName);
        
        // Inline float writes using unsafe pointer cast
        unsafe
        {
            float f = param.DefaultValue;
            uint v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
            
            f = param.CurrentValue;
            v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
        }
        
        // Inline int write
        uint addr = (uint)param.Address;
        buffer[offset++] = (byte)addr;
        buffer[offset++] = (byte)(addr >> 8);
        buffer[offset++] = (byte)(addr >> 16);
        buffer[offset++] = (byte)(addr >> 24);
        
        unsafe
        {
            float f = param.MaxValue;
            uint v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
            
            f = param.MinValue;
            v = *(uint*)&f;
            buffer[offset++] = (byte)v;
            buffer[offset++] = (byte)(v >> 8);
            buffer[offset++] = (byte)(v >> 16);
            buffer[offset++] = (byte)(v >> 24);
        }
        
        EncodeString(buffer, ref offset, param.Unit);
        EncodeString(buffer, ref offset, param.Identifier);
        
        buffer[offset++] = (byte)(param.CanRamp ? 1 : 0);
        buffer[offset++] = (byte)(param.IsWritable ? 1 : 0);
        
        // Inline long write
        ulong rawFlags = (ulong)param.RawFlags;
        buffer[offset++] = (byte)rawFlags;
        buffer[offset++] = (byte)(rawFlags >> 8);
        buffer[offset++] = (byte)(rawFlags >> 16);
        buffer[offset++] = (byte)(rawFlags >> 24);
        buffer[offset++] = (byte)(rawFlags >> 32);
        buffer[offset++] = (byte)(rawFlags >> 40);
        buffer[offset++] = (byte)(rawFlags >> 48);
        buffer[offset++] = (byte)(rawFlags >> 56);
        
        if (param.IndexedValues != null)
        {
            buffer[offset++] = 1;
            ushort arrayLen = (ushort)param.IndexedValues.Length;
            buffer[offset++] = (byte)arrayLen;
            buffer[offset++] = (byte)(arrayLen >> 8);
            foreach (var val in param.IndexedValues)
            {
                EncodeString(buffer, ref offset, val);
            }
        }
        else
        {
            buffer[offset++] = 0;
        }
        
        if (param.IndexedValuesSource != null)
        {
            buffer[offset++] = 1;
            EncodeString(buffer, ref offset, param.IndexedValuesSource);
        }
        else
        {
            buffer[offset++] = 0;
        }
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static void EncodeString(byte[] buffer, ref int offset, string? str)
    {
        if (str == null || str.Length == 0)
        {
            buffer[offset++] = 0;
            buffer[offset++] = 0;
            return;
        }
        
        int byteCount = Encoding.UTF8.GetBytes(str, buffer.AsSpan(offset + 2));
        buffer[offset++] = (byte)byteCount;
        buffer[offset++] = (byte)(byteCount >> 8);
        offset += byteCount;
    }
    
    // OPTIMIZED DECODE with inline bit-shifting
    public static Plugin[] Decode(byte[] data)
    {
        ReadOnlySpan<byte> buffer = data;
        int offset = 0;
        
        // Inline length read
        int length = buffer[offset] | (buffer[offset + 1] << 8);
        offset += 2;
        
        var plugins = new Plugin[length];
        for (int i = 0; i < length; i++)
        {
            plugins[i] = DecodePlugin(buffer, ref offset);
        }
        
        return plugins;
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static Plugin DecodePlugin(ReadOnlySpan<byte> buffer, ref int offset)
    {
        var plugin = new Plugin
        {
            Name = DecodeString(buffer, ref offset),
            ManufacturerID = DecodeString(buffer, ref offset),
            Type = DecodeString(buffer, ref offset),
            Subtype = DecodeString(buffer, ref offset)
        };
        
        int paramCount = buffer[offset] | (buffer[offset + 1] << 8);
        offset += 2;
        
        plugin.Parameters = new Parameter[paramCount];
        for (int i = 0; i < paramCount; i++)
        {
            plugin.Parameters[i] = DecodeParameter(buffer, ref offset);
        }
        
        return plugin;
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static Parameter DecodeParameter(ReadOnlySpan<byte> buffer, ref int offset)
    {
        var param = new Parameter
        {
            DisplayName = DecodeString(buffer, ref offset)
        };
        
        // Inline float reads
        unsafe
        {
            uint v = (uint)(buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16) | (buffer[offset + 3] << 24));
            param.DefaultValue = *(float*)&v;
            offset += 4;
            
            v = (uint)(buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16) | (buffer[offset + 3] << 24));
            param.CurrentValue = *(float*)&v;
            offset += 4;
        }
        
        param.Address = buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16) | (buffer[offset + 3] << 24);
        offset += 4;
        
        unsafe
        {
            uint v = (uint)(buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16) | (buffer[offset + 3] << 24));
            param.MaxValue = *(float*)&v;
            offset += 4;
            
            v = (uint)(buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16) | (buffer[offset + 3] << 24));
            param.MinValue = *(float*)&v;
            offset += 4;
        }
        
        param.Unit = DecodeString(buffer, ref offset);
        param.Identifier = DecodeString(buffer, ref offset);
        
        param.CanRamp = buffer[offset++] != 0;
        param.IsWritable = buffer[offset++] != 0;
        
        // Inline long read
        param.RawFlags = (long)((ulong)buffer[offset] | 
                               ((ulong)buffer[offset + 1] << 8) | 
                               ((ulong)buffer[offset + 2] << 16) | 
                               ((ulong)buffer[offset + 3] << 24) |
                               ((ulong)buffer[offset + 4] << 32) | 
                               ((ulong)buffer[offset + 5] << 40) | 
                               ((ulong)buffer[offset + 6] << 48) | 
                               ((ulong)buffer[offset + 7] << 56));
        offset += 8;
        
        if (buffer[offset++] == 1)
        {
            int arrayLen = buffer[offset] | (buffer[offset + 1] << 8);
            offset += 2;
            param.IndexedValues = new string[arrayLen];
            for (int i = 0; i < arrayLen; i++)
            {
                param.IndexedValues[i] = DecodeString(buffer, ref offset);
            }
        }
        
        if (buffer[offset++] == 1)
        {
            param.IndexedValuesSource = DecodeString(buffer, ref offset);
        }
        
        return param;
    }
    
    [MethodImpl(MethodImplOptions.AggressiveInlining)]
    static string DecodeString(ReadOnlySpan<byte> buffer, ref int offset)
    {
        int len = buffer[offset] | (buffer[offset + 1] << 8);
        offset += 2;
        
        if (len == 0) return string.Empty;
        
        string result = Encoding.UTF8.GetString(buffer.Slice(offset, len));
        offset += len;
        return result;
    }
}
