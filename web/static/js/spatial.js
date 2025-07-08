/**
 * Spatial Query Manager - Provides efficient object queries using server-side spatial indexing
 * Replaces inefficient client-side object iteration with optimized server queries
 */
class SpatialQueryManager {
    constructor(rpcClient) {
        this.rpc = rpcClient;
        this.cache = new Map();
        this.cacheTimeout = 1000; // 1 second cache timeout
    }

    /**
     * Get all objects within a rectangular area using server-side spatial indexing
     * @param {Object} rect - Rectangle parameters
     * @param {number} rect.minX - Minimum X coordinate
     * @param {number} rect.minY - Minimum Y coordinate  
     * @param {number} rect.maxX - Maximum X coordinate
     * @param {number} rect.maxY - Maximum Y coordinate
     * @param {string} sessionId - Player session ID
     * @returns {Promise<Array>} Array of objects within the range
     */
    async getObjectsInRange(rect, sessionId) {
        const cacheKey = `range_${rect.minX}_${rect.minY}_${rect.maxX}_${rect.maxY}`;
        
        // Check cache first
        const cached = this.cache.get(cacheKey);
        if (cached && (Date.now() - cached.timestamp) < this.cacheTimeout) {
            return cached.objects;
        }

        try {
            const result = await this.rpc.call('getObjectsInRange', {
                session_id: sessionId,
                min_x: rect.minX,
                min_y: rect.minY,
                max_x: rect.maxX,
                max_y: rect.maxY
            });

            if (result.success) {
                // Cache the result
                this.cache.set(cacheKey, {
                    objects: result.objects,
                    timestamp: Date.now()
                });
                
                return result.objects;
            } else {
                console.error('Range query failed:', result.error);
                return [];
            }
        } catch (error) {
            console.error('Range query error:', error);
            return [];
        }
    }

    /**
     * Get all objects within a circular area using server-side spatial indexing
     * @param {Object} center - Center position
     * @param {number} center.x - X coordinate of center
     * @param {number} center.y - Y coordinate of center
     * @param {number} radius - Search radius
     * @param {string} sessionId - Player session ID
     * @returns {Promise<Array>} Array of objects within the radius
     */
    async getObjectsInRadius(center, radius, sessionId) {
        const cacheKey = `radius_${center.x}_${center.y}_${radius}`;
        
        // Check cache first
        const cached = this.cache.get(cacheKey);
        if (cached && (Date.now() - cached.timestamp) < this.cacheTimeout) {
            return cached.objects;
        }

        try {
            const result = await this.rpc.call('getObjectsInRadius', {
                session_id: sessionId,
                center_x: center.x,
                center_y: center.y,
                radius: radius
            });

            if (result.success) {
                // Cache the result
                this.cache.set(cacheKey, {
                    objects: result.objects,
                    timestamp: Date.now()
                });
                
                return result.objects;
            } else {
                console.error('Radius query failed:', result.error);
                return [];
            }
        } catch (error) {
            console.error('Radius query error:', error);
            return [];
        }
    }

    /**
     * Get the k nearest objects to a position using server-side spatial indexing
     * @param {Object} center - Center position
     * @param {number} center.x - X coordinate of center
     * @param {number} center.y - Y coordinate of center
     * @param {number} k - Number of nearest objects to return
     * @param {string} sessionId - Player session ID
     * @param {string} [objectType='dynamic_objects'] - Type of objects being queried
     * @returns {Promise<Array>} Array of k nearest objects
     */
    async getNearestObjects(center, k, sessionId, objectType = 'dynamic_objects') {
        const nearestType = objectType.includes('static') ? 'nearest_static' : 'nearest_dynamic';
        const cacheInfo = this.generateCacheKey(
            `nearest_${center.x}_${center.y}_${k}`, 
            nearestType
        );
        const timeout = this.getCacheTimeout(nearestType, { k });
        
        // Check cache first with adaptive timeout
        const cached = this.cache.get(cacheInfo.fullKey);
        if (cached && (Date.now() - cached.timestamp) < timeout) {
            return cached.objects;
        }

        try {
            const result = await this.rpc.call('getNearestObjects', {
                session_id: sessionId,
                center_x: center.x,
                center_y: center.y,
                k: k
            });

            if (result.success) {
                // Cache the result with metadata
                this.cache.set(cacheInfo.fullKey, {
                    objects: result.objects,
                    timestamp: Date.now(),
                    queryType: nearestType,
                    timeout: timeout
                });
                
                return result.objects;
            } else {
                console.error('Nearest objects query failed:', result.error);
                return [];
            }
        } catch (error) {
            console.error('Nearest objects query error:', error);
            return [];
        }
    }

    /**
     * Clear the query cache (call when game state changes significantly)
     * @param {string} [objectType] - Optional: clear only specific object type caches
     */
    clearCache(objectType = null) {
        if (objectType) {
            // Clear only caches for specific object type
            for (const [key] of this.cache.entries()) {
                if (key.startsWith(`${objectType}:`)) {
                    this.cache.delete(key);
                }
            }
        } else {
            this.cache.clear();
        }
    }

    /**
     * Remove expired entries from cache using adaptive timeouts
     */
    cleanupCache() {
        const now = Date.now();
        for (const [key, value] of this.cache.entries()) {
            const timeout = value.timeout || this.defaultCacheTimeout;
            if (now - value.timestamp > timeout) {
                this.cache.delete(key);
            }
        }
    }

    /**
     * Get cache statistics for monitoring and optimization
     * @returns {Object} Cache statistics including hit rates, sizes, timeouts
     */
    getCacheStats() {
        const stats = {
            totalEntries: this.cache.size,
            typeBreakdown: {},
            averageAge: 0,
            expiredEntries: 0
        };

        const now = Date.now();
        let totalAge = 0;

        for (const [key, value] of this.cache.entries()) {
            const type = value.queryType || 'unknown';
            const age = now - value.timestamp;
            const timeout = value.timeout || this.defaultCacheTimeout;

            if (!stats.typeBreakdown[type]) {
                stats.typeBreakdown[type] = { count: 0, averageAge: 0 };
            }
            
            stats.typeBreakdown[type].count++;
            stats.typeBreakdown[type].averageAge += age;
            totalAge += age;

            if (age > timeout) {
                stats.expiredEntries++;
            }
        }

        // Calculate averages
        if (stats.totalEntries > 0) {
            stats.averageAge = totalAge / stats.totalEntries;
            
            for (const type in stats.typeBreakdown) {
                stats.typeBreakdown[type].averageAge /= stats.typeBreakdown[type].count;
            }
        }

        return stats;
    }

    /**
     * Configure cache timeouts for specific object types
     * @param {Object} timeoutConfig - Object mapping object types to timeout values
     */
    configureCacheTimeouts(timeoutConfig) {
        Object.assign(this.cacheTimeouts, timeoutConfig);
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SpatialQueryManager;
} else if (typeof window !== 'undefined') {
    window.SpatialQueryManager = SpatialQueryManager;
}
