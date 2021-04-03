

# ObjforceID 是一个面向超大型多租户系统的分布式id生成器

参考了 salesforce的id设计
以及 snowflake, sonyflake 等设计

ID 共18位, 统一用base62编码更紧凑, 且大小写不敏感;
salesforce的15位ID是大小写敏感的, 底层存储的是18位(大小写不敏感)
我们这里没有历史包袱，直接采用18位设计

ID中整体表征了如下信息:
0-3 3 bit, prefix: 业务填充的信息, 比如这条记录所代表的类型
3-5 2 bit, podIdentifier: pod实例，这个pod其实是集群的意思，不要和k8s的pod搞混
5-15 10 bit, numericIdentifier: 自增ID, 便于有序遍历, 这里默认基于 sonyflake(派生于snowflake) 进行改造
15-18 3 bit, subfix: 对前面15位id的3位校验位

### sonyflake的设计:
+-----------------------------------------------------------------------------+
| 1 Bit Unused | 39 Bit Timestamp |  8 Bit Sequence ID  |   16 Bit Machine ID |
+-----------------------------------------------------------------------------+
1 bit 保留
39 bit 为时间戳, 10ms的粒度
8 bit 为自增序列，也就是每10ms允许 2^8 = 256条新增, 也就是25600/秒的容量, 对于有一些业务, 可能容量不够
16 bit 为 machine id, 默认取当前私有ip后两字节, 相比snowflake的好处是默认不需要人工维护 machine id

### sonyflake 与 snowflake对比
sonyflake 相比snowflake, 兼容的有效时间从69年升至174年, 但是1秒最多生成的ID从409.6w降至2.56w条
sonyflake machineId = 16bit = 2^16 = 65535 理论上过大了, 但是基于ip自动提取, 小于16bit也无法唯一表征

### 设计权衡
- 针对 salesforce id的设计权衡
这里做了点小改动，去掉了salesforce中保留的 reserved 位, 把 数字自增列从9位 改成了10位. 主要是为了把 numericNumber的
容量进一步扩大，好兼容 objforce自增
62^10 = 839299365868340224

- 针对 sonyflake的改良
sonyflake 默认的表征范围
2^63 = 9223372036854775808
如果局限于 objforce id的表征范围 62^10 = 839299365868340224
那么
2^59 = 576460752303423488 < 62^10
也就是说还要额外多4bit的保留位 如果直接用sonyflake, 那么时间戳bit会减少4bit = 16, 意味着有效年限从174/16 = 11年, 这肯定不行.

所以我调整了 sonyflake的设计，把 machineID设置为12位, 把表征范围控制在 2^59 < 62^10 内
+-----------------------------------------------------------------------------+
| 5 Bit Unused | 39 Bit Timestamp |  8 Bit Sequence ID  |   12 Bit Machine ID |
+-----------------------------------------------------------------------------+

### 待改进
- 时钟回拨问题, 时钟回拨问题 sonyflake中直接sleep等待，有点粗暴, 这里直接拷贝过来的, 理论应该返回错误让业务处理
