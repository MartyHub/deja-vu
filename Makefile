.PHONY: clean db_start db_stop lint mysql_start mysql_stop pg_start pg_stop test

default: lint test

clean: db_stop
	rm -f coverage.out

db_start: mysql_start pg_start

db_stop: mysql_stop pg_stop

mysql_start:
	./scripts/mysql_start.sh

mysql_stop:
	./scripts/mysql_stop.sh

pg_start:
	./scripts/pg_start.sh

pg_stop:
	./scripts/pg_stop.sh

lint:
	./scripts/lint.sh

test: db_start
	./scripts/test.sh
