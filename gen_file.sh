#!/bin/bash

num_files=11  # 指定要创建的文件数量
file_size=1073741824  # 每个文件的大小，单位是字节
filename_base="/var/tmp/file_"

for ((i=0; i<$num_files; i++))
do
    filename="${filename_base}${i}"
    dd if=/dev/zero of="$filename" bs=$file_size count=1
done
dd if=/dev/zero of=disk.img bs=1G count=11