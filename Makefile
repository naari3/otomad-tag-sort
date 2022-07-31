.PHONY: create-index
bin:
	mkdir -p bin

clean:
	rm -rf bin/*

run: otomad-tag-sort
	./bin/otomad-tag-sort

run-collect: otomad-tag-sort-collect
	./bin/otomad-tag-sort-collect

run-tag-cache: otomad-tag-sort-tag-cache
	./bin/otomad-tag-sort-tag-cache

otomad-tag-sort: bin
	go build -o bin/otomad-tag-sort ./cmd/otomad-tag-sort

otomad-tag-sort-collect: bin
	go build -o bin/otomad-tag-sort-collect ./cmd/otomad-tag-sort-collect

otomad-tag-sort-tag-cache: bin
	go build -o bin/otomad-tag-sort-tag-cache ./cmd/otomad-tag-sort-tag-cache

create-index:
	cd docs/ && python ../generate_directory_index_caddystyle.py -r
