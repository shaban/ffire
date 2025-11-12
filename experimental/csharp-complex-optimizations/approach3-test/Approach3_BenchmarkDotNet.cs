using System;
using System.IO;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Text.Json;
using System.Buffers.Binary;
using BenchmarkDotNet.Attributes;
using BenchmarkDotNet.Running;

// Data structures
public struct Parameter
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
    public string[] IndexedValues { get; set; }
    public string? IndexedValuesSource { get; set; }
}

public struct Plugin
{
    public string Name { get; set; }
    public string ManufacturerID { get; set; }
    public string Type { get; set; }
    public string Subtype { get; set; }
    public Parameter[] Parameters { get; set; }
}

// APPROACH 3 (WINNER from scratchspace)
static class Approach3
{
    public static unsafe byte[] Encode(Plugin[] plugins)
    {
        int maxSize = 2; // Array length
        foreach (var plugin in plugins)
        {
            maxSize += 2 + (plugin.Name?.Length ?? 0) * 3;
            maxSize += 2 + (plugin.ManufacturerID?.Length ?? 0) * 3;
            maxSize += 2 + (plugin.Type?.Length ?? 0) * 3;
            maxSize += 2 + (plugin.Subtype?.Length ?? 0) * 3;
            maxSize += 2; // Parameter count
            
            if (plugin.Parameters != null)
            {
                foreach (var param in plugin.Parameters)
                {
                    maxSize += 2 + (param.DisplayName?.Length ?? 0) * 3;
                    maxSize += 4 + 4 + 4 + 4 + 4; // 5 floats/ints
                    maxSize += 2 + (param.Unit?.Length ?? 0) * 3;
                    maxSize += 2 + (param.Identifier?.Length ?? 0) * 3;
                    maxSize += 1 + 1; // 2 bools
                    maxSize += 8; // i64
                    maxSize += 1; // Optional array marker
                    if (param.IndexedValues != null)
                    {
                        maxSize += 2;
                        foreach (var val in param.IndexedValues)
                        {
                            maxSize += 2 + (val?.Length ?? 0) * 3;
                        }
                    }
                    maxSize += 1; // Optional string marker
                    if (param.IndexedValuesSource != null)
                    {
                        maxSize += 2 + param.IndexedValuesSource.Length * 3;
                    }
                }
            }
        }
        
        byte[] buffer = new byte[maxSize];
        int offset = 0;
        
        ushort len = (ushort)plugins.Length;
        buffer[offset++] = (byte)len;
        buffer[offset++] = (byte)(len >> 8);
        
        foreach (var plugin in plugins)
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
        
        Array.Resize(ref buffer, offset);
        return buffer;
    }
    
    static unsafe void EncodeParameter(byte[] buffer, ref int offset, Parameter param)
    {
        EncodeString(buffer, ref offset, param.DisplayName);
        
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
        
        int addr = param.Address;
        buffer[offset++] = (byte)addr;
        buffer[offset++] = (byte)(addr >> 8);
        buffer[offset++] = (byte)(addr >> 16);
        buffer[offset++] = (byte)(addr >> 24);
        
        f = param.MaxValue;
        v = *(uint*)&f;
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
        
        EncodeString(buffer, ref offset, param.Unit);
        EncodeString(buffer, ref offset, param.Identifier);
        
        buffer[offset++] = (byte)(param.CanRamp ? 1 : 0);
        buffer[offset++] = (byte)(param.IsWritable ? 1 : 0);
        
        long rawFlags = param.RawFlags;
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
        if (string.IsNullOrEmpty(str))
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

[MemoryDiagnoser]
[SimpleJob(warmupCount: 3, iterationCount: 10)]
public class Approach3Benchmark
{
    private Plugin[] _plugins = null!;
    private byte[] _encoded = null!;

    [GlobalSetup]
    public void Setup()
    {
        string jsonPath = "/Users/shaban/Code/ffire/testdata/json/complex.json";
        string json = File.ReadAllText(jsonPath);
        var options = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
        _plugins = JsonSerializer.Deserialize<Plugin[]>(json, options)!;
        _encoded = Approach3.Encode(_plugins);
        
        Console.WriteLine($"Loaded {_plugins.Length} plugins with {_plugins.Sum(p => p.Parameters?.Length ?? 0)} parameters");
        Console.WriteLine($"Encoded size: {_encoded.Length} bytes");
    }

    [Benchmark]
    public byte[] Encode() => Approach3.Encode(_plugins);

    [Benchmark]
    public Plugin[] Decode() => Approach3.Decode(_encoded);
}

class Program
{
    static void Main(string[] args)
    {
        BenchmarkRunner.Run<Approach3Benchmark>();
    }
}
