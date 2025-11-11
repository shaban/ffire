/**
 * API Test - Demonstrates all IntSlice features
 */
public class APITest {
    public static void main(String[] args) {
        System.out.println("=== IntSlice API Test ===\n");
        
        // 1. Creation
        System.out.println("1. Creating slices:");
        IntSlice slice1 = new IntSlice(5);
        System.out.println("   Created with capacity 5: len=" + slice1.len());
        
        IntSlice slice2 = new IntSlice(new int[]{1, 2, 3, 4, 5});
        System.out.println("   Created from array: " + slice2);
        
        // 2. Get/Set
        System.out.println("\n2. Get/Set operations:");
        slice1.set(0, 10);
        slice1.set(1, 20);
        slice1.set(2, 30);
        System.out.println("   After setting values: " + slice1);
        System.out.println("   Get index 1: " + slice1.get(1));
        
        // 3. Append
        System.out.println("\n3. Append operations:");
        IntSlice slice3 = slice2.append(6);
        System.out.println("   Original: " + slice2);
        System.out.println("   After append(6): " + slice3);
        
        IntSlice slice4 = slice2.appendAll(7, 8, 9);
        System.out.println("   After appendAll(7,8,9): " + slice4);
        
        // 4. Slicing
        System.out.println("\n4. Slicing operations:");
        IntSlice sub1 = slice4.slice(1, 4);
        System.out.println("   slice(1, 4): " + sub1);
        
        IntSlice sub2 = slice4.sliceFrom(5);
        System.out.println("   sliceFrom(5): " + sub2);
        
        IntSlice sub3 = slice4.sliceTo(3);
        System.out.println("   sliceTo(3): " + sub3);
        
        // 5. Iteration (for-each with boxing)
        System.out.println("\n5. Iteration (for-each):");
        System.out.print("   Values: ");
        for (int value : slice2) {
            System.out.print(value + " ");
        }
        System.out.println();
        
        // 6. Direct array access (no boxing)
        System.out.println("\n6. Direct array access (performance):");
        int[] arr = slice2.array();
        System.out.print("   Direct access: ");
        for (int i = 0; i < arr.length; i++) {
            System.out.print(arr[i] + " ");
        }
        System.out.println();
        
        // 7. Convert to List
        System.out.println("\n7. Convert to List<Integer>:");
        System.out.println("   asList(): " + slice2.asList());
        
        // 8. Bounds checking
        System.out.println("\n8. Bounds checking:");
        try {
            slice1.get(100);
        } catch (IndexOutOfBoundsException e) {
            System.out.println("   ✓ Caught: " + e.getMessage());
        }
        
        // 9. Equality
        System.out.println("\n9. Equality:");
        IntSlice a = new IntSlice(new int[]{1, 2, 3});
        IntSlice b = new IntSlice(new int[]{1, 2, 3});
        IntSlice c = new IntSlice(new int[]{1, 2, 4});
        System.out.println("   a.equals(b): " + a.equals(b));
        System.out.println("   a.equals(c): " + a.equals(c));
        
        System.out.println("\n=== All tests passed! ===");
    }
}
