# 基金查询API服务

基于Go语言开发的基金信息查询后端服务,支持获取基金详细信息和实时估值。

## 项目结构

```
fund/
├── main.go              # 程序入口
├── go.mod               # Go模块配置
├── model/               # 数据模型层
│   └── fund.go         # 基金数据模型
├── service/             # 业务逻辑层
│   └── fund_service.go # 基金服务
├── handler/             # 处理器层(控制器)
│   └── fund_handler.go # 基金接口处理器
├── router/              # 路由层
│   └── router.go       # 路由配置
├── middleware/          # 中间件层
│   └── cors.go         # 跨域中间件
└── README.md           # 项目文档
```

## 分层说明

### 1. Model 层 (`model/`)
- 定义数据结构和实体
- 只包含数据定义,不包含业务逻辑

### 2. Service 层 (`service/`)
- 业务逻辑处理
- 数据获取和解析
- 与外部API交互

### 3. Handler 层 (`handler/`)
- HTTP请求处理
- 参数验证
- 响应格式化

### 4. Router 层 (`router/`)
- 路由注册和管理
- URL路径映射

### 5. Middleware 层 (`middleware/`)
- 中间件逻辑(如CORS、日志、认证等)

## 功能特性

- 📊 获取基金详细信息(净值、增长率等)
- 📈 获取基金实时估值数据
- 📉 获取基金走势数据(支持周/月/季度等多种周期)
- 🔥 **日内实时数据采集** - 每30秒采集所有基金实时估值
- 📋 获取所有基金列表(26000+只基金)
- 🚀 简单易用的RESTful API
- ⚡ 高性能Go语言实现
- 🏗️ 清晰的分层架构

## 数据来源

- 基金详情: 东方财富网 `http://fund.eastmoney.com/pingzhongdata/`
- 实时估值: 天天基金网 `http://fundgz.1234567.com.cn/js/`

## 快速开始

### 安装运行

```bash
# 进入项目目录
cd /Users/hisen/go/src/fund

# 运行服务
go run main.go
```

服务默认运行在 `http://localhost:8080`

### API接口

#### 1. 查询基金详情

**请求:**
```
GET /api/fund/detail?code={基金代码}
```

**参数:**
- `code`: 基金代码(6位数字),例如: 001186

**示例:**
```bash
curl "http://localhost:8080/api/fund/detail?code=001186"
```

**响应:**
```json
{
  "code": "001186",
  "name": "富国文体健康股票",
  "currentPrice": "1.2345",
  "estimatePrice": "1.2380",
  "estimateRate": "0.28",
  "updateTime": "2026-02-02 14:30",
  "dayGrowth": "0.28",
  "weekGrowth": "1.45",
  "monthGrowth": "3.21",
  "threeMonth": "8.56",
  "sixMonth": "12.34",
  "yearGrowth": "18.90",
  "totalGrowth": "23.45"
}
```

**字段说明:**
- `code`: 基金代码
- `name`: 基金名称
- `currentPrice`: 当前单位净值
- `estimatePrice`: 实时估算净值
- `estimateRate`: 实时估算增长率(%)
- `updateTime`: 数据更新时间
- `dayGrowth`: 日增长率(%)
- `weekGrowth`: 周增长率(%)
- `monthGrowth`: 月增长率(%)
- `threeMonth`: 近3月增长率(%)
- `sixMonth`: 近6月增长率(%)
- `yearGrowth`: 近1年增长率(%)
- `totalGrowth`: 成立以来增长率(%)

#### 2. 查询基金走势

**请求:**
```
GET /api/fund/trend?code={基金代码}&period={周期}
```

**参数:**
- `code`: 基金代码(6位数字),例如: 001186
- `period`: 时间周期(可选,默认`month`)
  - `week`: 最近一周
  - `month`: 最近一个月
  - `quarter`: 最近一个季度
  - `half_year`: 最近半年
  - `year`: 最近一年
  - `three_years`: 最近三年
  - `all`: 全部历史数据

**示例:**
```bash
# 获取最近一个月走势
curl "http://localhost:8080/api/fund/trend?code=001186&period=month"

# 获取最近一年走势
curl "http://localhost:8080/api/fund/trend?code=001186&period=year"
```

**响应:**
```json
{
  "code": "001186",
  "name": "富国文体健康股票A",
  "period": "month",
  "data": [
    {
      "date": "2026-01-03",
      "value": 1.2345
    },
    {
      "date": "2026-01-06",
      "value": 1.2380
    }
  ]
}
```

**字段说明:**
- `code`: 基金代码
- `name`: 基金名称
- `period`: 查询的时间周期
- `data`: 净值数据点数组
  - `date`: 日期(YYYY-MM-DD)
  - `value`: 单位净值

#### 3. 查询基金日内数据 🔥 新功能

**请求:**
```
GET /api/fund/intraday?code={基金代码}
```

**参数:**
- `code`: 基金代码(6位数字)

**说明:**
- 返回当天从开盘到当前时刻的所有估值数据点
- 每30秒采集一次
- 仅在交易时间(工作日 9:30-15:00)采集
- 每天晚上21:00自动清理数据

**示例:**
```bash
curl "http://localhost:8080/api/fund/intraday?code=001186"
```

**响应:**
```json
{
  "code": "001186",
  "name": "富国文体健康股票A",
  "date": "2026-02-02",
  "data": [
    {
      "time": "09:35",
      "value": 3.011,
      "rate": 0.15
    },
    {
      "time": "10:05",
      "value": 3.015,
      "rate": 0.28
    }
  ]
}
```

#### 4. 获取基金列表 🔥 新功能

**请求:**
```
GET /api/fund/list?page={页码}&pageSize={每页数量}
```

**参数:**
- `page`: 页码,默认1
- `pageSize`: 每页数量,默认100,最大1000

**示例:**
```bash
curl "http://localhost:8080/api/fund/list?page=1&pageSize=10"
```

**响应:**
```json
{
  "total": 26078,
  "page": 1,
  "pageSize": 10,
  "data": [
    {
      "code": "000001",
      "name": "华夏成长混合",
      "type": "混合型-灵活"
    }
  ]
}
```

#### 5. 查询服务状态 🔥 新功能

**请求:**
```
GET /api/status
```

**响应:**
```json
{
  "status": "running",
  "fundCount": 26078,
  "dataCount": 15234,
  "currentTime": "2026-02-02 14:30:15"
}
```

#### 6. 健康检查

**请求:**
```
GET /health
```

**响应:**
```json
{
  "status": "ok"
}
```

## 错误处理

API使用标准HTTP状态码:

- `200`: 请求成功
- `400`: 请求参数错误
- `500`: 服务器内部错误

错误响应格式:
```json
{
  "error": "错误描述信息"
}
```

## 常见基金代码示例

- `001186`: 富国文体健康股票
- `000001`: 华夏成长混合
- `110022`: 易方达消费行业股票
- `320007`: 诺安成长混合

## 技术栈

- Go 1.21+
- 标准库 `net/http`
- JSON数据处理
- 正则表达式解析

## 代码规范

- 使用依赖注入
- 单一职责原则
- 清晰的错误处理
- 统一的响应格式

## 扩展建议

1. **配置管理**: 添加 `config` 包管理端口、超时等配置
2. **日志系统**: 添加结构化日志记录
3. **数据缓存**: 使用Redis缓存基金数据
4. **单元测试**: 为各层编写测试用例
5. **文档生成**: 集成Swagger自动生成API文档

## License

MIT
