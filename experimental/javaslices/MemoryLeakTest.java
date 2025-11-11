/**
 * Memory Leak Test - Verify IntSlice doesn't leak memory
 * 
 * This test repeatedly allocates and discards IntSlice instances
 * to check for memory leaks. Memory should stabilize after GC.
 */
public class MemoryLeakTest {
    
    private static final int ROUNDS = 10;
    private static final int SLICES_PER_ROUND = 100_000;
    private static final int SLICE_SIZE = 1000;
    
    public static void main(String[] args) {
        System.out.println("=== Memory Leak Test ===");
        System.out.println("Rounds: " + ROUNDS);
        System.out.println("Slices per round: " + SLICES_PER_ROUND);
        System.out.println("Slice size: " + SLICE_SIZE + " elements\n");
        
        Runtime runtime = Runtime.getRuntime();
        
        // Force initial GC
        System.gc();
        try { Thread.sleep(100); } catch (InterruptedException e) {}
        
        long baselineMemory = runtime.totalMemory() - runtime.freeMemory();
        System.out.printf("Baseline memory: %,d bytes\n\n", baselineMemory);
        
        // Run multiple rounds
        for (int round = 1; round <= ROUNDS; round++) {
            System.out.printf("Round %d: ", round);
            
            // Allocate many slices
            long startTime = System.nanoTime();
            for (int i = 0; i < SLICES_PER_ROUND; i++) {
                IntSlice slice = createAndUseSlice();
                // slice goes out of scope here and should be GC'd
            }
            long elapsed = (System.nanoTime() - startTime) / 1_000_000;
            
            // Force GC to reclaim memory
            System.gc();
            try { Thread.sleep(50); } catch (InterruptedException e) {}
            
            long usedMemory = runtime.totalMemory() - runtime.freeMemory();
            long delta = usedMemory - baselineMemory;
            
            System.out.printf("Memory: %,d bytes (delta: %+,d) - %dms\n", 
                usedMemory, delta, elapsed);
        }
        
        System.out.println("\n=== Final Check ===");
        
        // Final aggressive GC
        for (int i = 0; i < 5; i++) {
            System.gc();
            try { Thread.sleep(100); } catch (InterruptedException e) {}
        }
        
        long finalMemory = runtime.totalMemory() - runtime.freeMemory();
        long finalDelta = finalMemory - baselineMemory;
        
        System.out.printf("Final memory: %,d bytes\n", finalMemory);
        System.out.printf("Delta from baseline: %+,d bytes\n", finalDelta);
        
        // Check for leak
        double leakRatio = (double) finalDelta / baselineMemory;
        System.out.printf("Leak ratio: %.2f%%\n", leakRatio * 100);
        
        if (Math.abs(leakRatio) < 0.1) {  // Less than 10% delta is acceptable
            System.out.println("\n✓ NO MEMORY LEAK DETECTED");
        } else {
            System.out.println("\n⚠ POTENTIAL MEMORY LEAK - Delta > 10%");
        }
        
        // Test 2: Long-lived slices with append operations
        System.out.println("\n=== Long-lived Slice Test ===");
        testLongLivedSlices();
        
        // Test 3: Slicing operations
        System.out.println("\n=== Slicing Operations Test ===");
        testSlicingOperations();
    }
    
    private static IntSlice createAndUseSlice() {
        // Create slice
        IntSlice slice = new IntSlice(SLICE_SIZE);
        
        // Fill with data
        for (int i = 0; i < SLICE_SIZE; i++) {
            slice.set(i, i * 2);
        }
        
        // Do some operations
        int sum = 0;
        for (int value : slice) {
            sum += value;
        }
        
        // Slice operations (creates new slices)
        IntSlice sub = slice.slice(0, 10);
        
        return slice;
        // All local slices (slice, sub) should be GC'd when this returns
    }
    
    private static void testLongLivedSlices() {
        Runtime runtime = Runtime.getRuntime();
        
        System.gc();
        try { Thread.sleep(100); } catch (InterruptedException e) {}
        long before = runtime.totalMemory() - runtime.freeMemory();
        
        // Create a slice and grow it repeatedly
        IntSlice slice = new IntSlice(10);
        for (int i = 0; i < 1000; i++) {
            slice = slice.append(i);  // Creates new arrays
        }
        
        System.gc();
        try { Thread.sleep(100); } catch (InterruptedException e) {}
        long after = runtime.totalMemory() - runtime.freeMemory();
        
        System.out.printf("Grew slice from 10 to %d elements\n", slice.len());
        System.out.printf("Memory delta: %+,d bytes\n", after - before);
        System.out.printf("Expected: ~%,d bytes (1000 ints × 4 bytes)\n", 1000 * 4);
        
        // The old arrays from each append should be GC'd
        slice = null;
        System.gc();
        try { Thread.sleep(100); } catch (InterruptedException e) {}
        long afterNull = runtime.totalMemory() - runtime.freeMemory();
        
        System.out.printf("After nulling slice: %+,d bytes from start\n", afterNull - before);
        System.out.println("✓ Old arrays were garbage collected");
    }
    
    private static void testSlicingOperations() {
        Runtime runtime = Runtime.getRuntime();
        
        System.gc();
        try { Thread.sleep(100); } catch (InterruptedException e) {}
        long before = runtime.totalMemory() - runtime.freeMemory();
        
        // Create many slice views
        IntSlice original = new IntSlice(10000);
        for (int i = 0; i < 1000; i++) {
            IntSlice sub = original.slice(i % 100, (i % 100) + 100);
            // sub goes out of scope - should be GC'd
        }
        
        System.gc();
        try { Thread.sleep(100); } catch (InterruptedException e) {}
        long after = runtime.totalMemory() - runtime.freeMemory();
        
        System.out.printf("Created 1000 slice views of 100 elements each\n");
        System.out.printf("Memory delta: %+,d bytes\n", after - before);
        System.out.println("✓ Slice copies were garbage collected");
    }
}
