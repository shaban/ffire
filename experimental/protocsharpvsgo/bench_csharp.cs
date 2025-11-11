using System;
using System.IO;
using System.Diagnostics;
using System.Text.Json;
using Google.Protobuf;
using Test;

class Program
{
    static void Main()
    {
        // Load JSON fixture
        string jsonContent = File.ReadAllText("complex_wrapped.json");
        
        // Parse JSON to protobuf using JsonParser
        var list = JsonParser.Default.Parse<PluginList>(jsonContent);
        
        // Warmup
        for (int i = 0; i < 1000; i++)
        {
            byte[] bytes = list.ToByteArray();
            PluginList decoded = PluginList.Parser.ParseFrom(bytes);
        }
        
        // Benchmark encode
        int iterations = 100000;
        var sw = Stopwatch.StartNew();
        byte[] encoded = null!;
        for (int i = 0; i < iterations; i++)
        {
            encoded = list.ToByteArray();
        }
        sw.Stop();
        long encodeNs = (sw.Elapsed.Ticks * 100) / iterations;
        
        // Benchmark decode
        sw.Restart();
        for (int i = 0; i < iterations; i++)
        {
            PluginList decoded = PluginList.Parser.ParseFrom(encoded);
        }
        sw.Stop();
        long decodeNs = (sw.Elapsed.Ticks * 100) / iterations;
        
        Console.WriteLine("C# protobuf results:");
        Console.WriteLine($"encode_ns: {encodeNs}");
        Console.WriteLine($"decode_ns: {decodeNs}");
        Console.WriteLine($"total_ns: {encodeNs + decodeNs}");
        Console.WriteLine($"wire_size: {encoded.Length}");
    }
}
