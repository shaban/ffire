import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

/**
 * IntSlice - A Go-like slice wrapper around int[] primitive array.
 * 
 * Design goals:
 * - Zero boxing in codec paths (encode/decode)
 * - Iterable for Java for-each loops (accepts boxing here)
 * - Simple wrapper around primitive array
 * - Optional lazy conversion to List<Integer>
 */
public class IntSlice implements Iterable<Integer> {
    private int[] data;
    
    // Constructor with capacity
    public IntSlice(int capacity) {
        this.data = new int[capacity];
    }
    
    // Constructor from existing array
    public IntSlice(int[] array) {
        this.data = array;
    }
    
    // Go-like API
    public int len() {
        return data.length;
    }
    
    public int get(int index) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        return data[index];
    }
    
    public void set(int index, int value) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        data[index] = value;
    }
    
    // Direct array access (for performance-critical code)
    public int[] array() {
        return data;
    }
    
    // Append (creates new array - Go-like semantics)
    public IntSlice append(int value) {
        int[] newData = Arrays.copyOf(data, data.length + 1);
        newData[data.length] = value;
        return new IntSlice(newData);
    }
    
    // Append multiple values
    public IntSlice appendAll(int... values) {
        int[] newData = Arrays.copyOf(data, data.length + values.length);
        System.arraycopy(values, 0, newData, data.length, values.length);
        return new IntSlice(newData);
    }
    
    // Slice operation (Go-like)
    public IntSlice slice(int start, int end) {
        if (start < 0 || end > data.length || start > end) {
            throw new IndexOutOfBoundsException("slice bounds [" + start + ":" + end + "] out of range for length " + data.length);
        }
        return new IntSlice(Arrays.copyOfRange(data, start, end));
    }
    
    public IntSlice sliceFrom(int start) {
        return slice(start, data.length);
    }
    
    public IntSlice sliceTo(int end) {
        return slice(0, end);
    }
    
    // Iterable support (for Java for-each - accepts boxing)
    @Override
    public Iterator<Integer> iterator() {
        return new Iterator<Integer>() {
            private int index = 0;
            
            @Override
            public boolean hasNext() {
                return index < data.length;
            }
            
            @Override
            public Integer next() {
                if (!hasNext()) {
                    throw new NoSuchElementException();
                }
                return data[index++];  // Boxing happens here
            }
        };
    }
    
    // Lazy conversion to List (boxes all elements)
    public List<Integer> asList() {
        return Arrays.stream(data).boxed().collect(Collectors.toList());
    }
    
    // Codec operations (zero boxing)
    public void encodeTo(ByteBuffer buf) {
        buf.asIntBuffer().put(data);
        buf.position(buf.position() + data.length * 4);
    }
    
    public static IntSlice decodeFrom(ByteBuffer buf, int length) {
        IntSlice slice = new IntSlice(length);
        buf.asIntBuffer().get(slice.data, 0, length);
        buf.position(buf.position() + length * 4);
        return slice;
    }
    
    // Utility methods
    @Override
    public String toString() {
        return Arrays.toString(data);
    }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof IntSlice)) return false;
        IntSlice other = (IntSlice) obj;
        return Arrays.equals(this.data, other.data);
    }
    
    @Override
    public int hashCode() {
        return Arrays.hashCode(data);
    }
}
