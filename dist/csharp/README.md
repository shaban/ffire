# Test - C# Bindings

C# NuGet package with P/Invoke bindings.

## Build

```bash
cd src/Test
dotnet build
```

## Usage

```csharp
using Test;

byte[] data = File.ReadAllBytes("data.bin");
using (var msg = Message.Decode(data)) {
    byte[] encoded = msg.Encode();
}
```
