/**
 * Spatial Query Manager - TypeScript implementation with enhanced caching and validation
 * Provides efficient object queries using server-side spatial indexing
 */
import { logger } from './Logger';
import { createErrorHandler } from './ErrorHandler';
export class SpatialQueryManager {
    constructor(rpcCall, config = {}) {
        this.cache = new Map();
        this.errorHandler = createErrorHandler({ component: 'SpatialQueryManager' });
        this.logger = logger.createChildLogger('SpatialQueryManager');
        // Cache statistics
        this.cacheHits = 0;
        this.cacheMisses = 0;
        // Adaptive cache timeouts based on object type
        this.cacheTimeouts = new Map([
            ['static_objects', 30000], // 30 seconds - static objects change rarely
            ['dynamic_objects', 1000], // 1 second - dynamic objects change frequently
            ['players', 500], // 0.5 seconds - players move frequently
            ['items', 5000], // 5 seconds - items change moderately
            ['npcs', 2000], // 2 seconds - NPCs change regularly
            ['decorations', 60000], // 1 minute - decorations never change
        ]);
        this.rpc = rpcCall;
        this.config = {
            defaultCacheTimeout: 1000,
            maxCacheSize: 1000,
            cleanupInterval: 30000,
            adaptiveCaching: true,
            ...config
        };
        this.startCleanupTimer();
        this.logger.info('SpatialQueryManager initialized', { config: this.config });
    }
    /**
     * Get all objects within a rectangular area using server-side spatial indexing
     */
    async getObjectsInRange(rect, sessionId, objectType = 'dynamic_objects') {
        return this.errorHandler.wrapAsync(async () => {
            this.validateRectangle(rect);
            this.validateSessionId(sessionId);
            const cacheKey = this.generateRangeCacheKey(rect, objectType);
            // Check cache first
            const cached = this.getCachedResult(cacheKey);
            if (cached) {
                this.cacheHits++;
                this.logger.debug('Cache hit for range query', { rect, objectType });
                return cached;
            }
            this.cacheMisses++;
            this.logger.debug('Cache miss for range query', { rect, objectType });
            // Query server
            const params = {
                session_id: sessionId,
                query_type: 'range',
                params: {
                    minX: rect.x,
                    minY: rect.y,
                    maxX: rect.x + rect.width,
                    maxY: rect.y + rect.height,
                    object_type: objectType
                }
            };
            const result = await this.rpc('spatialQuery', params);
            const objects = this.validateAndTransformResult(result);
            // Cache the result
            this.setCachedResult(cacheKey, objects, objectType);
            return objects;
        }, 'getObjectsInRange', {
            userMessage: 'Failed to query objects in range',
            metadata: { rect, sessionId, objectType }
        })();
    }
    /**
     * Get all objects within a circular area using server-side spatial indexing
     */
    async getObjectsInRadius(center, radius, sessionId, objectType = 'dynamic_objects') {
        return this.errorHandler.wrapAsync(async () => {
            this.validatePosition(center);
            this.validateRadius(radius);
            this.validateSessionId(sessionId);
            const cacheKey = this.generateRadiusCacheKey(center, radius, objectType);
            // Check cache first
            const cached = this.getCachedResult(cacheKey);
            if (cached) {
                this.cacheHits++;
                this.logger.debug('Cache hit for radius query', { center, radius, objectType });
                return cached;
            }
            this.cacheMisses++;
            this.logger.debug('Cache miss for radius query', { center, radius, objectType });
            // Query server
            const params = {
                session_id: sessionId,
                query_type: 'radius',
                params: {
                    x: center.x,
                    y: center.y,
                    radius,
                    object_type: objectType
                }
            };
            const result = await this.rpc('spatialQuery', params);
            const objects = this.validateAndTransformResult(result);
            // Cache the result
            this.setCachedResult(cacheKey, objects, objectType);
            return objects;
        }, 'getObjectsInRadius', {
            userMessage: 'Failed to query objects in radius',
            metadata: { center, radius, sessionId, objectType }
        })();
    }
    /**
     * Get the k nearest objects to a position using server-side spatial indexing
     */
    async getNearestObjects(center, k, sessionId, objectType = 'dynamic_objects') {
        return this.errorHandler.wrapAsync(async () => {
            this.validatePosition(center);
            this.validateNearestCount(k);
            this.validateSessionId(sessionId);
            const cacheKey = this.generateNearestCacheKey(center, k, objectType);
            // Check cache first
            const cached = this.getCachedResult(cacheKey);
            if (cached) {
                this.cacheHits++;
                this.logger.debug('Cache hit for nearest query', { center, k, objectType });
                return cached;
            }
            this.cacheMisses++;
            this.logger.debug('Cache miss for nearest query', { center, k, objectType });
            // Query server
            const params = {
                session_id: sessionId,
                query_type: 'nearest',
                params: {
                    x: center.x,
                    y: center.y,
                    k,
                    object_type: objectType
                }
            };
            const result = await this.rpc('spatialQuery', params);
            const objects = this.validateAndTransformResult(result);
            // Cache the result
            this.setCachedResult(cacheKey, objects, objectType);
            return objects;
        }, 'getNearestObjects', {
            userMessage: 'Failed to query nearest objects',
            metadata: { center, k, sessionId, objectType }
        })();
    }
    /**
     * Clear the query cache
     */
    clearCache(objectType) {
        if (objectType) {
            // Clear only caches for specific object type
            const keysToDelete = [];
            for (const key of Array.from(this.cache.keys())) {
                if (key.includes(`type:${objectType}`)) {
                    keysToDelete.push(key);
                }
            }
            keysToDelete.forEach(key => this.cache.delete(key));
            this.logger.debug(`Cleared cache for object type: ${objectType}`, {
                clearedEntries: keysToDelete.length
            });
        }
        else {
            // Clear all caches
            const size = this.cache.size;
            this.cache.clear();
            this.logger.debug('Cleared all cache entries', { clearedEntries: size });
        }
    }
    /**
     * Get cache statistics for monitoring and optimization
     */
    getCacheStats() {
        const totalRequests = this.cacheHits + this.cacheMisses;
        const hitRate = totalRequests > 0 ? this.cacheHits / totalRequests : 0;
        // Calculate average TTL
        let totalTTL = 0;
        for (const entry of Array.from(this.cache.values())) {
            totalTTL += entry.ttl;
        }
        const averageTTL = this.cache.size > 0 ? totalTTL / this.cache.size : 0;
        return {
            hits: this.cacheHits,
            misses: this.cacheMisses,
            size: this.cache.size,
            hitRate,
            averageTTL
        };
    }
    /**
     * Configure cache timeouts for specific object types
     */
    configureCacheTimeouts(timeoutConfig) {
        for (const [objectType, timeout] of Object.entries(timeoutConfig)) {
            this.cacheTimeouts.set(objectType, timeout);
        }
        this.logger.info('Updated cache timeouts', { timeoutConfig });
    }
    /**
     * Cleanup and destroy the spatial query manager
     */
    destroy() {
        if (this.cleanupTimer) {
            clearInterval(this.cleanupTimer);
            this.cleanupTimer = undefined;
        }
        this.clearCache();
        this.logger.info('SpatialQueryManager destroyed');
    }
    // Private helper methods
    getCachedResult(key) {
        const entry = this.cache.get(key);
        if (!entry) {
            return null;
        }
        // Check if expired
        if (Date.now() > entry.timestamp + entry.ttl) {
            this.cache.delete(key);
            return null;
        }
        return entry.data;
    }
    setCachedResult(key, data, objectType) {
        // Enforce cache size limit
        if (this.cache.size >= this.config.maxCacheSize) {
            this.evictOldestEntries();
        }
        const ttl = this.getCacheTimeout(objectType);
        const entry = {
            data,
            timestamp: Date.now(),
            ttl
        };
        this.cache.set(key, entry);
    }
    getCacheTimeout(objectType) {
        if (this.config.adaptiveCaching && this.cacheTimeouts.has(objectType)) {
            return this.cacheTimeouts.get(objectType);
        }
        return this.config.defaultCacheTimeout;
    }
    evictOldestEntries() {
        // Remove 20% of oldest entries
        const entriesToRemove = Math.floor(this.cache.size * 0.2);
        const entries = Array.from(this.cache.entries()).sort((a, b) => a[1].timestamp - b[1].timestamp);
        for (let i = 0; i < entriesToRemove && i < entries.length; i++) {
            this.cache.delete(entries[i][0]);
        }
    }
    startCleanupTimer() {
        this.cleanupTimer = setInterval(() => {
            this.cleanupExpiredEntries();
        }, this.config.cleanupInterval);
    }
    cleanupExpiredEntries() {
        const now = Date.now();
        const keysToDelete = [];
        for (const [key, entry] of Array.from(this.cache.entries())) {
            if (now > entry.timestamp + entry.ttl) {
                keysToDelete.push(key);
            }
        }
        keysToDelete.forEach(key => this.cache.delete(key));
        if (keysToDelete.length > 0) {
            this.logger.debug('Cleaned up expired cache entries', {
                cleanedEntries: keysToDelete.length,
                remainingEntries: this.cache.size
            });
        }
    }
    // Cache key generation methods
    generateRangeCacheKey(rect, objectType) {
        return `range:${rect.x},${rect.y},${rect.width},${rect.height}:type:${objectType}`;
    }
    generateRadiusCacheKey(center, radius, objectType) {
        return `radius:${center.x},${center.y},${radius}:type:${objectType}`;
    }
    generateNearestCacheKey(center, k, objectType) {
        return `nearest:${center.x},${center.y},${k}:type:${objectType}`;
    }
    // Validation methods
    validateRectangle(rect) {
        if (!rect || typeof rect.x !== 'number' || typeof rect.y !== 'number' ||
            typeof rect.width !== 'number' || typeof rect.height !== 'number') {
            throw new Error('Invalid rectangle parameters');
        }
        if (rect.width <= 0 || rect.height <= 0) {
            throw new Error('Rectangle dimensions must be positive');
        }
    }
    validatePosition(position) {
        if (!position || typeof position.x !== 'number' || typeof position.y !== 'number') {
            throw new Error('Invalid position parameters');
        }
    }
    validateRadius(radius) {
        if (typeof radius !== 'number' || radius <= 0) {
            throw new Error('Radius must be a positive number');
        }
    }
    validateNearestCount(k) {
        if (typeof k !== 'number' || k <= 0 || !Number.isInteger(k)) {
            throw new Error('k must be a positive integer');
        }
    }
    validateSessionId(sessionId) {
        if (typeof sessionId !== 'string' || sessionId.trim().length === 0) {
            throw new Error('Invalid session ID');
        }
    }
    validateAndTransformResult(result) {
        if (!result || !Array.isArray(result.objects)) {
            throw new Error('Invalid spatial query result format');
        }
        // Transform and validate each object
        return result.objects.map((obj, index) => {
            if (!obj || typeof obj !== 'object') {
                throw new Error(`Invalid object at index ${index}`);
            }
            // Basic validation and transformation
            const gameObject = obj;
            if (!gameObject.id || !gameObject.position) {
                throw new Error(`Invalid game object structure at index ${index}`);
            }
            return gameObject;
        });
    }
}
//# sourceMappingURL=SpatialQueryManager.js.map