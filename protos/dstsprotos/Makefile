DSTS_PROTO=ghcr.io/hpinc/krypton/krypton-dstsprotos:latest
DSTS_PROTO_LOCAL=dsts_protos_container
DSTS_PROTOS_DIR=.

fetch-protos:
	docker pull $(DSTS_PROTO)
	docker create --name $(DSTS_PROTO_LOCAL) $(DSTS_PROTO) "" && \
	docker cp $(DSTS_PROTO_LOCAL):/protos $(DSTS_PROTOS_DIR) && \
	docker rm $(DSTS_PROTO_LOCAL) && \
	mv $(DSTS_PROTOS_DIR)/protos/* $(DSTS_PROTOS_DIR) && rm -rf $(DSTS_PROTOS_DIR)/protos
