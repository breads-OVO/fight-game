@echo off
chcp 65001 >nul
cd /d "%~dp0.."
echo 开始生成 Protobuf Go 代码...
echo 当前目录: %CD%
echo.


echo 生成所有 proto...
protoc --proto_path=. ^
       --go_out=. ^
       --go_opt=module=fight-game ^
       --go-grpc_out=. ^
       --go-grpc_opt=module=fight-game ^
       api/common/common.proto ^
       api/auth/auth_login.proto ^
       api/auth/auth_register.proto ^
       api/auth/auth_token.proto ^
       api/auth/auth_service.proto ^
       api/match/match_queue.proto ^
       api/match/match_service.proto ^
       api/game/game.proto ^
       api/game/game_service.proto
if errorlevel 1 (
    echo 错误: proto 生成失败
    pause
    exit /b 1
)

echo.
echo 所有 Proto 文件生成成功！
pause
