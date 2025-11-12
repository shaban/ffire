```

BenchmarkDotNet v0.15.6, macOS 26.0.1 (25A362) [Darwin 25.0.0]
Apple M1 Pro, 1 CPU, 8 logical and 8 physical cores
.NET SDK 9.0.109
  [Host]     : .NET 9.0.8 (9.0.8, 9.0.825.36511), Arm64 RyuJIT armv8.0-a
  Job-YFEFPZ : .NET 9.0.8 (9.0.8, 9.0.825.36511), Arm64 RyuJIT armv8.0-a

IterationCount=10  WarmupCount=3  

```
| Method | Mean     | Error     | StdDev    | Gen0   | Gen1   | Allocated |
|------- |---------:|----------:|----------:|-------:|-------:|----------:|
| Encode | 4.565 μs | 0.1149 μs | 0.0601 μs | 2.0447 |      - |  12.56 KB |
| Decode | 5.262 μs | 0.4653 μs | 0.2769 μs | 2.4872 | 0.1297 |  15.26 KB |
