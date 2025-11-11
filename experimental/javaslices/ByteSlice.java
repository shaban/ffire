import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

/**
 * ByteSlice - A Go-like slice wrapper around byte[] primitive array.
 * For int8, uint8, byte types.
 */
public class ByteSlice implements Iterable<Byte> {
    private byte[] data;
    
    public ByteSlice(int capacity) {
        this.data = new byte[capacity];
    }
    
    public ByteSlice(byte[] array) {
        this.data = array;
    }
    
    public int len() {
        return data.length;
    }
    
    public byte get(int index) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        return data[index];
    }
    
    public void set(int index, byte value) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        data[index] = value;
    }
    
    public byte[] array() {
        return data;
    }
    
    public ByteSlice append(byte value) {
        byte[] newData = Arrays.copyOf(data, data.length + 1);
        newData[data.length] = value;
        return new ByteSlice(newData);
    }
    
    public ByteSlice appendAll(byte... values) {
        byte[] newData = Arrays.copyOf(data, data.length + values.length);
        System.arraycopy(values, 0, newData, data.length, values.length);
        return new ByteSlice(newData);
    }
    
    public ByteSlice slice(int start, int end) {
        if (start < 0 || end > data.length || start > end) {
            throw new IndexOutOfBoundsException("slice bounds [" + start + ":" + end + "] out of range for length " + data.length);
        }
        return new ByteSlice(Arrays.copyOfRange(data, start, end));
    }
    
    public ByteSlice sliceFrom(int start) {
        return slice(start, data.length);
    }
    
    public ByteSlice sliceTo(int end) {
        return slice(0, end);
    }
    
    @Override
    public Iterator<Byte> iterator() {
        return new Iterator<Byte>() {
            private int index = 0;
            
            @Override
            public boolean hasNext() {
                return index < data.length;
            }
            
            @Override
            public Byte next() {
                if (!hasNext()) {
                    throw new NoSuchElementException();
                }
                return data[index++];
            }
        };
    }
    
    public List<Byte> asList() {
        Byte[] boxed = new Byte[data.length];
        for (int i = 0; i < data.length; i++) {
            boxed[i] = data[i];
        }
        return Arrays.asList(boxed);
    }
    
    public void encodeTo(ByteBuffer buf) {
        buf.put(data);
    }
    
    public static ByteSlice decodeFrom(ByteBuffer buf, int length) {
        ByteSlice slice = new ByteSlice(length);
        buf.get(slice.data, 0, length);
        return slice;
    }
    
    @Override
    public String toString() {
        return Arrays.toString(data);
    }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof ByteSlice)) return false;
        ByteSlice other = (ByteSlice) obj;
        return Arrays.equals(this.data, other.data);
    }
    
    @Override
    public int hashCode() {
        return Arrays.hashCode(data);
    }
}
