using System;
using System.IO;
using System.Linq;
using System.Text.Json;
using Blueprint;

class TestProgram
{
    static void Main()
    {
        // Load the same JSON that the benchmark uses
        string jsonPath = "/Users/shaban/Code/ffire/testdata/json/complex.json";
        string json = File.ReadAllText(jsonPath);
        var options = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
        var plugins = JsonSerializer.Deserialize<Plugin[]>(json, options)!;
        
        Console.WriteLine($"Loaded {plugins.Length} plugins");
        
        // Encode with blueprint (wrap in PluginListMessage like the generated code does)
        var msg = new PluginListMessage { Items = plugins };
        byte[] encoded = msg.Encode();
        
        Console.WriteLine($"Blueprint encoded size: {encoded.Length} bytes");
        
        // Write to file for comparison
        File.WriteAllBytes("blueprint_fixture.bin", encoded);
        Console.WriteLine("Written to blueprint_fixture.bin");
        
        // Now compare with generated benchmark fixture
        byte[] generatedFixture = File.ReadAllBytes("/Users/shaban/Code/ffire/benchmarks/generated/ffire_csharp_complex/csharp/fixture.bin");
        Console.WriteLine($"Generated fixture size: {generatedFixture.Length} bytes");
        
        Console.WriteLine($"\nSize difference: {Math.Abs(encoded.Length - generatedFixture.Length)} bytes");
        Console.WriteLine($"Fixtures are {(encoded.SequenceEqual(generatedFixture) ? "IDENTICAL" : "DIFFERENT")}");
    }
}
