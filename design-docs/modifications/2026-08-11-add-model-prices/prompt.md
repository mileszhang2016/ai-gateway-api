# 1

- 请基于 document-ai-gateway\迭代系统设计\v0.4\quota-rmb-support\quota-rmb-support-design.md
  - 在 ai-gateway-api\design-docs\modifications\2026-08-11-add-model-prices 中生成 api-changes.md
- 2.3.2
  - 对于 capabilities 和 supported_parameters，也应该为给定的枚举值
- 2.3.2
  - limits、prices、metadata中，可包含的字段，也应该是给定的枚举值
  - provider_specific_entry，可以暂时去掉

- 对于第5部分
  - 对/clusters端点，不会展示 `model_table`
  - model_table会在innerapi的ClusterConf的AIConf中出现

# 2

针对 ai-gateway-api\test

- 首先，对存量的测试例进行修正
- 然后，针对本次修改的功能，增加新的测试例
  - 在新增测试例前，请首先修改/编写测试说明文档（design.md）
  - 我review后，你再开始写测试例