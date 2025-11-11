import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

/**
 * LongSlice - A Go-like slice wrapper around long[] primitive array.
 * For int64, uint64 types.
 */
public class LongSlice implements Iterable<Long> {
    private long[] data;
    
    public LongSlice(int capacity) {
        this.data = new long[capacity];
    }
    
    public LongSlice(long[] array) {
        this.data = array;
    }
    
    public int len() {
        return data.length;
    }
    
    public long get(int index) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        return data[index];
    }
    
    public void set(int index, long value) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        data[index] = value;
    }
    
    public long[] array() {
        return data;
    }
    
    public LongSlice append(long value) {
        long[] newData = Arrays.copyOf(data, data.length + 1);
        newData[data.length] = value;
        return new LongSlice(newData);
    }
    
    public LongSlice appendAll(long... values) {
        long[] newData = Arrays.copyOf(data, data.length + values.length);
        System.arraycopy(values, 0, newData, data.length, values.length);
        return new LongSlice(newData);
    }
    
    public LongSlice slice(int start, int end) {
        if (start < 0 || end > data.length || start > end) {
            throw new IndexOutOfBoundsException("slice bounds [" + start + ":" + end + "] out of range for length " + data.length);
        }
        return new LongSlice(Arrays.copyOfRange(data, start, end));
    }
    
    public LongSlice sliceFrom(int start) {
        return slice(start, data.length);
    }
    
    public LongSlice sliceTo(int end) {
        return slice(0, end);
    }
    
    @Override
    public Iterator<Long> iterator() {
        return new Iterator<Long>() {
            private int index = 0;
            
            @Override
            public boolean hasNext() {
                return index < data.length;
            }
            
            @Override
            public Long next() {
                if (!hasNext()) {
                    throw new NoSuchElementException();
                }
                return data[index++];
            }
        };
    }
    
    public List<Long> asList() {
        return Arrays.stream(data).boxed().collect(Collectors.toList());
    }
    
    public void encodeTo(ByteBuffer buf) {
        buf.asLongBuffer().put(data);
        buf.position(buf.position() + data.length * 8);
    }
    
    public static LongSlice decodeFrom(ByteBuffer buf, int length) {
        LongSlice slice = new LongSlice(length);
        buf.asLongBuffer().get(slice.data, 0, length);
        buf.position(buf.position() + length * 8);
        return slice;
    }
    
    @Override
    public String toString() {
        return Arrays.toString(data);
    }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof LongSlice)) return false;
        LongSlice other = (LongSlice) obj;
        return Arrays.equals(this.data, other.data);
    }
    
    @Override
    public int hashCode() {
        return Arrays.hashCode(data);
    }
}
