

# ObjforceID 是一个面向超大型多租户系统的分布式id生成器

参考了 salesforce的id设计
以及 snowflake, sonyflake 等设计

这里做了点小改动，去掉了salesforce中保留的 reserved位, 把 数字自增列从9位 改成了10位. 主要这样可以无缝兼容 64bit 的设计

约束每个 pod 是局域网，那么 machineID 就不需要人工配置, sonyflake使用了16bit作为 machineID, 默认获取ip后两字节

2^64 = 18446744073709551616
62^10 = 839299365868340224