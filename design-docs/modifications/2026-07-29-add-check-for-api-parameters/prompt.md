# 第1轮

- 目前 design-docs\api-define 的文档中，对于各接口的输入参数的合法性条件，缺乏严格的说明
- 你是否可以尝试对于 design-docs\api-define 下各文档的各API的各参数增加合法性条件的说明
- 这个说明，后续将用于增强api代码实现中的合法性检查

# 第2轮

- 对于alb-pool.md
  - 2.2中
    - instances[].hostname：rfc对于hostname的规定？
    - instances[].ip：rfc对于ip地址（ipv4, ipv6）的要求？
    - instances[].ports：对于网络端口的格式要求？
- 对于alb-pool.md，2.2中
  - 对instances[].hostname，已经规定 类型为 [Hostname](./00-common.md#公共参数类型)。长度 ≥2 字符 这个要求还需要吗？
- 对于alb-pool.md，1中
  - 下面这句话好像不需要出现？
    - AI 网关实例池角色固定为 `COMMON`，无需 EPP Server 配置。

# 第3轮

对api-keys.md

- 2.1，输入参数，合法性条件
  - key: 增加，长度 1-128 字符
  - models：应该为“*”，或为clusters（clusters.md）的llm_config.models中出现的模型
  - subnet：需要在 00-common.md 中补充CIDR的格式说明
- 对1，合法性条件
  - TPMConfig
    - name：长度 1-128 字符
    - `model`：应该为“*”，或为clusters（clusters.md）的llm_config.models中出现的模型
    - `max_tokens`：大于等于0
  - RPMConfig
    - name：长度 1-128 字符
    - `model`：应该为“*”，或为clusters（clusters.md）的llm_config.models中出现的模型
    - `max_requests`：大于等于0
  - route_rules
    - rules：应该在 00-common.md 中补充说明

# 第4轮

- 对api-keys.md的第1部分
  - `quota_plan`、`rate_limit_policy`、`route_rules`：应该在 00-common.md 中统一补充说明，而不需要在api-keys.md, entities.md、global-route-rules.md等文件中反复出现
  - 在示例中，route_rules.rules 要给一个具体的例子，不要只是空数组
- 对00-common.md
  - 对于 公共参数类型 中的  6. 路由规则（RouteRule）至  **11. RPM 限流配置（RPMConfig）**
    - 这些属于比较复杂的架构，应该给出
      - 详细的结构说明
      - 示例

# 第5轮

对00-common.md，合法性条件

- 6. 路由规则（RouteRule）
  - `targets` 元素结构
    - `ClusterName`：必须是 clusters.md中/clusters中存在的cluster
    - `Model`: 如果不为空，必须在ClusterName对应的cluster的llm_config.models中
    - 另外
      - 一个targets里的多个target，（ClusterName，Model）的组合不能重复
  - `fallbacks` 元素结构
    - `ClusterName`：必须是 clusters.md中/clusters中存在的cluster
    - `Model`: 如果不为空，必须在ClusterName对应的cluster的llm_config.models中
  
- 9. 限流规则配置（RateLimitPolicy）
  - `tpm`：同一 `RateLimitPolicy` 内，多条 TPMConfig之间，（model，window_minutes，max_tokens，step_minutes）的组合不能重复
  - rpm：同一 `RateLimitPolicy` 内，多条RPMConfig之间，（model，window_minutes，max_requests）的组合不能重复

# 第6轮

请基于以下文档，对 ai-gateway-api\design-docs\sys-design 进行修改

- ai-gateway-api\design-docs\api-define
- design-docs\modifications\2026-07-29-add-check-for-api-parameters\api-changes.md
- design-docs\modifications\2026-07-29-add-check-for-api-parameters\design-changes.md

# 第7轮

请基于以下文档，对 ai-gateway-api的代码 进行修改

- ai-gateway-api\design-docs\api-define
- design-docs\modifications\2026-07-29-add-check-for-api-parameters\api-changes.md
- design-docs\modifications\2026-07-29-add-check-for-api-parameters\design-changes.md
- ai-gateway-api\design-docs\sys-design

# 第8轮

design-docs\api-define\OpenAPI接口定义 中增加了很多 合法性条件 的说明。

请基于这些说明，对 ai-gateway-api的代码 进行修改

# 第9轮

design-docs\api-define\OpenAPI接口定义 中增加了很多 合法性条件 的说明。

请基于这些说明，对ai-gateway-api\test 进行修改，以验证 ai-gateway-api 是否已支持 合法性条件 的检查

等修改完成后，再运行 ai-gateway-api\test 中的测试例

# 第10轮

design-docs\api-define\OpenAPI接口定义 中增加了很多 合法性条件 的说明。

请基于这些说明，对ai-gateway-api\test\integration\tests各目录下的设计文档（design.md） 进行修改，以验证 ai-gateway-api 是否已支持 合法性条件 的检查

之后，再补充和修改 ai-gateway-api\test\integration\tests各目录下的测试例代码

