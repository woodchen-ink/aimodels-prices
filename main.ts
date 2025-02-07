import { serve } from "https://deno.land/std/http/server.ts";

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
                <input type="number" id="channel_type" name="channel_type" required>
            </div>
            <div class="form-group">
                <label for="input">Input Price:</label>
                <input type="number" id="input" name="input" step="0.0001" required>
            </div>
            <div class="form-group">
                <label for="output">Output Price:</label>
                <input type="number" id="output" name="output" step="0.0001" required>
            </div>
            <button type="submit">Update Price</button>
            <div id="message"></div>
        </form>
    </div>

    <script>
        document.getElementById('priceForm').onsubmit = async (e) => {
            e.preventDefault();
            const messageDiv = document.getElementById('message');
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData.entries());
            
            try {
                const response = await fetch('/api/prices', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    messageDiv.className = 'message success';
                    messageDiv.textContent = '价格更新成功！';
                    e.target.reset();
                } else {
                    messageDiv.className = 'message error';
                    messageDiv.textContent = result.error || '更新失败';
                }
            } catch (error) {
                messageDiv.className = 'message error';
                messageDiv.textContent = '更新失败: ' + error.message;
            }
        };
    </script>
</body>
</html>`;

// 读取prices.json文件
async function readPrices(): Promise<any[]> {
    try {
        const decoder = new TextDecoder("utf-8");
        const data = await Deno.readFile("prices.json");
        return JSON.parse(decoder.decode(data));
    } catch {
        return [];
    }
}

// 写入prices.json文件
async function writePrices(prices: any[]): Promise<void> {
    const encoder = new TextEncoder();
    await Deno.writeFile("prices.json", encoder.encode(JSON.stringify(prices, null, 2)));
}

// 验证请求数据
function validateData(data: any): string | null {
    if (!data.model || !data.type || !data.channel_type || !data.input || !data.output) {
        return "所有字段都是必需的";
    }
    
    if (isNaN(data.channel_type) || isNaN(data.input) || isNaN(data.output)) {
        return "数字字段格式无效";
    }
    
    return null;
}

// 处理请求
async function handler(req: Request): Promise<Response> {
    const url = new URL(req.url);
    
    // 提供静态页面
    if (url.pathname === "/" || url.pathname === "/index.html") {
        return new Response(html, {
            headers: { "Content-Type": "text/html; charset=utf-8" }
        });
    }
    
    // API 端点
    if (url.pathname === "/api/prices") {
        if (req.method === "POST") {
            try {
                const data = await req.json();
                
                // 验证数据
                const error = validateData(data);
                if (error) {
                    return new Response(JSON.stringify({ error }), {
                        status: 400,
                        headers: { "Content-Type": "application/json" }
                    });
                }
                
                // 转换数据类型
                const newPrice = {
                    model: data.model,
                    type: data.type,
                    channel_type: parseInt(data.channel_type),
                    input: parseFloat(data.input),
                    output: parseFloat(data.output)
                };
                
                // 读取现有数据
                const prices = await readPrices();
                
                // 添加新数据
                prices.push(newPrice);
                
                // 保存数据
                await writePrices(prices);
                
                return new Response(JSON.stringify({ success: true }), {
                    headers: { "Content-Type": "application/json" }
                });
            } catch (error) {
                return new Response(JSON.stringify({ error: error.message }), {
                    status: 500,
                    headers: { "Content-Type": "application/json" }
                });
            }
        } else if (req.method === "GET") {
            const prices = await readPrices();
            return new Response(JSON.stringify(prices), {
                headers: { "Content-Type": "application/json" }
            });
        }
    }
    
    return new Response("Not Found", { status: 404 });
}

// 启动服务器
serve(handler); 