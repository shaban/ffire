CXX = clang++
SWIFTC = swiftc
CXXFLAGS = -std=c++17 -O2 -Wall -Wextra
SWIFT_FLAGS = -O
LIB = libffire.dylib
TARGET = test

.PHONY: all clean

all: $(LIB) $(TARGET)

# Build dynamic library with C ABI (reuse from common or build here)
$(LIB): ../common/generated_c.cpp ../common/generated_c.h ../common/generated.hpp
	$(CXX) $(CXXFLAGS) -shared -fPIC -o $(LIB) ../common/generated_c.cpp

# Build Swift executable that links to dylib
$(TARGET): test.swift $(LIB)
	$(SWIFTC) $(SWIFT_FLAGS) -o $(TARGET) test.swift -L. -lffire -Xlinker -rpath -Xlinker @executable_path

clean:
	rm -f $(TARGET) $(LIB) perf.json

.DEFAULT_GOAL := all
