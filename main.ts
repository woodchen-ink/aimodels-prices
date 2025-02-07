import { serve } from "https://deno.land/std@0.220.1/http/server.ts";
import { createHmac } from "crypto";

// 类型定义
interface Vendor {
    id: number;
    name: string;
    icon: string;
}

interface VendorResponse {
    data: {
        [key: string]: Vendor;
    };
}

interface Price {
    id?: string;
    model: string;
    billing_type: 'tokens' | 'times';
    channel_type: number;
    currency: 'CNY' | 'USD';
    input_price: number;
    output_price: number;
    input_ratio: number;
    output_ratio: number;
    price_source: string;
    status: 'pending' | 'approved' | 'rejected';
    created_by: string;
    created_at: string;
    reviewed_by?: string;
    reviewed_at?: string;
}

// 缓存供应商数据
let vendorsCache: VendorResponse | null = null;
let vendorsCacheTime: number = 0;
const CACHE_DURATION = 1000 * 60 * 5; // 5分钟缓存

// 获取供应商数据
async function getVendors(): Promise<VendorResponse> {
    const now = Date.now();
    if (vendorsCache && (now - vendorsCacheTime) < CACHE_DURATION) {
        return vendorsCache;
    }

    try {
        const response = await fetch('https://oapi.czl.net/api/ownedby');
        const data = await response.json() as VendorResponse;
        vendorsCache = data;
        vendorsCacheTime = now;
        return data;
    } catch (error) {
        console.error('获取供应商数据失败:', error);
        throw new Error('获取供应商数据失败');
    }
}

// 计算倍率
function calculateRatio(price: number, currency: 'CNY' | 'USD'): number {
    return currency === 'USD' ? price / 2 : price / 14;
}

// 验证价格数据
function validatePrice(data: any): string | null {
    if (!data.model || !data.billing_type || !data.channel_type || 
        !data.currency || data.input_price === undefined || data.output_price === undefined ||
        !data.price_source) {
        return "所有字段都是必需的";
    }

    if (data.billing_type !== 'tokens' && data.billing_type !== 'times') {
        return "计费类型必须是 tokens 或 times";
    }

    if (data.currency !== 'CNY' && data.currency !== 'USD') {
        return "币种必须是 CNY 或 USD";
    }

    if (isNaN(data.input_price) || isNaN(data.output_price)) {
        return "价格必须是数字";
    }

    if (data.input_price < 0 || data.output_price < 0) {
        return "价格不能为负数";
    }

    return null;
}

// 添加 Discourse SSO 配置
const DISCOURSE_URL = Deno.env.get('DISCOURSE_URL') || 'https://discourse.czl.net';
const DISCOURSE_SSO_SECRET = Deno.env.get('DISCOURSE_SSO_SECRET');

// 验证必需的环境变量
if (!DISCOURSE_SSO_SECRET) {
    console.error('错误: 必须设置 DISCOURSE_SSO_SECRET 环境变量');
    Deno.exit(1);
}

// 添加认证相关函数
async function verifyDiscourseSSO(request: Request): Promise<string | null> {
    const cookie = request.headers.get('cookie');
    if (!cookie) return null;

    const sessionMatch = cookie.match(/session=([^;]+)/);
    if (!sessionMatch) return null;

    const sessionId = sessionMatch[1];
    const session = await kv.get(['sessions', sessionId]);

    if (!session.value) return null;
    return session.value.username;
}

// 添加登录和登出函数
function generateSSO(returnUrl: string): string {
    const payload = Buffer.from(`return_sso_url=${encodeURIComponent(returnUrl)}`).toString('base64');
    const sig = createHmac('sha256', DISCOURSE_SSO_SECRET)
        .update(payload)
        .digest('hex');
    return `${DISCOURSE_URL}/session/sso_provider?sso=${encodeURIComponent(payload)}&sig=${sig}`;
}

// HTML 页面
const html = `<!DOCTYPE html>
<html>
<head>
    <title>AI Models Price API</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/@fortawesome/fontawesome-free@6.0.0/css/all.min.css" rel="stylesheet">
    <link href="https://cdn.jsdelivr.net/npm/animate.css@4.1.1/animate.min.css" rel="stylesheet">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            padding: 20px;
            background-color: #f7f9fc;
            color: #2c3e50;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .card {
            background-color: white;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            margin-bottom: 20px;
        }
        .vendor-icon {
            width: 24px;
            height: 24px;
            margin-right: 8px;
            vertical-align: middle;
        }
        .badge {
            font-size: 0.8em;
            padding: 5px 10px;
        }
        .badge-tokens {
            background-color: #4CAF50;
        }
        .badge-times {
            background-color: #2196F3;
        }
        .badge-pending {
            background-color: #FFC107;
        }
        .badge-approved {
            background-color: #4CAF50;
        }
        .badge-rejected {
            background-color: #F44336;
        }
        .table th {
            white-space: nowrap;
        }
        .source-link {
            max-width: 200px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            display: inline-block;
        }
        #loginStatus {
            margin-bottom: 20px;
        }
        .animate__animated {
            animation-duration: 0.5s;
        }
        .table-hover tbody tr:hover {
            background-color: rgba(26, 115, 232, 0.05);
            transition: background-color 0.2s ease;
        }
        .nav-tabs .nav-link {
            border: none;
            color: #666;
            padding: 1rem 1.5rem;
            font-weight: 500;
            transition: all 0.2s ease;
        }
        .nav-tabs .nav-link.active {
            color: #1a73e8;
            border-bottom: 2px solid #1a73e8;
            background: none;
        }
        .loading-spinner {
            width: 3rem;
            height: 3rem;
        }
        .toast {
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 1050;
        }
        .vendor-icon {
            transition: transform 0.2s ease;
        }
        .vendor-icon:hover {
            transform: scale(1.2);
        }
        .badge {
            transition: all 0.2s ease;
        }
        .badge:hover {
            transform: translateY(-1px);
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
    </style>
</head>
<body>
    <div class="container">
        <nav class="navbar navbar-expand-lg navbar-light bg-white rounded shadow-sm mb-4">
            <div class="container-fluid">
                <a class="navbar-brand" href="#">
                    <i class="fas fa-robot text-primary me-2"></i>
                    AI Models Price
                </a>
                <div class="d-flex align-items-center">
                    <div id="loginStatus"></div>
                </div>
            </div>
        </nav>

        <div class="row">
            <div class="col-12">
                <ul class="nav nav-tabs mb-4">
                    <li class="nav-item">
                        <a class="nav-link active" href="#prices" data-bs-toggle="tab">
                            <i class="fas fa-table me-2"></i>价格列表
                        </a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" href="#submit" data-bs-toggle="tab">
                            <i class="fas fa-plus-circle me-2"></i>提交价格
                        </a>
                    </li>
                </ul>

                <div class="tab-content">
                    <div class="tab-pane fade show active" id="prices">
                        <div class="card shadow-sm">
                            <div class="card-body">
                                <div class="table-responsive">
                                    <table class="table table-hover align-middle">
                                        <thead class="table-light">
                                            <tr>
                                                <th>模型名称</th>
                                                <th>计费类型</th>
                                                <th>供应商</th>
                                                <th>币种</th>
                                                <th>输入价格(M)</th>
                                                <th>输出价格(M)</th>
                                                <th>输入倍率</th>
                                                <th>输出倍率</th>
                                                <th>价格依据</th>
                                                <th>状态</th>
                                                <th>操作</th>
                                            </tr>
                                        </thead>
                                        <tbody id="priceTable">
                                            <tr>
                                                <td colspan="11" class="text-center py-5">
                                                    <div class="spinner-border text-primary loading-spinner" role="status">
                                                        <span class="visually-hidden">加载中...</span>
                                                    </div>
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="tab-pane fade" id="submit">
                        <div class="card shadow-sm">
                            <div class="card-body">
                                <form id="newPriceForm" class="needs-validation" novalidate>
                                    <div class="row">
                                        <div class="col-md-6 mb-3">
                                            <label class="form-label">模型名称</label>
                                            <input type="text" class="form-control" name="model" required>
                                        </div>
                                        <div class="col-md-6 mb-3">
                                            <label class="form-label">计费类型</label>
                                            <select class="form-select" name="billing_type" required>
                                                <option value="tokens">按量计费(tokens)</option>
                                                <option value="times">按次计费(times)</option>
                                            </select>
                                        </div>
                                        <div class="col-md-6 mb-3">
                                            <label class="form-label">供应商</label>
                                            <select class="form-select" name="channel_type" required>
                                                <option value="">选择供应商...</option>
                                            </select>
                                        </div>
                                        <div class="col-md-6 mb-3">
                                            <label class="form-label">币种</label>
                                            <select class="form-select" name="currency" required>
                                                <option value="CNY">人民币</option>
                                                <option value="USD">美元</option>
                                            </select>
                                        </div>
                                        <div class="col-md-6 mb-3">
                                            <label class="form-label">输入价格(M)</label>
                                            <input type="number" class="form-control" name="input_price" step="0.0001" min="0" required>
                                        </div>
                                        <div class="col-md-6 mb-3">
                                            <label class="form-label">输出价格(M)</label>
                                            <input type="number" class="form-control" name="output_price" step="0.0001" min="0" required>
                                        </div>
                                        <div class="col-12 mb-3">
                                            <label class="form-label">价格依据(官方文档链接)</label>
                                            <input type="url" class="form-control" name="price_source" required>
                                        </div>
                                    </div>
                                    <button type="submit" class="btn btn-primary">提交价格</button>
                                </form>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Toast 提示 -->
        <div class="toast-container">
            <div class="toast align-items-center text-white bg-success border-0" role="alert" aria-live="assertive" aria-atomic="true">
                <div class="d-flex">
                    <div class="toast-body"></div>
                    <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        let currentUser = null;
        let vendors = null;

        // 检查登录状态
        async function checkLoginStatus() {
            try {
                const response = await fetch('/api/auth/status');
                const data = await response.json();
                currentUser = data.user;
                updateLoginUI();
            } catch (error) {
                console.error('检查登录状态失败:', error);
            }
        }

        // 更新登录UI
        function updateLoginUI() {
            const loginStatus = document.getElementById('loginStatus');
            const submitForm = document.getElementById('submitForm');
            
            if (currentUser) {
                loginStatus.innerHTML = \`
                    <span class="me-2">欢迎, \${currentUser}</span>
                    <button onclick="logout()" class="btn btn-outline-danger btn-sm">退出</button>
                \`;
                submitForm.style.display = 'block';
            } else {
                loginStatus.innerHTML = '<button onclick="login()" class="btn btn-primary btn-sm">通过 Discourse 登录</button>';
                submitForm.style.display = 'none';
            }
        }

        // 加载供应商数据
        async function loadVendors() {
            try {
                const response = await fetch('https://oapi.czl.net/api/ownedby');
                const data = await response.json();
                vendors = data.data;
                
                // 更新供应商选择框
                const select = document.querySelector('select[name="channel_type"]');
                Object.entries(vendors).forEach(([id, vendor]) => {
                    const option = document.createElement('option');
                    option.value = id;
                    option.textContent = vendor.name;
                    select.appendChild(option);
                });
            } catch (error) {
                console.error('加载供应商数据失败:', error);
            }
        }

        // 加载价格数据
        async function loadPrices() {
            try {
                const response = await fetch('/api/prices');
                const prices = await response.json();
                const tbody = document.getElementById('priceTable');
                tbody.innerHTML = '';

                prices.forEach(price => {
                    const vendor = vendors[price.channel_type];
                    const tr = document.createElement('tr');
                    tr.innerHTML = \`
                        <td>\${price.model}</td>
                        <td><span class="badge badge-\${price.billing_type}">\${price.billing_type === 'tokens' ? '按量计费' : '按次计费'}</span></td>
                        <td>
                            <img src="\${vendor?.icon}" class="vendor-icon" alt="\${vendor?.name}">
                            \${vendor?.name || '未知供应商'}
                        </td>
                        <td>\${price.currency}</td>
                        <td>\${price.input_price}</td>
                        <td>\${price.output_price}</td>
                        <td>\${price.input_ratio.toFixed(4)}</td>
                        <td>\${price.output_ratio.toFixed(4)}</td>
                        <td><a href="\${price.price_source}" target="_blank" class="source-link">查看来源</a></td>
                        <td><span class="badge badge-\${price.status}">\${price.status}</span></td>
                        <td>
                            \${currentUser === 'wood' && price.status === 'pending' ? \`
                                <button onclick="reviewPrice('\${price.id}', 'approved')" class="btn btn-success btn-sm">通过</button>
                                <button onclick="reviewPrice('\${price.id}', 'rejected')" class="btn btn-danger btn-sm">拒绝</button>
                            \` : ''}
                        </td>
                    \`;
                    tbody.appendChild(tr);
                });
            } catch (error) {
                console.error('加载价格数据失败:', error);
            }
        }

        // 提交新价格
        document.getElementById('newPriceForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData.entries());
            
            try {
                const response = await fetch('/api/prices', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                
                if (response.ok) {
                    alert('提交成功，等待审核');
                    e.target.reset();
                    loadPrices();
                } else {
                    const error = await response.json();
                    alert(error.message || '提交失败');
                }
            } catch (error) {
                console.error('提交价格失败:', error);
                alert('提交失败');
            }
        };

        // 审核价格
        async function reviewPrice(id, status) {
            try {
                const response = await fetch(\`/api/prices/\${id}/review\`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ status })
                });
                
                if (response.ok) {
                    alert('审核成功');
                    loadPrices();
                } else {
                    const error = await response.json();
                    alert(error.message || '审核失败');
                }
            } catch (error) {
                console.error('审核价格失败:', error);
                alert('审核失败');
            }
        }

        // 初始化
        async function init() {
            await Promise.all([
                checkLoginStatus(),
                loadVendors()
            ]);
            loadPrices();
        }

        // 添加 Toast 提示函数
        function showToast(message, type = 'success') {
            const toast = document.querySelector('.toast');
            toast.className = \`toast align-items-center text-white bg-\${type} border-0\`;
            toast.querySelector('.toast-body').textContent = message;
            const bsToast = new bootstrap.Toast(toast);
            bsToast.show();
        }

        // 登录函数
        function login() {
            const returnUrl = \`\${window.location.origin}/auth/callback\`;
            window.location.href = \`/api/auth/login?return_url=\${encodeURIComponent(returnUrl)}\`;
        }

        // 登出函数
        async function logout() {
            try {
                await fetch('/api/auth/logout', { method: 'POST' });
                window.location.reload();
            } catch (error) {
                console.error('登出失败:', error);
                showToast('登出失败', 'danger');
            }
        }

        init();
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

// 修改处理函数
async function handler(req: Request): Promise<Response> {
    const url = new URL(req.url);
    
    const headers = {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type, Cookie",
        "Access-Control-Allow-Credentials": "true"
    };

    if (req.method === "OPTIONS") {
        return new Response(null, { headers });
    }

    // 登录处理
    if (url.pathname === "/api/auth/login") {
        const params = new URLSearchParams(url.search);
        const returnUrl = params.get('return_url');
        if (!returnUrl) {
            return new Response(JSON.stringify({ error: "缺少 return_url 参数" }), {
                status: 400,
                headers: {
                    "Content-Type": "application/json",
                    ...headers
                }
            });
        }

        const ssoUrl = generateSSO(returnUrl);
        return new Response(null, {
            status: 302,
            headers: {
                ...headers,
                "Location": ssoUrl
            }
        });
    }

    // SSO 回调处理
    if (url.pathname === "/auth/callback") {
        const params = new URLSearchParams(url.search);
        const sso = params.get('sso');
        const sig = params.get('sig');

        if (!sso || !sig) {
            return new Response("Invalid SSO parameters", { 
                status: 400,
                headers: {
                    "Content-Type": "text/plain",
                    ...headers
                }
            });
        }

        try {
            // 验证签名
            const expectedSig = createHmac('sha256', DISCOURSE_SSO_SECRET)
                .update(sso)
                .digest('hex');

            if (sig !== expectedSig) {
                throw new Error('Invalid signature');
            }

            // 解码 payload
            const payload = Buffer.from(sso, 'base64').toString();
            const payloadParams = new URLSearchParams(payload);
            const username = payloadParams.get('username');

            if (!username) {
                throw new Error('Missing username');
            }

            // 设置 session cookie
            const sessionId = crypto.randomUUID();
            await kv.set(['sessions', sessionId], { 
                username,
                created_at: new Date().toISOString()
            }, { expireIn: 24 * 60 * 60 * 1000 }); // 24小时过期

            return new Response(null, {
                status: 302,
                headers: {
                    ...headers,
                    "Location": "/",
                    "Set-Cookie": `session=${sessionId}; Path=/; HttpOnly; SameSite=Lax; Max-Age=86400`
                }
            });
        } catch (error) {
            console.error('SSO 回调处理失败:', error);
            return new Response("SSO verification failed", { 
                status: 400,
                headers: {
                    "Content-Type": "text/plain",
                    ...headers
                }
            });
        }
    }

    // 登出处理
    if (url.pathname === "/api/auth/logout" && req.method === "POST") {
        const cookie = req.headers.get('cookie');
        if (cookie) {
            const sessionMatch = cookie.match(/session=([^;]+)/);
            if (sessionMatch) {
                const sessionId = sessionMatch[1];
                await kv.delete(['sessions', sessionId]);
            }
        }

        return new Response(JSON.stringify({ success: true }), {
            headers: {
                ...headers,
                "Content-Type": "application/json",
                "Set-Control-Allow-Credentials": "true",
                "Set-Cookie": "session=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0"
            }
        });
    }

    // 认证状态检查
    if (url.pathname === "/api/auth/status") {
        const username = await verifyDiscourseSSO(req);
        return new Response(JSON.stringify({ 
            authenticated: !!username,
            user: username
        }), {
            headers: { 
                "Content-Type": "application/json",
                ...headers 
            }
        });
    }

    // 价格审核
    if (url.pathname.match(/^\/api\/prices\/\d+\/review$/)) {
        const username = await verifyDiscourseSSO(req);
        if (!username || username !== 'wood') {
            return new Response(JSON.stringify({ error: "未授权" }), {
                status: 403,
                headers: { 
                    "Content-Type": "application/json",
                    ...headers 
                }
            });
        }

        if (req.method === "POST") {
            try {
                const id = url.pathname.split('/')[3];
                const { status } = await req.json();
                
                if (status !== 'approved' && status !== 'rejected') {
                    throw new Error("无效的状态");
                }

                const prices = await readPrices();
                const priceIndex = prices.findIndex(p => p.id === id);
                
                if (priceIndex === -1) {
                    throw new Error("价格记录不存在");
                }

                prices[priceIndex].status = status;
                prices[priceIndex].reviewed_by = username;
                prices[priceIndex].reviewed_at = new Date().toISOString();

                await writePrices(prices);

                return new Response(JSON.stringify({ success: true }), {
                    headers: { 
                        "Content-Type": "application/json",
                        ...headers 
                    }
                });
            } catch (error) {
                return new Response(JSON.stringify({ 
                    error: error.message || "审核失败"
                }), {
                    status: 400,
                    headers: { 
                        "Content-Type": "application/json",
                        ...headers 
                    }
                });
            }
        }
    }

    // 提交新价格
    if (url.pathname === "/api/prices" && req.method === "POST") {
        const username = await verifyDiscourseSSO(req);
        if (!username) {
            return new Response(JSON.stringify({ error: "请先登录" }), {
                status: 401,
                headers: { 
                    "Content-Type": "application/json",
                    ...headers 
                }
            });
        }

        try {
            let rawData;
            const contentType = req.headers.get("content-type") || "";
            
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

            // 处理数据
            const newPrice: Price = {
                model: String(rawData.model).trim(),
                billing_type: rawData.billing_type as 'tokens' | 'times',
                channel_type: Number(rawData.channel_type),
                currency: rawData.currency as 'CNY' | 'USD',
                input_price: Number(rawData.input_price),
                output_price: Number(rawData.output_price),
                input_ratio: calculateRatio(Number(rawData.input_price), rawData.currency as 'CNY' | 'USD'),
                output_ratio: calculateRatio(Number(rawData.output_price), rawData.currency as 'CNY' | 'USD'),
                price_source: String(rawData.price_source),
                status: 'pending',
                created_by: username,
                created_at: new Date().toISOString()
            };

            // 验证数据
            const error = validatePrice(newPrice);
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
            
            // 生成唯一ID
            newPrice.id = Date.now().toString();
            
            // 添加新数据
            prices.push(newPrice);
            
            // 保存数据
            await writePrices(prices);
            
            return new Response(JSON.stringify({ 
                success: true,
                data: newPrice
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
    }

    // 获取价格列表
    if (url.pathname === "/api/prices" && req.method === "GET") {
        const prices = await readPrices();
        return new Response(JSON.stringify(prices), {
            headers: { 
                "Content-Type": "application/json",
                ...headers 
            }
        });
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