bin:
	mkdir -p bin

clean:
	rm -rf bin/*

run: otagged
	./bin/otagged

run-collect: otagged-collect
	./bin/otagged-collect

otagged: bin
	go build -o bin/otagged ./cmd/otagged

otagged-collect: bin
	go build -o bin/otagged-collect ./cmd/otagged-collect
