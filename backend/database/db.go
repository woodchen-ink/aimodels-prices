package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"aimodels-prices/config"
	"aimodels-prices/models"
)

// DB 是数据库连接的全局实例
var DB *gorm.DB

// Cache 接口定义了缓存的基本操作
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiration time.Duration)
	Delete(key string)
	Clear()
}

// MemoryCache 是一个简单的内存缓存实现
type MemoryCache struct {
	items map[string]cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

// 全局缓存实例
var GlobalCache Cache

// NewMemoryCache 创建一个新的内存缓存
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]cacheItem),
	}

	// 启动一个后台协程定期清理过期项
	go cache.janitor()

	return cache
}

// Get 从缓存中获取值
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// 检查是否过期
	if item.expiration > 0 && item.expiration < time.Now().UnixNano() {
		return nil, false
	}

	return item.value, true
}

// Set 设置缓存值
func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) {
	var exp int64

	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:      value,
		expiration: exp,
	}
}

// Delete 删除缓存项
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
}

// janitor 定期清理过期的缓存项
func (c *MemoryCache) janitor() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.deleteExpired()
	}
}

// deleteExpired 删除所有过期的项
func (c *MemoryCache) deleteExpired() {
	now := time.Now().UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if v.expiration > 0 && v.expiration < now {
			delete(c.items, k)
		}
	}
}

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	var err error

	// 构建MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// 连接MySQL
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	// 获取底层的SQL DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(20)           // 增加最大连接数
	sqlDB.SetMaxIdleConns(10)           // 增加空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 设置连接最大生命周期

	// 初始化缓存
	GlobalCache = NewMemoryCache()

	// 启动定期缓存任务
	go startCacheJobs()

	// 自动迁移表结构
	if err = migrateModels(); err != nil {
		return fmt.Errorf("failed to migrate models: %v", err)
	}

	return nil
}

// startCacheJobs 启动定期缓存任务
func startCacheJobs() {
	// 每5分钟执行一次
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 立即执行一次
	cacheCommonData()

	for {
		<-ticker.C
		cacheCommonData()
	}
}

// cacheCommonData 缓存常用数据
func cacheCommonData() {
	log.Println("开始自动缓存常用数据...")

	// 缓存所有模型类型
	cacheModelTypes()

	// 缓存所有提供商
	cacheProviders()

	// 缓存价格倍率
	cachePriceRates()

	log.Println("自动缓存常用数据完成")
}

// cacheModelTypes 缓存所有模型类型
func cacheModelTypes() {
	var types []models.ModelType
	if err := DB.Order("sort_order ASC, type_key ASC").Find(&types).Error; err != nil {
		log.Printf("缓存模型类型失败: %v", err)
		return
	}

	GlobalCache.Set("model_types", types, 30*time.Minute)
	log.Printf("已缓存 %d 个模型类型", len(types))
}

// cacheProviders 缓存所有提供商
func cacheProviders() {
	var providers []models.Provider
	if err := DB.Order("id").Find(&providers).Error; err != nil {
		log.Printf("缓存提供商失败: %v", err)
		return
	}

	GlobalCache.Set("providers", providers, 30*time.Minute)
	log.Printf("已缓存 %d 个提供商", len(providers))
}

// cachePriceRates 缓存价格倍率
func cachePriceRates() {
	// 获取所有已批准的价格
	var prices []models.Price
	if err := DB.Where("status = 'approved'").Find(&prices).Error; err != nil {
		log.Printf("缓存价格倍率失败: %v", err)
		return
	}

	// 按模型分组
	modelMap := make(map[string]map[uint]models.Price)
	for _, price := range prices {
		if _, exists := modelMap[price.Model]; !exists {
			modelMap[price.Model] = make(map[uint]models.Price)
		}
		modelMap[price.Model][price.ChannelType] = price
	}

	// 缓存常用的价格查询
	cachePriceQueries()
}

// cachePriceQueries 缓存常用的价格查询
func cachePriceQueries() {
	// 缓存第一页数据（无筛选条件）
	cachePricePage(1, 20, "", "")

	// 获取所有模型类型
	var modelTypes []models.ModelType
	if err := DB.Find(&modelTypes).Error; err != nil {
		log.Printf("获取模型类型失败: %v", err)
		return
	}

	// 获取所有提供商
	var providers []models.Provider
	if err := DB.Find(&providers).Error; err != nil {
		log.Printf("获取提供商失败: %v", err)
		return
	}

	// 为每种模型类型缓存第一页数据
	for _, mt := range modelTypes {
		cachePricePage(1, 20, "", mt.TypeKey)
	}

	// 为每个提供商缓存第一页数据
	for _, p := range providers {
		channelType := fmt.Sprintf("%d", p.ID)
		cachePricePage(1, 20, channelType, "")
	}
}

// cachePricePage 缓存特定页的价格数据
func cachePricePage(page, pageSize int, channelType, modelType string) {
	offset := (page - 1) * pageSize

	// 构建查询
	query := DB.Model(&models.Price{})

	// 添加筛选条件
	if channelType != "" {
		query = query.Where("channel_type = ?", channelType)
	}
	if modelType != "" {
		query = query.Where("model_type = ?", modelType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("计算价格总数失败: %v", err)
		return
	}

	// 获取分页数据
	var prices []models.Price
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&prices).Error; err != nil {
		log.Printf("获取价格数据失败: %v", err)
		return
	}

	result := map[string]interface{}{
		"total": total,
		"data":  prices,
	}

	// 构建缓存键
	cacheKey := fmt.Sprintf("prices_page_%d_size_%d_channel_%s_type_%s",
		page, pageSize, channelType, modelType)

	// 存入缓存，有效期5分钟
	GlobalCache.Set(cacheKey, result, 5*time.Minute)
	log.Printf("已缓存价格查询: %s", cacheKey)
}

// migrateModels 自动迁移模型到数据库表
func migrateModels() error {
	// 自动迁移模型
	if err := DB.AutoMigrate(
		&models.ModelType{},
		&models.Price{},
		&models.Provider{},
		&models.User{},
		&models.Session{},
	); err != nil {
		log.Printf("Failed to migrate tables: %v", err)
		return err
	}

	return nil
}
