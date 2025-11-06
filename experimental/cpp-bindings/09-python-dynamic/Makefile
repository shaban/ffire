CXX = clang++
CXXFLAGS = -std=c++17 -O2 -Wall -Wextra
LIB = libffire.dylib

.PHONY: all clean

all: $(LIB)

# Build dynamic library with C ABI
$(LIB): ../common/generated_c.cpp ../common/generated_c.h ../common/generated.hpp
	$(CXX) $(CXXFLAGS) -shared -fPIC -o $(LIB) ../common/generated_c.cpp

clean:
	rm -f $(LIB) perf.json

.DEFAULT_GOAL := all
