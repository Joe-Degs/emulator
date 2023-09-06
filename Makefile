BINARY=simpmulator
ARGS?=testdata/musl/hello/hello

run: $(BINARY)
	./$< $(ARGS)

$(BINARY): gen
	go build -o $@

debug: $(BINARY)
	./$< -v -elf-info -dump-state -verbose-pc $(ARGS)

debug-inst: $(BINARY)
	./$< -v -elf-info -dump-state -verbose-inst -verbose-pc $(ARGS)

gen:
	#go get .
	go generate .

clean:
	rm $(BINARY)
