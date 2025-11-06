import 'dart:ffi';
import 'dart:io';
import 'dart:typed_data';

class TestException implements Exception {
  final String message;
  TestException(this.message);
  @override
  String toString() => message;
}

class _NativeLibrary {
  static final DynamicLibrary _lib = _loadLibrary();

  static DynamicLibrary _loadLibrary() {
    if (Platform.isMacOS) {
      return DynamicLibrary.open('libffire.dylib');
    } else if (Platform.isLinux) {
      return DynamicLibrary.open('libffire.so');
    } else if (Platform.isWindows) {
      return DynamicLibrary.open('ffire.dll');
    }
    throw UnsupportedError('Platform not supported');
  }

  static final ffire_decode = _lib
      .lookup<NativeFunction<Pointer<Void> Function(Pointer<Uint8>, Int32)>>('ffire_decode')
      .asFunction<Pointer<Void> Function(Pointer<Uint8>, int)>();

  static final ffire_encode = _lib
      .lookup<NativeFunction<Pointer<Void> Function(Pointer<Void>)>>('ffire_encode')
      .asFunction<Pointer<Void> Function(Pointer<Void>)>();

  static final ffire_free = _lib
      .lookup<NativeFunction<Void Function(Pointer<Void>)>>('ffire_free')
      .asFunction<void Function(Pointer<Void>)>();

  static final ffire_get_error = _lib
      .lookup<NativeFunction<Pointer<Utf8> Function()>>('ffire_get_error')
      .asFunction<Pointer<Utf8> Function()>();

  static String getLastError() {
    final errPtr = ffire_get_error();
    if (errPtr.address == 0) return 'Unknown error';
    return errPtr.toDartString();
  }
}

class Message {
  Pointer<Void>? _handle;
  bool _disposed = false;

  Message._(this._handle);

  static Message decode(Uint8List data) {
    final dataPtr = malloc.allocate<Uint8>(data.length);
    final dataList = dataPtr.asTypedList(data.length);
    dataList.setAll(0, data);

    final handle = _NativeLibrary.ffire_decode(dataPtr, data.length);
    malloc.free(dataPtr);

    if (handle.address == 0) {
      final error = _NativeLibrary.getLastError();
      throw TestException('Decode failed: $error');
    }

    return Message._(handle);
  }

  Uint8List encode() {
    if (_disposed) {
      throw StateError('Message has been disposed');
    }

    final resultPtr = _NativeLibrary.ffire_encode(_handle!);
    if (resultPtr.address == 0) {
      final error = _NativeLibrary.getLastError();
      throw TestException('Encode failed: $error');
    }

    // Read length (first 4 bytes) and data
    final lengthPtr = resultPtr.cast<Uint32>();
    final length = lengthPtr.value;
    final dataPtr = Pointer<Uint8>.fromAddress(resultPtr.address + 4);
    final data = Uint8List.fromList(dataPtr.asTypedList(length));

    return data;
  }

  void dispose() {
    if (!_disposed && _handle != null) {
      _NativeLibrary.ffire_free(_handle!);
      _handle = null;
      _disposed = true;
    }
  }
}
