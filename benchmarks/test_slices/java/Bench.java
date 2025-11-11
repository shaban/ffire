import test.IntListMessage;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Instant;

public class Bench {
    public static void main(String[] args) throws IOException {
        // Load fixture
        byte[] fixtureData = Files.readAllBytes(Paths.get("fixture.bin"));
        
        int iterations = 100000;
        boolean jsonOutput = "1".equals(System.getenv("BENCH_JSON"));
        
        // Warmup
        for (int i = 0; i < 1000; i++) {
            IntListMessage msg = IntListMessage.decode(fixtureData);
            byte[] encoded = msg.encode();
        }
        
        // Benchmark decode
        long decodeStart = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            IntListMessage msg = IntListMessage.decode(fixtureData);
        }
        long decodeEnd = System.nanoTime();
        long decodeTimeNs = decodeEnd - decodeStart;
        
        // Benchmark encode (decode once, then encode many times)
        IntListMessage msg = IntListMessage.decode(fixtureData);
        long encodeStart = System.nanoTime();
        byte[] encoded = null;
        for (int i = 0; i < iterations; i++) {
            encoded = msg.encode();
        }
        long encodeEnd = System.nanoTime();
        long encodeTimeNs = encodeEnd - encodeStart;
        
        // Calculate metrics
        long encodeNs = encodeTimeNs / iterations;
        long decodeNs = decodeTimeNs / iterations;
        long totalNs = encodeNs + decodeNs;
        
        if (jsonOutput) {
            // Output JSON for automation
            System.out.println("{");
            System.out.println("  \"language\": \"Java\",");
            System.out.println("  \"format\": \"ffire\",");
            System.out.println("  \"message\": \"array_int\",");
            System.out.println("  \"iterations\": " + iterations + ",");
            System.out.println("  \"encode_ns\": " + encodeNs + ",");
            System.out.println("  \"decode_ns\": " + decodeNs + ",");
            System.out.println("  \"total_ns\": " + totalNs + ",");
            System.out.println("  \"wire_size\": " + encoded.length + ",");
            System.out.println("  \"fixture_size\": " + fixtureData.length + ",");
            System.out.println("  \"timestamp\": \"" + Instant.now().toString() + "\"");
            System.out.println("}");
        } else {
            // Print human-readable results
            System.out.println("ffire benchmark: array_int");
            System.out.println("Iterations:  " + iterations);
            System.out.println("Encode:      " + encodeNs + " ns/op");
            System.out.println("Decode:      " + decodeNs + " ns/op");
            System.out.println("Total:       " + totalNs + " ns/op");
            System.out.println("Wire size:   " + encoded.length + " bytes");
            System.out.println("Fixture:     " + fixtureData.length + " bytes");
            double totalTimeS = (encodeTimeNs + decodeTimeNs) / 1e9;
            System.out.printf("Total time:  %.3fs%n", totalTimeS);
        }
    }
}
