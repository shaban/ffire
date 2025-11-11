import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

/**
 * DoubleSlice - A Go-like slice wrapper around double[] primitive array.
 * For float64 type.
 */
public class DoubleSlice implements Iterable<Double> {
    private double[] data;
    
    public DoubleSlice(int capacity) {
        this.data = new double[capacity];
    }
    
    public DoubleSlice(double[] array) {
        this.data = array;
    }
    
    public int len() {
        return data.length;
    }
    
    public double get(int index) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        return data[index];
    }
    
    public void set(int index, double value) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        data[index] = value;
    }
    
    public double[] array() {
        return data;
    }
    
    public DoubleSlice append(double value) {
        double[] newData = Arrays.copyOf(data, data.length + 1);
        newData[data.length] = value;
        return new DoubleSlice(newData);
    }
    
    public DoubleSlice appendAll(double... values) {
        double[] newData = Arrays.copyOf(data, data.length + values.length);
        System.arraycopy(values, 0, newData, data.length, values.length);
        return new DoubleSlice(newData);
    }
    
    public DoubleSlice slice(int start, int end) {
        if (start < 0 || end > data.length || start > end) {
            throw new IndexOutOfBoundsException("slice bounds [" + start + ":" + end + "] out of range for length " + data.length);
        }
        return new DoubleSlice(Arrays.copyOfRange(data, start, end));
    }
    
    public DoubleSlice sliceFrom(int start) {
        return slice(start, data.length);
    }
    
    public DoubleSlice sliceTo(int end) {
        return slice(0, end);
    }
    
    @Override
    public Iterator<Double> iterator() {
        return new Iterator<Double>() {
            private int index = 0;
            
            @Override
            public boolean hasNext() {
                return index < data.length;
            }
            
            @Override
            public Double next() {
                if (!hasNext()) {
                    throw new NoSuchElementException();
                }
                return data[index++];
            }
        };
    }
    
    public List<Double> asList() {
        return Arrays.stream(data).boxed().collect(Collectors.toList());
    }
    
    public void encodeTo(ByteBuffer buf) {
        buf.asDoubleBuffer().put(data);
        buf.position(buf.position() + data.length * 8);
    }
    
    public static DoubleSlice decodeFrom(ByteBuffer buf, int length) {
        DoubleSlice slice = new DoubleSlice(length);
        buf.asDoubleBuffer().get(slice.data, 0, length);
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
        if (!(obj instanceof DoubleSlice)) return false;
        DoubleSlice other = (DoubleSlice) obj;
        return Arrays.equals(this.data, other.data);
    }
    
    @Override
    public int hashCode() {
        return Arrays.hashCode(data);
    }
}
