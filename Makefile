


db:
	docker stack deploy --compose-file tools/compose/db.docker-stack.yml db

rm:
	docker service rm $(shell docker service ls -q) || true
