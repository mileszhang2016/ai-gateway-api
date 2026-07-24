-- pools 表初始化数据
INSERT INTO `pools` (
    `id`,
    `name`,
    `product_id`,
    `ready`,
    `instance_detail`,
    `type`,
    `tag`,
    `role`,
    `created_at`,
    `updated_at`
) VALUES (
    1,
    'BFE.aipool',
    1,
    1,
    '[{"Name":"127.0.0.1","Addr":"127.0.0.1","Port":8080,"Ports":{"Default":8080},"tags":{"key":"value"},"Weight":1,"Disable":false}]',
    1,
    1,
    'COMMON',
    NOW(),
    NOW()
);

-- bfe_clusters 表初始化数据
INSERT INTO `bfe_clusters` (
    `id`,
    `name`,
    `pool_name`,
    `capacity`,
    `enabled`,
    `gtc_enabled`,
    `gtc_manual_enabled`,
    `exempt_traffic_check`,
    `created_at`,
    `updated_at`
) VALUES (
    1,
    'BFE-AI_product.szyf',
    'BFE.aipool',
    0,
    1,
    1,
    1,
    0,
    NOW(),
    NOW()
);
