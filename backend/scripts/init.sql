-- 初始化模型类型数据
INSERT INTO model_type (type_key, type_label) VALUES
('text2text', '文生文'),
('text2image', '文生图'),
('text2speech', '文生音'),
('speech2text', '音生文'),
('image2text', '图生文'),
('embedding', '向量'),
('other', '其他')
ON DUPLICATE KEY UPDATE type_label = VALUES(type_label); 