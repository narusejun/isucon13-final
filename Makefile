HOSTNAME:=$(shell hostname)
BRANCH:=master
PPROF_ID:=latest

.PHONY: download-pgo
download-pgo:
	wget https://192.168.1.100:9000/api/pprof/data/$(PPROF_ID)?label=$(HOSTNAME) -O app/webapp/go/pgo.pb.gz

.PHONY: deploy
deploy: checkout start

.PHONY: deploy-pgo
deploy-pgo: download-pgo deploy

.PHONY: checkout
checkout:
	git fetch && \
	git reset --hard origin/$(BRANCH)  && \
	git switch -C $(BRANCH) origin/$(BRANCH)

.PHONY: start
start:
	cd common && ./deploy.sh
