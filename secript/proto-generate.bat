@echo off
chcp 65001 >nul
cd /d "%~dp0.."
echo 开始生成 Protobuf Go 代码...
echo 当前目录: %CD%
echo.

echo 生成 common proto...
protoc --proto_path=api^
       --go_out=pb^
       --go_opt=paths=source_relative api/common/common.proto
if errorlevel 1 (
    echo 错误: common proto 生成失败
    pause
    exit /b 1
)

echo 生成 gateway proto...
protoc --proto_path=api^
        --go_out=pb^
        --go_opt=paths=source_relative^
        --go-grpc_out=pb ^
        --go-grpc_opt=paths=source_relative api/gateway/gateway.proto
if errorlevel 1 (
    echo 错误: gateway proto 生成失败
    pause
    exit /b 1
)


echo.
echo 所有 Proto 文件生成成功！
pause
