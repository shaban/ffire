using System;
using System.Diagnostics;
using System.IO;
using Google.Protobuf;
using Test;

class Program
{
    const int Iterations = 100_000;
    const int Warmup = 1000;

    static void Main(string[] args)
    {
        // Load JSON fixture
        string jsonPath = "../../testdata/json/complex.json";
        if (!File.Exists(jsonPath))
        {
            Console.WriteLine($"Error: {jsonPath} not found");
            return;
        }

        string json = File.ReadAllText(jsonPath);
        
        // Parse JSON to protobuf message
        var pluginList = new PluginList();
        var parser = new JsonParser(JsonParser.Settings.Default);
        
        // Parse JSON array into repeated field
        var jsonArray = System.Text.Json.JsonSerializer.Deserialize<System.Text.Json.JsonElement>(json);
        foreach (var item in jsonArray.EnumerateArray())
        {
            var plugin = parser.Parse<Plugin>(item.GetRawText());
            pluginList.Plugins.Add(plugin);
        }

        Console.WriteLine($"Loaded {pluginList.Plugins.Count} plugins");
        Console.WriteLine($"Running {Iterations:N0} iterations...\n");

        // Serialize once to get wire size
        byte[] encoded = pluginList.ToByteArray();
        Console.WriteLine($"Wire size: {encoded.Length} bytes\n");

        // Warmup
        for (int i = 0; i < Warmup; i++)
        {
            byte[] data = pluginList.ToByteArray();
            var decoded = PluginList.Parser.ParseFrom(data);
        }

        // Benchmark encode
        var sw = Stopwatch.StartNew();
        for (int i = 0; i < Iterations; i++)
        {
            byte[] data = pluginList.ToByteArray();
        }
        sw.Stop();
        long encodeNs = (sw.ElapsedTicks * 1_000_000_000) / (Stopwatch.Frequency * Iterations);
        Console.WriteLine($"Encode: {encodeNs} ns/op");

        // Benchmark decode
        sw.Restart();
        for (int i = 0; i < Iterations; i++)
        {
            var decoded = PluginList.Parser.ParseFrom(encoded);
        }
        sw.Stop();
        long decodeNs = (sw.ElapsedTicks * 1_000_000_000) / (Stopwatch.Frequency * Iterations);
        Console.WriteLine($"Decode: {decodeNs} ns/op");
        Console.WriteLine($"Total:  {encodeNs + decodeNs} ns/op");

        // Verify correctness
        var verify = PluginList.Parser.ParseFrom(encoded);
        if (verify.Plugins.Count == pluginList.Plugins.Count)
        {
            Console.WriteLine($"\n✓ Correctness verified: {verify.Plugins.Count} plugins");
        }
    }
}
