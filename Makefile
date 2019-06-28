clean:
		rm -rf build
		mkdir build

build: clean
		go build -o build/fastgrace

zip: build
		rm -f fastgrace.zip
		cd build; zip fastgrace.zip fastgrace
