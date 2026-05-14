# 配置
MOD_NAME  = github.com/ZOEKOFK/video_web_v3/api_gateway
IDL_FILE  = idl/common.proto idl/users.proto
IDL_DIR   = idl
OUT_DIR   = ./api_gateway
PROTO_PATH = idl

# 初始化项目
init:
	hz new \
    	--mod=$(MOD_NAME) \
    	--idl=$(IDL_FILE) \
    	--out_dir=$(OUT_DIR) \
    	--proto_path=$(PROTO_PATH)

# 更新 common.proto 生成的代码
update-common:
	hz update \
	--mod=$(MOD_NAME) \
	--idl=$(IDL_DIR)/common.proto \
	--out_dir=$(OUT_DIR) \
	--model_dir=biz/model/common \
	--proto_path=.

# 更新 users.proto 生成的代码
update-user:
	hz update \
	--mod=$(MOD_NAME)  \
	--idl=idl/users.proto \
	--out_dir=$(OUT_DIR) \
	--proto_path=.

gen-pb:
	protoc \
	--proto_path=. \
    --go_out=app/pb \
    --go-grpc_out=app/pb \
    idl/common.proto \
