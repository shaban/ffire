import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;

/**
 * FloatSlice - A Go-like slice wrapper around float[] primitive array.
 * For float32 type.
 */
public class FloatSlice implements Iterable<Float> {
    private float[] data;
    
    public FloatSlice(int capacity) {
        this.data = new float[capacity];
    }
    
    public FloatSlice(float[] array) {
        this.data = array;
    }
    
    public int len() {
        return data.length;
    }
    
    public float get(int index) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        return data[index];
    }
    
    public void set(int index, float value) {
        if (index < 0 || index >= data.length) {
            throw new IndexOutOfBoundsException("index " + index + " out of bounds for length " + data.length);
        }
        data[index] = value;
    }
    
    public float[] array() {
        return data;
    }
    
    public FloatSlice append(float value) {
        float[] newData = Arrays.copyOf(data, data.length + 1);
        newData[data.length] = value;
        return new FloatSlice(newData);
    }
    
    public FloatSlice appendAll(float... values) {
        float[] newData = Arrays.copyOf(data, data.length + values.length);
        System.arraycopy(values, 0, newData, data.length, values.length);
        return new FloatSlice(newData);
    }
    
    public FloatSlice slice(int start, int end) {
        if (start < 0 || end > data.length || start > end) {
            throw new IndexOutOfBoundsException("slice bounds [" + start + ":" + end + "] out of range for length " + data.length);
        }
        return new FloatSlice(Arrays.copyOfRange(data, start, end));
    }
    
    public FloatSlice sliceFrom(int start) {
        return slice(start, data.length);
    }
    
    public FloatSlice sliceTo(int end) {
        return slice(0, end);
    }
    
    @Override
    public Iterator<Float> iterator() {
        return new Iterator<Float>() {
            private int index = 0;
            
            @Override
            public boolean hasNext() {
                return index < data.length;
            }
            
            @Override
            public Float next() {
                if (!hasNext()) {
                    throw new NoSuchElementException();
                }
                return data[index++];
            }
        };
    }
    
    public List<Float> asList() {
        Float[] boxed = new Float[data.length];
        for (int i = 0; i < data.length; i++) {
            boxed[i] = data[i];
        }
        return Arrays.asList(boxed);
    }
    
    public void encodeTo(ByteBuffer buf) {
        buf.asFloatBuffer().put(data);
        buf.position(buf.position() + data.length * 4);
    }
    
    public static FloatSlice decodeFrom(ByteBuffer buf, int length) {
        FloatSlice slice = new FloatSlice(length);
        buf.asFloatBuffer().get(slice.data, 0, length);
        buf.position(buf.position() + length * 4);
        return slice;
    }
    
    @Override
    public String toString() {
        return Arrays.toString(data);
    }
    
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof FloatSlice)) return false;
        FloatSlice other = (FloatSlice) obj;
        return Arrays.equals(this.data, other.data);
    }
    
    @Override
    public int hashCode() {
        return Arrays.hashCode(data);
    }
}
