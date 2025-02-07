import { serve } from "https://deno.land/std@0.220.1/http/server.ts";

// HTML 页面
const html = `<!DOCTYPE html>
<html>
<head>
    <title>AI Models Price Update</title>
    <style>
        body { font-family: Arial; max-width: 800px; margin: 20px auto; padding: 20px; background: #f5f5f5; }
        .container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        button { background: #4CAF50; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #45a049; }
        .message { margin-top: 10px; padding: 10px; border-radius: 4px; }
        .error { background: #ffebee; color: #c62828; }
        .success { background: #e8f5e9; color: #2e7d32; }
        pre { background: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>AI Models Price Update</h1>
        <form id="priceForm">
            <div class="form-group">
                <label for="model">Model Name:</label>
                <input type="text" id="model" name="model" required>
            </div>
            <div class="form-group">
                <label for="type">Type:</label>
                <input type="text" id="type" name="type" required>
            </div>
            <div class="form-group">
                <label for="channel_type">Channel Type:</label>
                <input type="number" id="channel_type" name="channel_type" min="0" step="1" required>
            </div>
            <div class="form-group">
                <label for="input">Input Price:</label>
                <input type="number" id="input" name="input" min="0" step="0.0001" required>
            </div>
            <div class="form-group">
                <label for="output">Output Price:</label>
                <input type="number" id="output" name="output" min="0" step="0.0001" required>
            </div>
            <button type="submit">Update Price</button>
            <div id="message"></div>
            <div id="debug" style="margin-top: 20px;"></div>
        </form>
    </div>

    <script>
        document.getElementById('priceForm').onsubmit = async (e) => {
            e.preventDefault();
            const messageDiv = document.getElementById('message');
            const debugDiv = document.getElementById('debug');
            
            try {
                // 获取表单数据
                const data = {
                    model: document.getElementById('model').value.trim(),
                    type: document.getElementById('type').value.trim(),
                    channel_type: Number(document.getElementById('channel_type').value),
                    input: Number(document.getElementById('input').value),
                    output: Number(document.getElementById('output').value)
                };

                // 显示发送的数据（调试用）
                debugDiv.innerHTML = '<strong>发送的数据:</strong><pre>' + JSON.stringify(data, null, 2) + '</pre>';
                
                const response = await fetch('/api/prices', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                // 显示服务器响应（调试用）
                debugDiv.innerHTML += '<strong>服务器响应:</strong><pre>' + JSON.stringify(result, null, 2) + '</pre>';
                
                if (response.ok) {
                    messageDiv.className = 'message success';
                    messageDiv.textContent = '价格更新成功！';
                    e.target.reset();
                    debugDiv.innerHTML = ''; // 清除调试信息
                } else {
                    messageDiv.className = 'message error';
                    messageDiv.textContent = result.error || '更新失败';
                }
            } catch (error) {
                console.error('Error:', error);
                messageDiv.className = 'message error';
                messageDiv.textContent = '更新失败: ' + error.message;
            }
        };
    </script>
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

// 处理请求
async function handler(req: Request): Promise<Response> {
    const url = new URL(req.url);
    
    // 添加 CORS 支持
    const headers = {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type",
    };

    // 处理 OPTIONS 请求
    if (req.method === "OPTIONS") {
        return new Response(null, { headers });
    }
    
    // API 端点
    if (url.pathname === "/api/prices") {
        if (req.method === "POST") {
            try {
                let data;
                const contentType = req.headers.get("content-type") || "";
                
                if (contentType.includes("application/json")) {
                    data = await req.json();
                } else if (contentType.includes("application/x-www-form-urlencoded")) {
                    const formData = await req.formData();
                    data = {};
                    for (const [key, value] of formData.entries()) {
                        // 如果值包含逗号，只取第一个值
                        const actualValue = String(value).split(',')[0];
                        data[key] = actualValue;
                    }
                } else {
                    throw new Error("不支持的内容类型");
                }
                
                console.log("Received raw data:", data); // 调试日志
                
                // 清理和转换数据
                const cleanData = {
                    model: String(data.model).trim(),
                    type: String(data.type).trim(),
                    channel_type: Number(String(data.channel_type).split(',')[0]),
                    input: Number(String(data.input).split(',')[0]),
                    output: Number(String(data.output).split(',')[0])
                };
                
                console.log("Cleaned data:", cleanData); // 调试日志
                
                // 验证数据
                const error = validateData(cleanData);
                if (error) {
                    return new Response(JSON.stringify({ error }), {
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
                prices.push(cleanData);
                
                // 保存数据
                await writePrices(prices);
                
                return new Response(JSON.stringify({ 
                    success: true,
                    data: cleanData 
                }), {
                    headers: { 
                        "Content-Type": "application/json",
                        ...headers 
                    }
                });
            } catch (error) {
                console.error("Processing error:", error); // 调试日志
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