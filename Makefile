BINARY=simpmulator
ARGS?=testdata/musl/hello/hello

run: $(BINARY)
	./$< $(ARGS)

build: $(BINARY) gen
	go build -o $<

gen:
	go get .
	go generate .

clean:
	rm $(BINARY)
