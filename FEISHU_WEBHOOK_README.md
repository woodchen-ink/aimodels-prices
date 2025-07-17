# 飞书Webhook通知功能

本系统支持通过飞书Webhook发送待审核价格的通知，帮助管理员及时了解需要处理的价格审核请求。

## 功能特性

- 🕐 **定期检查**：每5分钟自动检查一次待审核价格
- 📊 **智能汇总**：按厂商分组统计待审核价格数量
- 🔔 **智能通知**：30分钟内不重复发送相同通知，避免打扰
- 📋 **详细信息**：显示最近的5个待审核价格详情
- 🎨 **美观卡片**：使用飞书卡片格式，信息展示清晰美观

## 配置方法

### 1. 创建飞书自定义机器人

1. 在飞书群聊中，点击右上角设置按钮
2. 选择"群机器人" -> "添加机器人" -> "自定义机器人"
3. 填写机器人名称（如：AI模型价格通知）
4. 复制生成的Webhook地址

### 2. 配置环境变量

#### 方法一：Docker Compose配置

编辑 `docker-compose.yml` 文件：

```yaml
services:
  aimodels-prices:
    environment:
      - FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url
```

#### 方法二：环境变量文件

在 `data/.env` 文件中添加：

```bash
FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url
```

#### 方法三：系统环境变量

```bash
export FEISHU_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-url
```

### 3. 重启服务

配置完成后重启应用：

```bash
docker-compose down
docker-compose up -d
```

## 通知内容

### 通知时机

- 系统每5分钟检查一次待审核价格
- 只有当存在待审核价格时才发送通知
- 30分钟内不会重复发送相同的通知

### 通知内容

通知卡片包含以下信息：

1. **总计统计**：待审核价格总数
2. **分厂商统计**：按厂商分组的价格数量
3. **价格详情**：最近的5个待审核价格信息，包括：
   - 模型名称
   - 所属厂商
   - 创建者
4. **操作提醒**：提示管理员及时处理

### 示例通知

```
🔍 待审核价格检查报告 - 8个待审核

📋 待审核价格统计

总计： 8 个模型价格待审核

分厂商统计：
- OpenAI：3 个模型
- Anthropic：2 个模型
- 字节跳动：3 个模型

最近待审核价格（最多显示5个）：
1. gpt-4o-mini (OpenAI) - 创建者：张三
2. claude-3-sonnet (Anthropic) - 创建者：李四
3. doubao-pro-4k (字节跳动) - 创建者：王五
4. gpt-4o (OpenAI) - 创建者：赵六
5. claude-3-haiku (Anthropic) - 创建者：钱七

...还有 3 个价格等待审核

⏰ 请及时处理待审核价格！
```

## 功能说明

### 自动化检查

- 定时任务每5分钟运行一次
- 自动查询数据库中状态为 `pending` 的价格记录
- 如果没有待审核价格，不会发送通知

### 防止打扰

- 系统记录上次发送通知的时间
- 30分钟内不会重复发送相同内容的通知
- 避免频繁通知造成打扰

### 异步处理

- 通知发送采用异步方式
- 不会阻塞主要业务流程
- 即使通知发送失败，也不影响系统正常运行

## 故障排除

### 通知未收到

1. **检查配置**：确认 `FEISHU_WEBHOOK_URL` 环境变量已正确设置
2. **检查网络**：确认服务器能够访问飞书API
3. **检查日志**：查看应用日志中是否有错误信息
4. **检查机器人**：确认飞书群中的机器人未被移除

### 查看日志

```bash
# 查看容器日志
docker-compose logs -f aimodels-prices

# 查看最近的日志
docker-compose logs --tail=100 aimodels-prices
```

### 常见错误

- `webhook returned status code: 400`：Webhook地址错误或格式不正确
- `failed to send webhook: connection refused`：网络连接问题
- `未配置飞书webhook，跳过通知`：环境变量未设置

## 安全说明

- Webhook URL包含敏感token，请妥善保管
- 建议定期更换Webhook URL
- 不要在公开的代码仓库中暴露Webhook URL

## 技术实现

- 使用Go的cron库实现定时任务
- 采用飞书卡片格式发送美观的通知
- 支持异步发送，不阻塞主流程
- 智能去重，避免重复通知 