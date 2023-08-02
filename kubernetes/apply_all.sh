#!/bin/bash

# 定义kubernetes目录的路径
KUBERNETES_DIR="./"

# 遍历kubernetes目录下的每一个子目录
for COMPONENT_DIR in $(find $KUBERNETES_DIR -type d -mindepth 1); do
    # 对子目录中的每一个.yaml文件执行kubectl apply操作
    for YAML_FILE in $(find $COMPONENT_DIR -name "*.yaml"); do
        kubectl apply -f $YAML_FILE
    done
done