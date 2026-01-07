# 步骤 1：先获取并配置 GOPATH（临时生效）
# 获取 GOPATH 并赋值给变量（自动适配你的路径）
GOPATH=$(go env GOPATH)
# 将 GOPATH/bin 加入 PATH（解决插件找不到的问题）
export PATH=$PATH:$GOPATH/bin

# 步骤 2：验证插件是否存在（可选，但建议执行）
# 检查 protoc-gen-go 是否存在
if [ -f "$GOPATH/bin/protoc-gen-go" ]; then
  echo "protoc-gen-go 插件存在"
else
  echo "protoc-gen-go 插件不存在，正在安装..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# 检查 protoc-gen-go-grpc 是否存在
if [ -f "$GOPATH/bin/protoc-gen-go-grpc" ]; then
  echo "protoc-gen-go-grpc 插件存在"
else
  echo "protoc-gen-go-grpc 插件不存在，正在安装兼容版本..."
  # 使用与项目gRPC v1.56.3兼容的版本
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
fi


# 步骤 3：生成 Go 代码（核心命令）
# 生成基础 pb.go 文件
protoc --go_out=. --go_opt=paths=source_relative dtmgimp.proto
# 生成 grpc 相关 pb.go 文件
protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative dtmgimp.proto