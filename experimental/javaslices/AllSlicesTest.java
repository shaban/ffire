import java.nio.ByteBuffer;
import java.nio.ByteOrder;

/**
 * AllSlicesTest - Validates all slice types work correctly
 */
public class AllSlicesTest {
    
    public static void main(String[] args) {
        System.out.println("=== Testing All Slice Types ===\n");
        
        boolean allPassed = true;
        
        allPassed &= testByteSlice();
        allPassed &= testShortSlice();
        allPassed &= testIntSlice();
        allPassed &= testLongSlice();
        allPassed &= testFloatSlice();
        allPassed &= testDoubleSlice();
        
        System.out.println("\n=== Final Result ===");
        if (allPassed) {
            System.out.println("✓ ALL TESTS PASSED");
        } else {
            System.out.println("✗ SOME TESTS FAILED");
            System.exit(1);
        }
    }
    
    private static boolean testByteSlice() {
        System.out.println("Testing ByteSlice (int8/uint8)...");
        try {
            // Create and populate
            ByteSlice slice = new ByteSlice(new byte[]{1, 2, 3, 4, 5});
            assert slice.len() == 5 : "len() failed";
            assert slice.get(2) == 3 : "get() failed";
            
            // Append
            ByteSlice appended = slice.append((byte)6);
            assert appended.len() == 6 : "append() failed";
            
            // Slice
            ByteSlice sub = slice.slice(1, 4);
            assert sub.len() == 3 : "slice() failed";
            assert sub.get(0) == 2 : "slice() content failed";
            
            // Encode/Decode
            ByteBuffer buf = ByteBuffer.allocate(100);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            slice.encodeTo(buf);
            
            buf.flip();
            ByteSlice decoded = ByteSlice.decodeFrom(buf, 5);
            assert decoded.equals(slice) : "encode/decode failed";
            
            System.out.println("  ✓ ByteSlice passed\n");
            return true;
        } catch (AssertionError e) {
            System.out.println("  ✗ ByteSlice failed: " + e.getMessage() + "\n");
            return false;
        }
    }
    
    private static boolean testShortSlice() {
        System.out.println("Testing ShortSlice (int16/uint16)...");
        try {
            ShortSlice slice = new ShortSlice(new short[]{100, 200, 300, 400, 500});
            assert slice.len() == 5;
            assert slice.get(2) == 300;
            
            ShortSlice appended = slice.append((short)600);
            assert appended.len() == 6;
            
            ShortSlice sub = slice.slice(1, 4);
            assert sub.len() == 3;
            assert sub.get(0) == 200;
            
            ByteBuffer buf = ByteBuffer.allocate(100);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            slice.encodeTo(buf);
            
            buf.flip();
            ShortSlice decoded = ShortSlice.decodeFrom(buf, 5);
            assert decoded.equals(slice);
            
            System.out.println("  ✓ ShortSlice passed\n");
            return true;
        } catch (AssertionError e) {
            System.out.println("  ✗ ShortSlice failed: " + e.getMessage() + "\n");
            return false;
        }
    }
    
    private static boolean testIntSlice() {
        System.out.println("Testing IntSlice (int32/uint32)...");
        try {
            IntSlice slice = new IntSlice(new int[]{1000, 2000, 3000, 4000, 5000});
            assert slice.len() == 5;
            assert slice.get(2) == 3000;
            
            IntSlice appended = slice.append(6000);
            assert appended.len() == 6;
            
            IntSlice sub = slice.slice(1, 4);
            assert sub.len() == 3;
            assert sub.get(0) == 2000;
            
            ByteBuffer buf = ByteBuffer.allocate(100);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            slice.encodeTo(buf);
            
            buf.flip();
            IntSlice decoded = IntSlice.decodeFrom(buf, 5);
            assert decoded.equals(slice);
            
            System.out.println("  ✓ IntSlice passed\n");
            return true;
        } catch (AssertionError e) {
            System.out.println("  ✗ IntSlice failed: " + e.getMessage() + "\n");
            return false;
        }
    }
    
    private static boolean testLongSlice() {
        System.out.println("Testing LongSlice (int64/uint64)...");
        try {
            LongSlice slice = new LongSlice(new long[]{10000L, 20000L, 30000L, 40000L, 50000L});
            assert slice.len() == 5;
            assert slice.get(2) == 30000L;
            
            LongSlice appended = slice.append(60000L);
            assert appended.len() == 6;
            
            LongSlice sub = slice.slice(1, 4);
            assert sub.len() == 3;
            assert sub.get(0) == 20000L;
            
            ByteBuffer buf = ByteBuffer.allocate(100);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            slice.encodeTo(buf);
            
            buf.flip();
            LongSlice decoded = LongSlice.decodeFrom(buf, 5);
            assert decoded.equals(slice);
            
            System.out.println("  ✓ LongSlice passed\n");
            return true;
        } catch (AssertionError e) {
            System.out.println("  ✗ LongSlice failed: " + e.getMessage() + "\n");
            return false;
        }
    }
    
    private static boolean testFloatSlice() {
        System.out.println("Testing FloatSlice (float32)...");
        try {
            FloatSlice slice = new FloatSlice(new float[]{1.1f, 2.2f, 3.3f, 4.4f, 5.5f});
            assert slice.len() == 5;
            assert Math.abs(slice.get(2) - 3.3f) < 0.01f;
            
            FloatSlice appended = slice.append(6.6f);
            assert appended.len() == 6;
            
            FloatSlice sub = slice.slice(1, 4);
            assert sub.len() == 3;
            assert Math.abs(sub.get(0) - 2.2f) < 0.01f;
            
            ByteBuffer buf = ByteBuffer.allocate(100);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            slice.encodeTo(buf);
            
            buf.flip();
            FloatSlice decoded = FloatSlice.decodeFrom(buf, 5);
            assert decoded.equals(slice);
            
            System.out.println("  ✓ FloatSlice passed\n");
            return true;
        } catch (AssertionError e) {
            System.out.println("  ✗ FloatSlice failed: " + e.getMessage() + "\n");
            return false;
        }
    }
    
    private static boolean testDoubleSlice() {
        System.out.println("Testing DoubleSlice (float64)...");
        try {
            DoubleSlice slice = new DoubleSlice(new double[]{1.1, 2.2, 3.3, 4.4, 5.5});
            assert slice.len() == 5;
            assert Math.abs(slice.get(2) - 3.3) < 0.01;
            
            DoubleSlice appended = slice.append(6.6);
            assert appended.len() == 6;
            
            DoubleSlice sub = slice.slice(1, 4);
            assert sub.len() == 3;
            assert Math.abs(sub.get(0) - 2.2) < 0.01;
            
            ByteBuffer buf = ByteBuffer.allocate(100);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            slice.encodeTo(buf);
            
            buf.flip();
            DoubleSlice decoded = DoubleSlice.decodeFrom(buf, 5);
            assert decoded.equals(slice);
            
            System.out.println("  ✓ DoubleSlice passed\n");
            return true;
        } catch (AssertionError e) {
            System.out.println("  ✗ DoubleSlice failed: " + e.getMessage() + "\n");
            return false;
        }
    }
}
