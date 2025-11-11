import java.util.List;

/**
 * Ergonomics Test - How does IntSlice feel to use in real code?
 */
public class ErgonomicsTest {
    
    public static void main(String[] args) {
        System.out.println("=== IntSlice Ergonomics Test ===\n");
        
        // Scenario 1: Building data from scratch
        System.out.println("1. Building data incrementally:");
        IntSlice numbers = buildSequence(1, 10);
        System.out.println("   Built sequence: " + numbers);
        
        // Scenario 2: Processing/transforming data
        System.out.println("\n2. Processing data:");
        IntSlice doubled = doubleValues(numbers);
        System.out.println("   Doubled: " + doubled);
        
        IntSlice filtered = filterEven(numbers);
        System.out.println("   Even numbers: " + filtered);
        
        // Scenario 3: Aggregation
        System.out.println("\n3. Aggregation:");
        int sum = sumValues(numbers);
        System.out.println("   Sum: " + sum);
        
        int max = maxValue(numbers);
        System.out.println("   Max: " + max);
        
        // Scenario 4: Java standard library integration
        System.out.println("\n4. Integration with Java stdlib:");
        List<Integer> list = numbers.asList();
        System.out.println("   As List: " + list);
        System.out.println("   List.contains(5): " + list.contains(5));
        
        // Scenario 5: For-each iteration (most common use case)
        System.out.println("\n5. For-each iteration (ergonomics):");
        System.out.print("   ");
        for (int value : numbers) {
            System.out.print(value + " ");
        }
        System.out.println();
        
        // Scenario 6: Index-based access (when you need it)
        System.out.println("\n6. Index-based access:");
        System.out.println("   First: " + numbers.get(0));
        System.out.println("   Last: " + numbers.get(numbers.len() - 1));
        System.out.println("   Middle: " + numbers.get(numbers.len() / 2));
        
        // Scenario 7: Slicing (Go-like operations)
        System.out.println("\n7. Slicing operations:");
        IntSlice first5 = numbers.sliceTo(5);
        System.out.println("   First 5: " + first5);
        
        IntSlice last5 = numbers.sliceFrom(numbers.len() - 5);
        System.out.println("   Last 5: " + last5);
        
        IntSlice middle = numbers.slice(3, 7);
        System.out.println("   Middle [3:7]: " + middle);
        
        // Scenario 8: Mutation
        System.out.println("\n8. Mutation:");
        IntSlice mutable = new IntSlice(new int[]{1, 2, 3, 4, 5});
        System.out.println("   Before: " + mutable);
        mutable.set(2, 999);
        System.out.println("   After set(2, 999): " + mutable);
        
        // Scenario 9: Chaining operations (immutable style)
        System.out.println("\n9. Chaining operations:");
        IntSlice chain = new IntSlice(new int[]{1, 2, 3})
            .append(4)
            .append(5)
            .appendAll(6, 7, 8);
        System.out.println("   Chained: " + chain);
        
        // Scenario 10: Direct array access for algorithms
        System.out.println("\n10. Direct array access (performance):");
        int[] arr = numbers.array();
        System.out.print("   Manual loop: ");
        for (int i = 0; i < arr.length; i++) {
            System.out.print(arr[i] * 2 + " ");
        }
        System.out.println();
        
        System.out.println("\n=== Ergonomics Assessment ===");
        System.out.println("✓ For-each syntax works naturally");
        System.out.println("✓ Index access is straightforward");
        System.out.println("✓ Slicing feels Go-like");
        System.out.println("✓ Converts to List when needed");
        System.out.println("✓ Direct array access available for performance");
        System.out.println("✓ Immutable operations (append) return new slice");
    }
    
    // Helper functions demonstrating common patterns
    
    static IntSlice buildSequence(int start, int end) {
        IntSlice result = new IntSlice(0);
        for (int i = start; i <= end; i++) {
            result = result.append(i);
        }
        return result;
    }
    
    static IntSlice doubleValues(IntSlice input) {
        int[] result = new int[input.len()];
        for (int i = 0; i < input.len(); i++) {
            result[i] = input.get(i) * 2;
        }
        return new IntSlice(result);
    }
    
    static IntSlice filterEven(IntSlice input) {
        IntSlice result = new IntSlice(0);
        for (int value : input) {
            if (value % 2 == 0) {
                result = result.append(value);
            }
        }
        return result;
    }
    
    static int sumValues(IntSlice input) {
        int sum = 0;
        for (int value : input) {
            sum += value;
        }
        return sum;
    }
    
    static int maxValue(IntSlice input) {
        if (input.len() == 0) {
            throw new IllegalArgumentException("empty slice");
        }
        int max = input.get(0);
        for (int value : input) {
            if (value > max) {
                max = value;
            }
        }
        return max;
    }
}
