using System;
using System.Diagnostics;
using System.IO;

// Copy Parameter and Plugin from Blueprint.cs here for comparison
// This will show if there's an actual performance difference

class Program
{
    static void Main()
    {
        // Load fixture
        byte[] fixture = File.ReadAllBytes("../../csharp-complex-optimizations/blueprint-test/complex.bin");
        
        // Decode into blueprint format
        var blueprintMsg = Blueprint.PluginListMessage.Decode(fixture);
        
        // Decode into generated format
        var generatedMsg = Test.PluginListMessage.Decode(fixture);
        
        const int warmup = 1000;
        const int iterations = 100000;
        
        // Warmup blueprint
        for (int i = 0; i < warmup; i++)
        {
            byte[] _ = blueprintMsg.Encode();
        }
        
        // Measure blueprint encode
        Stopwatch sw1 = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            byte[] _ = blueprintMsg.Encode();
        }
        sw1.Stop();
        long blueprintNs = (long)((double)sw1.ElapsedTicks / Stopwatch.Frequency * 1_000_000_000 / iterations);
        
        // Warmup generated
        for (int i = 0; i < warmup; i++)
        {
            byte[] _ = generatedMsg.Encode();
        }
        
        // Measure generated encode
        Stopwatch sw2 = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            byte[] _ = generatedMsg.Encode();
        }
        sw2.Stop();
        long generatedNs = (long)((double)sw2.ElapsedTicks / Stopwatch.Frequency * 1_000_000_000 / iterations);
        
        Console.WriteLine($"Blueprint encode: {blueprintNs} ns");
        Console.WriteLine($"Generated encode: {generatedNs} ns");
        Console.WriteLine($"Ratio: {(double)generatedNs / blueprintNs:F2}x");
    }
}
