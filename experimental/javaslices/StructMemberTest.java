import java.nio.ByteBuffer;
import java.nio.ByteOrder;

/**
 * Test if slices work as struct members (like in complex.ffi)
 * 
 * Schema equivalent:
 * type Parameter struct {
 *     DisplayName string
 *     Values      []int32     // IntSlice as member
 *     Flags       []bool      // Would need BoolSlice
 * }
 */
public class StructMemberTest {
    
    // Simulated struct with slice members
    static class Parameter {
        public String displayName;
        public IntSlice values;      // Slice as member
        public FloatSlice weights;   // Another slice
        
        public Parameter(String name, IntSlice values, FloatSlice weights) {
            this.displayName = name;
            this.values = values;
            this.weights = weights;
        }
        
        // Encode the struct with slice members
        public void encodeTo(ByteBuffer buf) {
            // Encode string
            byte[] nameBytes = displayName.getBytes();
            buf.putShort((short) nameBytes.length);
            buf.put(nameBytes);
            
            // Encode IntSlice
            buf.putShort((short) values.len());
            values.encodeTo(buf);
            
            // Encode FloatSlice
            buf.putShort((short) weights.len());
            weights.encodeTo(buf);
        }
        
        // Decode the struct with slice members
        public static Parameter decodeFrom(ByteBuffer buf) {
            // Decode string
            int nameLen = buf.getShort() & 0xFFFF;
            byte[] nameBytes = new byte[nameLen];
            buf.get(nameBytes);
            String name = new String(nameBytes);
            
            // Decode IntSlice
            int valuesLen = buf.getShort() & 0xFFFF;
            IntSlice values = IntSlice.decodeFrom(buf, valuesLen);
            
            // Decode FloatSlice
            int weightsLen = buf.getShort() & 0xFFFF;
            FloatSlice weights = FloatSlice.decodeFrom(buf, weightsLen);
            
            return new Parameter(name, values, weights);
        }
        
        @Override
        public String toString() {
            return String.format("Parameter{name='%s', values=%s, weights=%s}", 
                displayName, values, weights);
        }
    }
    
    public static void main(String[] args) {
        System.out.println("=== Slice as Struct Member Test ===\n");
        
        // Create struct with slice members
        IntSlice values = new IntSlice(new int[]{1, 2, 3, 4, 5});
        FloatSlice weights = new FloatSlice(new float[]{1.1f, 2.2f, 3.3f});
        Parameter param = new Parameter("TestParam", values, weights);
        
        System.out.println("1. Created struct with slice members:");
        System.out.println("   " + param);
        
        // Encode
        ByteBuffer buf = ByteBuffer.allocate(1000);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        param.encodeTo(buf);
        int encodedSize = buf.position();
        
        System.out.println("\n2. Encoded to " + encodedSize + " bytes");
        
        // Decode
        buf.flip();
        Parameter decoded = Parameter.decodeFrom(buf);
        
        System.out.println("\n3. Decoded struct:");
        System.out.println("   " + decoded);
        
        // Verify
        boolean nameMatch = param.displayName.equals(decoded.displayName);
        boolean valuesMatch = param.values.equals(decoded.values);
        boolean weightsMatch = param.weights.equals(decoded.weights);
        
        System.out.println("\n4. Verification:");
        System.out.println("   Name match: " + nameMatch);
        System.out.println("   Values match: " + valuesMatch);
        System.out.println("   Weights match: " + weightsMatch);
        
        // Test accessing slice members
        System.out.println("\n5. Accessing slice members:");
        System.out.println("   param.values.get(2) = " + decoded.values.get(2));
        System.out.println("   param.weights.len() = " + decoded.weights.len());
        
        // Test iteration
        System.out.print("   Iterating values: ");
        for (int v : decoded.values) {
            System.out.print(v + " ");
        }
        System.out.println();
        
        // Test mutation
        System.out.println("\n6. Mutation:");
        decoded.values.set(2, 999);
        System.out.println("   After set(2, 999): " + decoded.values);
        
        if (nameMatch && valuesMatch && weightsMatch) {
            System.out.println("\n✓ SLICES WORK AS STRUCT MEMBERS!");
        } else {
            System.out.println("\n✗ FAILED");
            System.exit(1);
        }
        
        // Test complex scenario: array of structs with slice members
        System.out.println("\n=== Array of Structs with Slice Members ===");
        testArrayOfStructs();
    }
    
    static void testArrayOfStructs() {
        // Create multiple Parameters
        Parameter[] params = new Parameter[3];
        params[0] = new Parameter("Param1", 
            new IntSlice(new int[]{1, 2, 3}),
            new FloatSlice(new float[]{1.1f}));
        params[1] = new Parameter("Param2",
            new IntSlice(new int[]{4, 5}),
            new FloatSlice(new float[]{2.2f, 3.3f}));
        params[2] = new Parameter("Param3",
            new IntSlice(new int[]{6, 7, 8, 9}),
            new FloatSlice(new float[]{4.4f}));
        
        System.out.println("Created array of 3 structs with slice members");
        
        // Encode all
        ByteBuffer buf = ByteBuffer.allocate(1000);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        buf.putShort((short) params.length);
        for (Parameter p : params) {
            p.encodeTo(buf);
        }
        
        System.out.println("Encoded to " + buf.position() + " bytes");
        
        // Decode all
        buf.flip();
        int count = buf.getShort() & 0xFFFF;
        Parameter[] decoded = new Parameter[count];
        for (int i = 0; i < count; i++) {
            decoded[i] = Parameter.decodeFrom(buf);
        }
        
        System.out.println("Decoded " + count + " structs:");
        for (int i = 0; i < decoded.length; i++) {
            System.out.println("  [" + i + "] " + decoded[i]);
        }
        
        System.out.println("\n✓ ARRAY OF STRUCTS WITH SLICE MEMBERS WORKS!");
    }
}
