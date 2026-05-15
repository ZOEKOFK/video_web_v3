# 配置
MOD_NAME  = github.com/ZOEKOFK/video_web_v3/api_gateway
IDL_FILE  = idl/common.proto idl/users.proto
IDL_DIR   = idl
OUT_DIR   = .
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
	--idl=$(IDL_DIR)/users.proto \
	--out_dir=$(OUT_DIR) \
	--proto_path=.

# 更新 users.proto 生成的代码
update:
	hz update \
	--mod=$(MOD_NAME)  \
	--idl=idl/interaction.proto \
	--out_dir=$(OUT_DIR) \
	--proto_path=.

