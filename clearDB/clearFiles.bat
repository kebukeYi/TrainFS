@echo off
setlocal enabledelayedexpansion

:: 指定要清理的目录列表
set directories=F:\TrainFS\DataNode\DataNode1 F:\TrainFS\DataNode\DataNode2 F:\TrainFS\DataNode\DataNode3 F:\TrainFS\NameNode1 F:\TrainFS\NameNode2

:: 遍历每个目录并删除其中的所有文件
for %%d in (%directories%) do (
    if exist "%%d\" (
        echo Removing all files from directory: %%d
        del /S /Q "%%d\*"
        if errorlevel 1 (
            echo Failed to remove files from directory: %%d
        ) else (
            echo Successfully removed files from directory: %%d
        )
    ) else (
        echo Directory not found: %%d
    )
)

echo Done.
pause
