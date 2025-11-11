import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

/**
 * ShortSlice - A Go-like slice wrapper around short[] primitive array.
 * For int16, uint16 types.
 */
public class ShortSlice implements Iterable<Short> {
    private short[] data;
    
    public ShortSlice(int capacity) {
        this.data = new short[capacity];
    }
    
    public ShortSlice(short[] array) {
        this.data = array;
    }
    
    public int len() {
        return data.length;
    }
    
    public short get(int index) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        return data[index];
    }
    
    public void set(int index, short value) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        data[index] = value;
    }
    
    public short[] array() {
        return data;
    }
    
    public ShortSlice append(short value) {
        short[] newData = Arrays.copyOf(data, data.length + 1);
        newData[data.length] = value;
        return new ShortSlice(newData);
    }
    
    public ShortSlice appendAll(short... values) {
        short[] newData = Arrays.copyOf(data, data.length + values.length);
        System.arraycopy(values, 0, newData, data.length, values.length);
        return new ShortSlice(newData);
    }
    
    public ShortSlice slice(int start, int end) {
        if (start < 0 || end > data.length || start > end) {
            throw new IndexOutOfBoundsException("slice bounds [" + start + ":" + end + "] out of range for length " + data.length);
        }
        return new ShortSlice(Arrays.copyOfRange(data, start, end));
    }
    
    public ShortSlice sliceFrom(int start) {
        return slice(start, data.length);
    }
    
    public ShortSlice sliceTo(int end) {
        return slice(0, end);
    }
    
    @Override
    public Iterator<Short> iterator() {
        return new Iterator<Short>() {
            private int index = 0;
            
            @Override
            public boolean hasNext() {
                return index < data.length;
            }
            
            @Override
            public Short next() {
                if (!hasNext()) {
                    throw new NoSuchElementException();
                }
                return data[index++];
            }
        };
    }
    
    public List<Short> asList() {
        Short[] boxed = new Short[data.length];
        for (int i = 0; i < data.length; i++) {
            boxed[i] = data[i];
        }
        return Arrays.asList(boxed);
    }
    
    public void encodeTo(ByteBuffer buf) {
        buf.asShortBuffer().put(data);
        buf.position(buf.position() + data.length * 2);
    }
    
    public static ShortSlice decodeFrom(ByteBuffer buf, int length) {
        ShortSlice slice = new ShortSlice(length);
        buf.asShortBuffer().get(slice.data, 0, length);
        buf.position(buf.position() + length * 2);
        return slice;
    }
    
    @Override
    public String toString() {
        return Arrays.toString(data);
    }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof ShortSlice)) return false;
        ShortSlice other = (ShortSlice) obj;
        return Arrays.equals(this.data, other.data);
    }
    
    @Override
    public int hashCode() {
        return Arrays.hashCode(data);
    }
}
