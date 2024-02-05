%.sql: .FORCE
	cat $(*).sql | docker compose exec -T db bash -c 'psql -U postgres -d postgres'
drop-all:
	docker compose exec -T db bash -c 'psql -U postgres -d postgres -c "drop schema public cascade; create schema public;"'

.FORCE: