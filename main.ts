import { serve } from "https://deno.land/std@0.220.1/http/server.ts";

// 在文件开头添加接口定义
interface Price {
    model: string;
    type: string;
    channel_type: number;
    input: number;
    output: number;
}

// HTML 页面
const html = `<!DOCTYPE html>
<html>
<head>
    <title>AI Models Price API</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 40px auto;
            padding: 20px;
            background-color: #f7f9fc;
            color: #2c3e50;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        h1 {
            color: #1a73e8;
            margin-bottom: 30px;
            text-align: center;
            font-size: 2.2em;
        }
        .link-card {
            background-color: #f8f9fa;
            border-left: 4px solid #1a73e8;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 8px;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        .link-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(0, 0, 0, 0.1);
        }
        .link-title {
            font-weight: 600;
            color: #1a73e8;
            margin-bottom: 10px;
            font-size: 1.1em;
        }
        a {
            color: #1a73e8;
            text-decoration: none;
            word-break: break-all;
        }
        a:hover {
            text-decoration: underline;
        }
        .description {
            color: #666;
            font-size: 0.9em;
            margin-top: 10px;
        }
        footer {
            margin-top: 40px;
            text-align: center;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>AI Models Price API</h1>
        
        <div class="link-card">
            <div class="link-title">📊 模型价格表格</div>
            <a href="https://czl-logistics.feishu.cn/base/YFQhbCITwaWZblsessyctQNlnde?from=from_copylink" target="_blank">
                在飞书多维表格中查看完整价格表
            </a>
            <div class="description">
                查看所有 AI 模型的详细价格信息，包括输入输出价格、通道类型等
            </div>
        </div>

        <div class="link-card">
            <div class="link-title">🔄 JSON API 接口</div>
            <a href="https://woodchen-aimodels-price.deno.dev/api/prices" target="_blank">
                获取价格数据的 JSON 格式
            </a>
            <div class="description">
                用于程序接入的 JSON 格式数据，支持实时获取最新价格信息
            </div>
        </div>

        <div class="link-card">
            <div class="link-title">📝 提交/更新价格</div>
            <a href="https://czl-logistics.feishu.cn/share/base/form/shrcnrFG5qhUStivKiGtevuByyc" target="_blank">
                提交新的模型价格信息
            </a>
            <div class="description">
                通过飞书表单提交新的模型价格或更新现有模型的价格信息
            </div>
        </div>

        <footer>
            © ${new Date().getFullYear()} AI Models Price API - Powered by Deno Deploy
        </footer>
    </div>
</body>
</html>`;

// 使用 Deno KV 存储数据
const kv = await Deno.openKv();

// 读取价格数据
async function readPrices(): Promise<any[]> {
    const prices = await kv.get(["prices"]);
    return prices.value || [];
}

// 写入价格数据
async function writePrices(prices: any[]): Promise<void> {
    await kv.set(["prices"], prices);
}

// 修改验证函数
function validateData(data: any): string | null {
    if (!data.model || !data.type || data.channel_type === undefined || data.input === undefined || data.output === undefined) {
        return "所有字段都是必需的";
    }
    
    // 确保数字字段是数字类型
    const channel_type = Number(data.channel_type);
    const input = Number(data.input);
    const output = Number(data.output);
    
    if (isNaN(channel_type) || isNaN(input) || isNaN(output)) {
        return "数字字段格式无效";
    }
    
    // 验证数字范围（允许等于0）
    if (channel_type < 0 || input < 0 || output < 0) {
        return "数字不能小于0";
    }
    
    return null;
}

// 修改处理函数
async function handler(req: Request): Promise<Response> {
    const url = new URL(req.url);
    
    const headers = {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type",
    };

    if (req.method === "OPTIONS") {
        return new Response(null, { headers });
    }
    
    if (url.pathname === "/api/prices") {
        if (req.method === "POST") {
            try {
                let rawData;
                const contentType = req.headers.get("content-type") || "";
                
                // 获取原始数据
                if (contentType.includes("application/json")) {
                    rawData = await req.json();
                } else if (contentType.includes("application/x-www-form-urlencoded")) {
                    const formData = await req.formData();
                    rawData = {};
                    for (const [key, value] of formData.entries()) {
                        rawData[key] = value;
                    }
                } else {
                    throw new Error("不支持的内容类型");
                }
                
                console.log("Received raw data:", rawData);

                // 修改数组声明
                let dataArray: Price[] = [];
                
                // 如果数据中的字段包含逗号，说明是批量数据
                if (typeof rawData.model === 'string' && rawData.model.includes(',')) {
                    const models = rawData.model.split(',');
                    const types = rawData.type.split(',');
                    const channelTypes = rawData.channel_type.split(',');
                    const inputs = rawData.input.split(',');
                    const outputs = rawData.output.split(',');
                    
                    // 确保所有数组长度一致
                    const length = Math.min(
                        models.length,
                        types.length,
                        channelTypes.length,
                        inputs.length,
                        outputs.length
                    );
                    
                    // 构建数据数组
                    for (let i = 0; i < length; i++) {
                        if (models[i] && types[i] && channelTypes[i] && inputs[i] && outputs[i]) {
                            dataArray.push({
                                model: models[i].trim(),
                                type: types[i].trim(),
                                channel_type: Number(channelTypes[i]),
                                input: Number(inputs[i]),
                                output: Number(outputs[i])
                            });
                        }
                    }
                } else {
                    // 单条数据
                    dataArray.push({
                        model: String(rawData.model).trim(),
                        type: String(rawData.type).trim(),
                        channel_type: Number(rawData.channel_type),
                        input: Number(rawData.input),
                        output: Number(rawData.output)
                    });
                }
                
                console.log("Processed data array:", dataArray);
                
                // 验证所有数据
                const errors = [];
                const validData = [];
                
                for (const data of dataArray) {
                    const error = validateData(data);
                    if (error) {
                        errors.push({ data, error });
                    } else {
                        validData.push(data);
                    }
                }
                
                if (errors.length > 0) {
                    return new Response(JSON.stringify({ 
                        error: "部分数据验证失败",
                        details: errors
                    }), {
                        status: 400,
                        headers: { 
                            "Content-Type": "application/json",
                            ...headers 
                        }
                    });
                }
                
                // 读取现有数据
                const prices = await readPrices();
                
                // 添加新数据
                prices.push(...validData);
                
                // 保存数据
                await writePrices(prices);
                
                return new Response(JSON.stringify({ 
                    success: true,
                    processed: validData.length,
                    data: validData
                }), {
                    headers: { 
                        "Content-Type": "application/json",
                        ...headers 
                    }
                });
            } catch (error) {
                console.error("Processing error:", error);
                return new Response(JSON.stringify({ 
                    error: error.message,
                    details: "数据处理失败，请检查输入格式"
                }), {
                    status: 500,
                    headers: { 
                        "Content-Type": "application/json",
                        ...headers 
                    }
                });
            }
        } else if (req.method === "GET") {
            const prices = await readPrices();
            return new Response(JSON.stringify(prices), {
                headers: { 
                    "Content-Type": "application/json",
                    ...headers 
                }
            });
        }
    }
    
    // 提供静态页面
    if (url.pathname === "/" || url.pathname === "/index.html") {
        return new Response(html, {
            headers: { 
                "Content-Type": "text/html; charset=utf-8",
                ...headers 
            }
        });
    }
    
    return new Response("Not Found", { 
        status: 404,
        headers 
    });
}

// 启动服务器
serve(handler); 