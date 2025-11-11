import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.ArrayList;
import java.util.List;

/**
 * Performance comparison between IntSlice and ArrayList<Integer>
 */
public class PerformanceTest {
    private static final int SIZE = 5000;  // Match array_int benchmark
    private static final int ITERATIONS = 10_000;
    
    public static void main(String[] args) {
        System.out.println("Performance Test: IntSlice vs ArrayList<Integer>");
        System.out.println("Array size: " + SIZE + " elements");
        System.out.println("Iterations: " + ITERATIONS);
        System.out.println();
        
        // Quick warmup
        System.out.println("Warming up (100 iterations)...");
        for (int i = 0; i < 100; i++) {
            testIntSliceEncode();
            testIntSliceDecode();
            testArrayListEncode();
            testArrayListDecode();
        }
        System.out.println("Warmup complete\n");
        
        // Test IntSlice
        long intSliceEncodeTime = testIntSliceEncode();
        long intSliceDecodeTime = testIntSliceDecode();
        
        // Test ArrayList
        long arrayListEncodeTime = testArrayListEncode();
        long arrayListDecodeTime = testArrayListDecode();
        
        // Results
        System.out.println("\n=== RESULTS ===");
        System.out.println("\nIntSlice (primitive array):");
        System.out.printf("  Encode: %,d ns/op\n", intSliceEncodeTime / ITERATIONS);
        System.out.printf("  Decode: %,d ns/op\n", intSliceDecodeTime / ITERATIONS);
        System.out.printf("  Total:  %,d ns/op\n", (intSliceEncodeTime + intSliceDecodeTime) / ITERATIONS);
        
        System.out.println("\nArrayList<Integer> (boxed):");
        System.out.printf("  Encode: %,d ns/op\n", arrayListEncodeTime / ITERATIONS);
        System.out.printf("  Decode: %,d ns/op\n", arrayListDecodeTime / ITERATIONS);
        System.out.printf("  Total:  %,d ns/op\n", (arrayListEncodeTime + arrayListDecodeTime) / ITERATIONS);
        
        System.out.println("\nSpeedup:");
        System.out.printf("  Encode: %.2fx faster\n", (double) arrayListEncodeTime / intSliceEncodeTime);
        System.out.printf("  Decode: %.2fx faster\n", (double) arrayListDecodeTime / intSliceDecodeTime);
        System.out.printf("  Total:  %.2fx faster\n", 
            (double) (arrayListEncodeTime + arrayListDecodeTime) / (intSliceEncodeTime + intSliceDecodeTime));
        
        // Memory test
        System.out.println("\n=== MEMORY TEST ===");
        testMemoryUsage();
    }
    
    private static long testIntSliceEncode() {
        IntSlice slice = new IntSlice(SIZE);
        for (int i = 0; i < SIZE; i++) {
            slice.set(i, i);
        }
        
        ByteBuffer buf = ByteBuffer.allocate(2 + SIZE * 4);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        
        long start = System.nanoTime();
        for (int iter = 0; iter < ITERATIONS; iter++) {
            buf.clear();
            buf.putShort((short) SIZE);
            slice.encodeTo(buf);
        }
        long elapsed = System.nanoTime() - start;
        
        return elapsed;
    }
    
    private static long testIntSliceDecode() {
        // Create test data
        ByteBuffer buf = ByteBuffer.allocate(2 + SIZE * 4);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        buf.putShort((short) SIZE);
        for (int i = 0; i < SIZE; i++) {
            buf.putInt(i);
        }
        
        long start = System.nanoTime();
        for (int iter = 0; iter < ITERATIONS; iter++) {
            buf.position(2);
            IntSlice slice = IntSlice.decodeFrom(buf, SIZE);
        }
        long elapsed = System.nanoTime() - start;
        
        return elapsed;
    }
    
    private static long testArrayListEncode() {
        List<Integer> list = new ArrayList<>(SIZE);
        for (int i = 0; i < SIZE; i++) {
            list.add(i);
        }
        
        ByteBuffer buf = ByteBuffer.allocate(2 + SIZE * 4);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        
        long start = System.nanoTime();
        for (int iter = 0; iter < ITERATIONS; iter++) {
            buf.clear();
            buf.putShort((short) list.size());
            for (Integer value : list) {
                buf.putInt(value);  // Auto-unboxing
            }
        }
        long elapsed = System.nanoTime() - start;
        
        return elapsed;
    }
    
    private static long testArrayListDecode() {
        // Create test data
        ByteBuffer buf = ByteBuffer.allocate(2 + SIZE * 4);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        buf.putShort((short) SIZE);
        for (int i = 0; i < SIZE; i++) {
            buf.putInt(i);
        }
        
        long start = System.nanoTime();
        for (int iter = 0; iter < ITERATIONS; iter++) {
            buf.position(2);
            int len = buf.getShort() & 0xFFFF;
            List<Integer> list = new ArrayList<>(len);
            for (int i = 0; i < len; i++) {
                list.add(buf.getInt());  // Auto-boxing
            }
        }
        long elapsed = System.nanoTime() - start;
        
        return elapsed;
    }
    
    private static void testMemoryUsage() {
        Runtime runtime = Runtime.getRuntime();
        
        // Test IntSlice memory
        runtime.gc();
        long beforeIntSlice = runtime.totalMemory() - runtime.freeMemory();
        IntSlice[] slices = new IntSlice[1000];
        for (int i = 0; i < slices.length; i++) {
            slices[i] = new IntSlice(SIZE);
        }
        long afterIntSlice = runtime.totalMemory() - runtime.freeMemory();
        long intSliceMemory = afterIntSlice - beforeIntSlice;
        
        // Test ArrayList memory
        runtime.gc();
        long beforeArrayList = runtime.totalMemory() - runtime.freeMemory();
        @SuppressWarnings("unchecked")
        List<Integer>[] lists = new ArrayList[1000];
        for (int i = 0; i < lists.length; i++) {
            lists[i] = new ArrayList<>(SIZE);
            for (int j = 0; j < SIZE; j++) {
                lists[i].add(j);
            }
        }
        long afterArrayList = runtime.totalMemory() - runtime.freeMemory();
        long arrayListMemory = afterArrayList - beforeArrayList;
        
        System.out.printf("IntSlice:          %,d bytes (1000 slices × %d ints)\n", intSliceMemory, SIZE);
        System.out.printf("ArrayList<Integer>: %,d bytes (1000 lists × %d Integers)\n", arrayListMemory, SIZE);
        System.out.printf("Memory overhead:   %.2fx more for ArrayList\n", (double) arrayListMemory / intSliceMemory);
    }
}
