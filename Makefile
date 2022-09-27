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

run-db-create: otomad-tag-sort-db-create
	./bin/otomad-tag-sort-db-create full

# run-mysql-create: otomad-tag-sort-mysql-create
# 	./bin/otomad-tag-sort-mysql-create full

run-db-update: otomad-tag-sort-db-create
	./bin/otomad-tag-sort-db-create update

otomad-tag-sort: bin
	go build -o bin/otomad-tag-sort ./cmd/otomad-tag-sort

otomad-tag-sort-collect: bin
	go build -o bin/otomad-tag-sort-collect ./cmd/otomad-tag-sort-collect

otomad-tag-sort-tag-cache: bin
	go build -o bin/otomad-tag-sort-tag-cache ./cmd/otomad-tag-sort-tag-cache

otomad-tag-sort-db-create: bin
	go build -o bin/otomad-tag-sort-db-create ./cmd/otomad-tag-sort-db-create

# otomad-tag-sort-mysql-create: bin
# 	go build -o bin/otomad-tag-sort-mysql-create ./cmd/otomad-tag-sort-mysql-create

create-index:
	cd docs/ && python ../generate_directory_index_caddystyle.py -r
