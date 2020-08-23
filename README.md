# 背景
leveldb是一个google实现的非常高效的kv数据库，能够支持billion级别的数据量。
https://github.com/google/leveldb

## 准备工作
首先要了解leveldb的基本架构，下面我就列几个参考资料：
https://yuerblog.cc/wp-content/uploads/leveldb%E5%AE%9E%E7%8E%B0%E8%A7%A3%E6%9E%90.pdf

## 开始搬砖
1. skiplist