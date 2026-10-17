BUILD_DIR ?= build

.PHONY: all clean

all:
	cmake -S . -B $(BUILD_DIR) -DCMAKE_BUILD_TYPE=Release
	cmake --build $(BUILD_DIR) --parallel

clean:
	cmake -E rm -rf $(BUILD_DIR)
