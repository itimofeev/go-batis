

run-env: stop-env
	docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password --name=db postgres:13

stop-env:
	docker rm -f db || true

lint:
	#go get golang.org/x/tools/cmd/goimports
	goimports -w -local github.com/itimofeev/simple-billing internal cmd

	#curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.39.0
	GO111MODULE=on GL_DEBUG=debug L_DEBUG=linters_output GOPACKAGESPRINTGOLISTERRORS=1 golangci-lint -v run
