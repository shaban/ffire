CXX = clang++
CXXFLAGS = -std=c++17 -O2 -Wall -Wextra -fobjc-arc
FRAMEWORKS = -framework Foundation
TARGET = test
LIB = libffire.dylib

.PHONY: all clean

all: $(TARGET)

# Build dynamic library with C ABI
$(LIB): ../common/generated_c.cpp ../common/generated_c.h ../common/generated.hpp
	$(CXX) $(CXXFLAGS) -shared -fPIC -o $(LIB) ../common/generated_c.cpp

$(TARGET): test.mm $(LIB)
	$(CXX) $(CXXFLAGS) $(FRAMEWORKS) -o $(TARGET) test.mm -L. -lffire -Wl,-rpath,@executable_path

clean:
	rm -f $(TARGET) $(LIB) *.o perf.json

.DEFAULT_GOAL := all
