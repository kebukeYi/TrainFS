#!/bin/bash

# 定义要删除文件的目录
TARGET_DIR="/path/to/your/directory"

# 检查目录是否存在
if [ ! -d "$TARGET_DIR" ]; then
    echo "目录 $TARGET_DIR 不存在，请检查路径是否正确。"
    exit 1
fi

# 删除指定目录下的所有文件
echo "正在删除 $TARGET_DIR 下的所有文件..."
find "$TARGET_DIR" -maxdepth 1 -type f -delete

# 检查删除操作是否成功
if [ $? -eq 0 ]; then
    echo "文件删除成功。"
else
    echo "文件删除过程中出现错误。"
fi
