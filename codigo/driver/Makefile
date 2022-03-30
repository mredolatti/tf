.PHONY: cmake cmake-build

cmake:
	cmake -DCMAKE_BUILD_TYPE=Debug -DCMAKE_EXPORT_COMPILE_COMMANDS=on -B ./build .
	cp build/compile_commands.json .

cmake-build:
	cmake --build ./build --verbose

clean:
	rm -Rf build/*
