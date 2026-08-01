# 第一轮

ai-gateway-api, 对 design-docs\api-define\OpenAPI接口定义\clusters.md 做修改

- 1. 数据模型

  - **Instance 结构**
    - 修改前：hostname和ip都是必填
    - 修改后：hostname和ip，两者至少要填一个

- 对 2.1 创建集群 和 2.4 更新集群基本配置 时

  - 遵从以上修改



# 第二轮

ai-gateway-api, 对 design-docs\api-define\OpenAPI接口定义\clusters.md 做修改

- 1. 数据模型

  - **Instance 结构**
    - 修改前：
      - hostname和ip，两者至少要填一个
      - ports：map 至少1个元素，且必须包含 `Default` 端口
    - 修改后：
      - hostname：改名为name，选填，长度为1-128字符。如果不填，会设置为和Addr相同
      - ip：改名为addr，必填，类型为Hostname
      - ports：改名为port, 类型为 [Port](./00-common.md#3-网络端口port)

- 对 2.1 创建集群 和 2.4 更新集群基本配置 时

  - 遵从以上修改

# 第三轮

ai-gateway-api, 对 design-docs\api-define\OpenAPI接口定义\clusters.md 做修改

- 1. 数据模型

  - 字段说明
    -  `instance_pool` 
      - 合法性条件，改为：必填；至少1个元素；同一集群内 ，对于name不为空的instance，`name` 不能重复; 同一集群内，`(name, addr)` 组合不能重复；至少有一个实例 `weight > 0`

- **表：被动健康检查**
  - host: 
    - 补充描述，改为：为空时使用 `instance_pool` 中第一个实例的 addr
    - 合法性条件，改为：非必填；为空时使用 `instance_pool` 首个实例的 `addr`

# 第四轮

ai-gateway-api, 对 design-docs\api-define\OpenAPI接口定义\clusters.md

- 2.1 创建集群
  - 对于 basic，protocol的默认值为https



# 第五轮

